// proxy bridges between IRC clients (RFC1459) and fancyirc servers.
//
// Proxy instances are supposed to be long-running, and ideally as close to the
// IRC client as possible, e.g. on the same machine. When running on the same
// machine, there should not be any network problems between the IRC client and
// the proxy. Network problems between the proxy and a fancyirc network are
// handled transparently.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"fancyirc/types"

	"github.com/sorcix/irc"
)

var (
	network = flag.String("network",
		"",
		`DNS name to connect to (e.g. "robustirc.net"). The _robustirc._tcp SRV record must be present.`)

	serversList = flag.String("servers",
		"",
		"(comma-separated) list of host:port network addresses of the server(s) to connect to")

	listen = flag.String("listen",
		"localhost:6667",
		"host:port to listen on for IRC connections")

	socks = flag.String("socks", "", "host:port to listen on for SOCKS5 connections")
)

const (
	pathCreateSession = "/robustirc/v1/session"
	pathDeleteSession = "/robustirc/v1/%s"
	pathPostMessage   = "/robustirc/v1/%s/message"
	pathGetMessages   = "/robustirc/v1/%s/messages?lastseen=%s"
)

// TODO(secure): persistent state:
// - the last known server(s) in the network. added to *servers
// - for resuming sessions (later): the last seen message id, perhaps setup messages (JOINs, MODEs, …)
// for hosted mode, this state is stored per-nickname, ideally encrypted with password

type proxy struct {
	network       string
	servers       []string
	currentMaster string
	currentId     int
	mu            sync.RWMutex
}

func newProxy(network string, servers []string) (*proxy, error) {
	p := &proxy{
		network: network,
	}

	if network != "" {
		// Try to resolve the DNS name up to 5 times. This is to be nice to
		// people in environments with flaky network connections at boot, who,
		// for some reason, don’t run this program under systemd with
		// Restart=on-failure.
		try := 0
		for {
			_, addrs, err := net.LookupSRV("robustirc", "tcp", network)
			if err != nil {
				log.Println(err)
				if try < 4 {
					time.Sleep(time.Duration(int64(math.Pow(2, float64(try)))) * time.Second)
				} else {
					log.Fatalf("DNS lookup failed 5 times, exiting\n")
				}
				try++
				continue
			}
			for _, addr := range addrs {
				target := addr.Target
				if target[len(target)-1] == '.' {
					target = target[:len(target)-1]
				}
				p.servers = append(p.servers, fmt.Sprintf("%s:%d", target, addr.Port))
			}
			break
		}
	}

	p.servers = append(p.servers, servers...)

	if len(p.servers) == 0 {
		return nil, errors.New("connect without network and servers")
	}

	p.currentMaster = p.servers[0]

	return p, nil
}

// servers returns all configured servers, with the last-known master prepended.
func (p *proxy) getServers() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return append([]string{p.currentMaster}, p.servers...)
}

func (p *proxy) sendFancyMessage(logPrefix, sessionauth, method string, targets []string, path string, data []byte) (*http.Response, error) {
	var (
		resp   *http.Response
		target string
	)
	for {
		var soonest time.Duration
		target = ""
		for target == "" {
			target, soonest = nextCandidate(targets)
			if target == "" {
				log.Printf("%s Waiting %v for back-off time to expire…\n", logPrefix, soonest)
				time.Sleep(soonest)
			}
		}

		log.Printf("%s targets = %v, candidate = %s\n", logPrefix, targets, target)

		var err error
		req, err := http.NewRequest(method, fmt.Sprintf("https://%s%s", target, path), bytes.NewBuffer(data))
		if err != nil {
			return nil, err
		}
		req.Header.Set("X-Session-Auth", sessionauth)
		req.Header.Set("Content-Type", "application/json")
		resp, err = http.DefaultClient.Do(req)

		if err != nil {
			log.Printf("%s %v\n", logPrefix, err)
			serverFailed(target)
			continue
		}

		if resp.StatusCode == http.StatusTemporaryRedirect {
			loc := resp.Header.Get("Location")
			if loc == "" {
				return nil, fmt.Errorf("Redirect has no Location header")
			}
			u, err := url.Parse(loc)
			if err != nil {
				return nil, fmt.Errorf("Could not parse redirection %q: %v", loc, err)
			}

			resp.Body.Close()

			log.Printf("%s %q redirects us to %q\n", logPrefix, target, u.Host)

			// Even though the server did not actually fail, it did not answer our
			// request either. To prevent hammering it, mark it as failed for
			// back-off purposes.
			serverFailed(target)
			targets = append([]string{u.Host}, targets...)
			continue
		}

		if resp.StatusCode != 200 {
			data, _ := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			log.Printf("%s sendFancyMessage(%q) failed with %v: %q", logPrefix, path, resp.Status, string(data))
			serverFailed(target)
			continue
		}

		break
	}
	log.Printf("%s ->fancy: %q\n", logPrefix, string(data))

	p.mu.Lock()
	p.currentMaster = target
	p.mu.Unlock()
	return resp, nil
}

