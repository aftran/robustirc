// Code generated by protoc-gen-go.
// source: types.proto
// DO NOT EDIT!

/*
Package proto is a generated protocol buffer package.

It is generated from these files:
	types.proto
	snapshot.proto

It has these top-level messages:
	RobustId
	RobustMessage
	SnapshotLog
	Timestamp
	Snapshot
*/
package proto

import proto1 "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto1.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
const _ = proto1.ProtoPackageIsVersion1

type RobustMessage_RobustType int32

const (
	RobustMessage_CREATE_SESSION   RobustMessage_RobustType = 0
	RobustMessage_DELETE_SESSION   RobustMessage_RobustType = 1
	RobustMessage_IRC_FROM_CLIENT  RobustMessage_RobustType = 2
	RobustMessage_IRC_TO_CLIENT    RobustMessage_RobustType = 3
	RobustMessage_PING             RobustMessage_RobustType = 4
	RobustMessage_MESSAGE_OF_DEATH RobustMessage_RobustType = 5
	RobustMessage_CONFIG           RobustMessage_RobustType = 6
	RobustMessage_ANY              RobustMessage_RobustType = 7
)

var RobustMessage_RobustType_name = map[int32]string{
	0: "CREATE_SESSION",
	1: "DELETE_SESSION",
	2: "IRC_FROM_CLIENT",
	3: "IRC_TO_CLIENT",
	4: "PING",
	5: "MESSAGE_OF_DEATH",
	6: "CONFIG",
	7: "ANY",
}
var RobustMessage_RobustType_value = map[string]int32{
	"CREATE_SESSION":   0,
	"DELETE_SESSION":   1,
	"IRC_FROM_CLIENT":  2,
	"IRC_TO_CLIENT":    3,
	"PING":             4,
	"MESSAGE_OF_DEATH": 5,
	"CONFIG":           6,
	"ANY":              7,
}

func (x RobustMessage_RobustType) String() string {
	return proto1.EnumName(RobustMessage_RobustType_name, int32(x))
}
func (RobustMessage_RobustType) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{1, 0} }

type RobustId struct {
	Id    int64 `protobuf:"fixed64,1,opt,name=id" json:"id,omitempty"`
	Reply int64 `protobuf:"fixed64,2,opt,name=reply" json:"reply,omitempty"`
}