func (p *proxy) sendIRCMessage(logPrefix string, ircConn *irc.Conn, msg irc.Message) {
	if err := ircConn.Encode(&msg); err != nil {
		log.Printf("%s Error sending IRC message %q: %v. Closing connection.\n", logPrefix, msg.Bytes(), err)
		// This leads to an error in .Decode(), terminating the handleIRC goroutine.
		ircConn.Close()
		return
	}
	log.Printf("%s ->irc: %q\n", logPrefix, msg.Bytes())
}

func (p *proxy) createFancySession(logPrefix string) (session string, sessionauth string, prefix irc.Prefix, err error) {
	var resp *http.Response
	resp, err = p.sendFancyMessage(logPrefix, "", "POST", p.getServers(), pathCreateSession, []byte{})
	if err != nil {
		return
	}
	defer resp.Body.Close()
	// We need to read the entire body, otherwise net/http will not re-use this
	// connection.
	defer ioutil.ReadAll(resp.Body)

	var createSessionReply struct {
		Sessionid   string
		Sessionauth string
		Prefix      string
	}

	if err = json.NewDecoder(resp.Body).Decode(&createSessionReply); err != nil {
		return
	}

	session = createSessionReply.Sessionid
	sessionauth = createSessionReply.Sessionauth
	prefix = irc.Prefix{Name: createSessionReply.Prefix}
	return
}

func (p *proxy) deleteFancySession(logPrefix, sessionauth, session string, quitmsg string) error {
	deleteSessionRequest := struct {
		Quitmessage string
	}{
		quitmsg,
	}
	b, err := json.Marshal(deleteSessionRequest)
	if err != nil {
		return err
	}
	resp, err := p.sendFancyMessage(logPrefix, sessionauth, "DELETE", p.getServers(), fmt.Sprintf(pathDeleteSession, session), b)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// We need to read the entire body, otherwise net/http will not re-use this
	// connection.
	defer ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return fmt.Errorf("got %v, expected 200", resp.Status)
	}

	log.Printf("%s deleted session\n", logPrefix)

	return nil
}