func (m *RobustId) Reset()                    { *m = RobustId{} }
func (m *RobustId) String() string            { return proto1.CompactTextString(m) }
func (*RobustId) ProtoMessage()               {}
func (*RobustId) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type RobustMessage struct {
	Id      *RobustId                `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	Session *RobustId                `protobuf:"bytes,2,opt,name=session" json:"session,omitempty"`
	Type    RobustMessage_RobustType `protobuf:"varint,3,opt,name=type,enum=proto.RobustMessage_RobustType" json:"type,omitempty"`
	Data    string                   `protobuf:"bytes,4,opt,name=data" json:"data,omitempty"`
	// TODO: use oneof for the following to save space?
	Servers         []string `protobuf:"bytes,5,rep,name=servers" json:"servers,omitempty"`
	CurrentMaster   string   `protobuf:"bytes,6,opt,name=current_master,json=currentMaster" json:"current_master,omitempty"`
	ClientMessageId uint64   `protobuf:"varint,7,opt,name=client_message_id,json=clientMessageId" json:"client_message_id,omitempty"`
	Revision        uint64   `protobuf:"varint,8,opt,name=revision" json:"revision,omitempty"`
}

func (m *RobustMessage) Reset()                    { *m = RobustMessage{} }
func (m *RobustMessage) String() string            { return proto1.CompactTextString(m) }
func (*RobustMessage) ProtoMessage()               {}
func (*RobustMessage) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *RobustMessage) GetId() *RobustId {
	if m != nil {
		return m.Id
	}
	return nil
}

func (m *RobustMessage) GetSession() *RobustId {
	if m != nil {
		return m.Session
	}
	return nil
}

type SnapshotLog struct {
	Index uint64         `protobuf:"fixed64,1,opt,name=index" json:"index,omitempty"`
	Term  uint64         `protobuf:"fixed64,2,opt,name=term" json:"term,omitempty"`
	Msg   *RobustMessage `protobuf:"bytes,3,opt,name=msg" json:"msg,omitempty"`
}

func (m *SnapshotLog) Reset()                    { *m = SnapshotLog{} }
func (m *SnapshotLog) String() string            { return proto1.CompactTextString(m) }
func (*SnapshotLog) ProtoMessage()               {}
func (*SnapshotLog) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *SnapshotLog) GetMsg() *RobustMessage {
	if m != nil {
		return m.Msg
	}
	return nil
}

func init() {
	proto1.RegisterType((*RobustId)(nil), "proto.RobustId")
	proto1.RegisterType((*RobustMessage)(nil), "proto.RobustMessage")
	proto1.RegisterType((*SnapshotLog)(nil), "proto.SnapshotLog")
	proto1.RegisterEnum("proto.RobustMessage_RobustType", RobustMessage_RobustType_name, RobustMessage_RobustType_value)
}

var fileDescriptor0 = []byte{
	// 407 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x6c, 0x51, 0xc1, 0x6e, 0x9b, 0x40,
	0x10, 0xad, 0x0d, 0xc6, 0xee, 0x58, 0xb6, 0x37, 0xd3, 0x1c, 0x50, 0x2f, 0x89, 0x2c, 0xb5, 0x6a,
	0x7b, 0x88, 0x2a, 0xe7, 0x0b, 0x2c, 0x82, 0x5d, 0x24, 0x03, 0xd5, 0xc2, 0xa5, 0xa7, 0x15, 0x09,
	0x2b, 0x17, 0x29, 0x06, 0xb4, 0x4b, 0xa2, 0xe6, 0x33, 0xfa, 0x49, 0xfd, 0xb3, 0x2e, 0x03, 0xa4,
	0x8d, 0x94, 0xd3, 0xce, 0xbc, 0x79, 0x6f, 0x66, 0x76, 0x1e, 0xcc, 0x9b, 0xa7, 0x5a, 0xea, 0xab,
	0x5a, 0x55, 0x4d, 0x85, 0x13, 0x7a, 0xd6, 0x5f, 0x61, 0xc6, 0xab, 0xdb, 0x07, 0xdd, 0x04, 0x39,
	0x2e, 0x61, 0x5c, 0xe4, 0xee, 0xe8, 0x72, 0xf4, 0x89, 0x71, 0x13, 0xe1, 0x39, 0x4c, 0x94, 0xac,
	0xef, 0x9f, 0xdc, 0x31, 0x41, 0x5d, 0xb2, 0xfe, 0x63, 0xc1, 0xa2, 0x93, 0x84, 0x52, 0xeb, 0xec,
	0x28, 0xf1, 0xe2, 0x59, 0x37, 0xdf, 0xac, 0xba, 0xf6, 0x57, 0x43, 0x53, 0x6a, 0xf4, 0x19, 0xa6,
	0xda, 0x70, 0x8b, 0xaa, 0xa4, 0x56, 0xaf, 0xb0, 0x86, 0x3a, 0x5e, 0x83, 0xdd, 0x6e, 0xe9, 0x5a,
	0x86, 0xb7, 0xdc, 0x5c, 0xbc, 0xe0, 0xf5, 0xf3, 0xfa, 0x2c, 0x35, 0x34, 0x4e, 0x64, 0x44, 0xb0,
	0xf3, 0xac, 0xc9, 0x5c, 0xdb, 0x88, 0xde, 0x72, 0x8a, 0xd1, 0x6d, 0x67, 0xaa, 0x47, 0xa9, 0xb4,
	0x3b, 0xb9, 0xb4, 0x0c, 0x3c, 0xa4, 0xf8, 0x01, 0x96, 0x77, 0x0f, 0x4a, 0xc9, 0xb2, 0x11, 0xa7,
	0x4c, 0x37, 0x52, 0xb9, 0x0e, 0xe9, 0x16, 0x3d, 0x1a, 0x12, 0x88, 0x5f, 0xe0, 0xec, 0xee, 0xbe,
	0x20, 0x56, 0x37, 0x57, 0x98, 0x4f, 0x4e, 0x0d, 0xd3, 0xe6, 0xab, 0xae, 0xd0, 0xef, 0x63, 0x2e,
	0xf7, 0x1e, 0x66, 0x4a, 0x3e, 0x16, 0xf4, 0xc3, 0x19, 0x51, 0x9e, 0xf3, 0xf5, 0xef, 0x11, 0xc0,
	0xbf, 0x8d, 0xcd, 0xae, 0x4b, 0x8f, 0xfb, 0xdb, 0xd4, 0x17, 0x89, 0x9f, 0x24, 0x41, 0x1c, 0xb1,
	0x37, 0x2d, 0x76, 0xe3, 0x1f, 0xfc, 0xff, 0xb0, 0x11, 0xbe, 0x83, 0x55, 0xc0, 0x3d, 0xb1, 0xe3,
	0x71, 0x28, 0xbc, 0x43, 0xe0, 0x47, 0x29, 0x1b, 0xe3, 0x19, 0x2c, 0x5a, 0x30, 0x8d, 0x07, 0xc8,
	0xc2, 0x19, 0xd8, 0xdf, 0x83, 0x68, 0xcf, 0x6c, 0x63, 0x17, 0x0b, 0x8d, 0x7c, 0xbb, 0xf7, 0x45,
	0xbc, 0x13, 0x37, 0x66, 0xc6, 0x37, 0x36, 0x41, 0x00, 0xc7, 0x8b, 0xa3, 0x5d, 0xb0, 0x67, 0x0e,
	0x4e, 0xc1, 0xda, 0x46, 0x3f, 0xd8, 0x74, 0x2d, 0x60, 0x9e, 0x94, 0x59, 0xad, 0x7f, 0x56, 0xcd,
	0xa1, 0x3a, 0xb6, 0x46, 0x17, 0x65, 0x2e, 0x7f, 0x91, 0x87, 0x0e, 0xef, 0x92, 0xf6, 0xaa, 0xe6,
	0x0e, 0x27, 0xb2, 0xcc, 0xe1, 0x14, 0xe3, 0x47, 0xb0, 0x4e, 0xfa, 0x48, 0xee, 0xcc, 0x37, 0xe7,
	0xaf, 0xb9, 0xc3, 0x5b, 0xc2, 0xad, 0x43, 0x95, 0xeb, 0xbf, 0x01, 0x00, 0x00, 0xff, 0xff, 0xe1,
	0xfe, 0xab, 0xd4, 0x73, 0x02, 0x00, 0x00,
}