func (p *proxy) handleIRC(conn net.Conn) {
	var (
		logPrefix       = conn.RemoteAddr().String()
		ircConn         = irc.NewConn(conn)
		ircErrors       = make(chan error)
		ircMessages     = make(chan irc.Message)
		fancyMessages   = make(chan string)
		stopGetMessages = make(chan bool)

		ircPrefix   irc.Prefix
		session     string
		sessionauth string
		quitmsg     string
		done        bool
		pingSent    bool
		err         error
	)

	type postMessageRequest struct {
		Data            string
		ClientMessageId int
	}

	session, sessionauth, ircPrefix, err = p.createFancySession(logPrefix)
	if err != nil {
		log.Printf("%s Could not create RobustIRC session: %v\n", logPrefix, err)
		p.sendIRCMessage(logPrefix, ircConn, irc.Message{
			Command:  "ERROR",
			Trailing: fmt.Sprintf("Could not create RobustIRC session: %v", err),
		})

		ircConn.Close()
		return
	}

	go func() {
		for {
			message, err := ircConn.Decode()
			if err != nil {
				ircErrors <- err
				return
			}
			log.Printf("%s <-irc: %q\n", logPrefix, message.Bytes())
			ircMessages <- *message
		}
	}()

	go func() {
		var lastSeen types.FancyId

		for !done {
			host, resp := p.getMessages(logPrefix, sessionauth, session, lastSeen)

			// We set the host as currentMaster, not because the host is the
			// master, but because it is reachable. When sending messages, we will
			// either reach the master by chance or get redirected, at which point
			// we update currentMaster.
			p.mu.Lock()
			p.currentMaster = host
			p.mu.Unlock()

			dec := json.NewDecoder(resp.Body)
			msgchan := make(chan types.FancyMessage)
			errchan := make(chan error)

			go func() {
				for {
					var msg types.FancyMessage
					if err := dec.Decode(&msg); err != nil {
						errchan <- err
						return
					}
					msgchan <- msg
				}
			}()

		Readloop:
			for !done {
				select {
				case err := <-errchan:
					log.Printf("%s Protocol error on %q: Could not decode response chunk as JSON: %v\n", logPrefix, host, err)
					serverFailed(host)
					break Readloop

				case <-time.After(1 * time.Minute):
					log.Printf("%s Timeout (60s) on GetMessages, reconnecting…\n", logPrefix)
					serverFailed(host)
					break Readloop

				case <-stopGetMessages:
					log.Printf("%s GetMessages aborted.\n", logPrefix)
					break Readloop

				case msg := <-msgchan:
					if msg.Type == types.FancyPing {
						p.mu.Lock()
						p.servers = msg.Servers
						p.currentMaster = msg.Currentmaster
						p.mu.Unlock()
						log.Printf("received ping (%+v). Servers are now %v\n", msg, p.getServers())
					} else if msg.Type == types.FancyIRCToClient {
						log.Printf("%s <-fancy: %q\n", logPrefix, msg.Data)
						fancyMessages <- msg.Data
						lastSeen = msg.Id
					}
				}
			}
			resp.Body.Close()
		}

		close(fancyMessages)
	}()

	// Cancel the GetMessages goroutine, read all remaining messages to prevent
	// goroutine hangs, then delete the session.
	defer func() {
		stopGetMessages <- true
		for _ = range fancyMessages {
		}

		if err := p.deleteFancySession(logPrefix, sessionauth, session, quitmsg); err != nil {
			log.Printf("%s Could not delete session: %v\n", logPrefix, err)
		}
	}()

	keepaliveTimer := time.After(1 * time.Minute)
	for {
		select {
		case <-time.After(1 * time.Minute):
			// After no traffic in either direction for 1 minute, we send a PING
			// message. If a PING message was already sent, this means that we did
			// not receive a PONG message, so we close the connection with a
			// timeout.
			if pingSent {
				quitmsg = "ping timeout"
				ircConn.Close()
			} else {
				p.sendIRCMessage(logPrefix, ircConn, irc.Message{
					Prefix:  &ircPrefix,
					Command: irc.PING,
					Params:  []string{"robustirc.proxy"},
				})
			}

		case <-keepaliveTimer:
			keepaliveTimer = time.After(1 * time.Minute)

			// Cannot fail, no user input.
			b, _ := json.Marshal(postMessageRequest{Data: "PING keepalive", ClientMessageId: p.currentId})
			resp, err := p.sendFancyMessage(logPrefix, sessionauth, "POST", p.getServers(), fmt.Sprintf(pathPostMessage, session), b)
			if err != nil {
				log.Printf("keepalive ping could not be sent: %v\n", err)
				break
			}
			p.currentId++

			// We need to read the entire body, otherwise net/http will not
			// re-use this connection.
			ioutil.ReadAll(resp.Body)
			resp.Body.Close()

		case err := <-ircErrors:
			log.Printf("Error in IRC client connection: %v\n", err)
			done = true
			return

		case msg := <-fancyMessages:
			ircmsg := irc.ParseMessage(msg)
			if ircmsg.Command == irc.PONG && len(ircmsg.Params) > 0 && ircmsg.Params[0] == "keepalive" {
				log.Printf("Swallowing keepalive PONG from server to avoid confusing the IRC client.\n")
				break
			}
			if _, err := fmt.Fprintf(conn, "%s\n", msg); err != nil {
				log.Printf("Error sending to IRC client: %v\n", err)
				done = true
				return
			}

		case message := <-ircMessages:
			switch message.Command {
			case irc.PONG:
				log.Printf("%s received PONG reply.\n", logPrefix)
				pingSent = false
			case irc.PING:
				p.sendIRCMessage(logPrefix, ircConn, irc.Message{
					Prefix:  &ircPrefix,
					Command: irc.PONG,
					Params:  message.Params,
				})
			case irc.QUIT:
				quitmsg = message.Trailing
				ircConn.Close()
			default:
				b, err := json.Marshal(postMessageRequest{Data: string(message.Bytes()), ClientMessageId: p.currentId})
				if err != nil {
					log.Printf("Message could not be encoded as JSON: %v\n", err)
					p.sendIRCMessage(logPrefix, ircConn, irc.Message{
						Prefix:   &ircPrefix,
						Command:  irc.ERROR,
						Trailing: fmt.Sprintf("Message could not be encoded as JSON: %v", err),
					})
					ircConn.Close()
					break
				}

				resp, err := p.sendFancyMessage(logPrefix, sessionauth, "POST", p.getServers(), fmt.Sprintf(pathPostMessage, session), b)
				if err != nil {
					// TODO(secure): what should we do here?
					log.Printf("message could not be sent: %v\n", err)
				}
				p.currentId++

				// We need to read the entire body, otherwise net/http will not
				// re-use this connection.
				ioutil.ReadAll(resp.Body)
				resp.Body.Close()
				keepaliveTimer = time.After(1 * time.Minute)
			}
		}
	}
}
func main() {
	flag.Parse()

	rand.Seed(time.Now().Unix())

	if *socks != "" {
		go func() {
			if err := listenAndServeSocks(*socks); err != nil {
				log.Fatal(err)
			}
		}()
	}

	if *network == "" && *serversList == "" {
		log.Fatal("You must specify either -network or -servers.")
	}

	var servers []string
	if *serversList != "" {
		// Start with any server. Will be overwritten later.
		servers = strings.Split(*serversList, ",")
		if len(servers) == 0 {
			log.Fatalf("Invalid -servers value (%q). Need at least one server.\n", *serversList)
		}
	}

	p, err := newProxy(*network, servers)
	if err != nil {
		log.Fatalf("Could not create proxy: %v\n", err)
	}

	ln, err := net.Listen("tcp", *listen)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("RobustIRC IRC bridge listening on %q\n", *listen)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Could not accept IRC client connection: %v\n", err)
			continue
		}
		go p.handleIRC(conn)
	}
}
