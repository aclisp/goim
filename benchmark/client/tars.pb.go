// Code generated by protoc-gen-go. DO NOT EDIT.
// source: tars.proto

/*
Package main is a generated protocol buffer package.

It is generated from these files:
	tars.proto
	attentionlist.proto

It has these top-level messages:
	RPCInput
	RPCOutput
	ServerPush
	MultiPush
	PQueryUserAttentionListReq
	PQueryUserAttentionListRsp
*/
package main

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type RPCInput struct {
	ServiceName   string            `protobuf:"bytes,1,opt,name=serviceName" json:"serviceName,omitempty"`
	MethodName    string            `protobuf:"bytes,2,opt,name=methodName" json:"methodName,omitempty"`
	RequestBuffer []byte            `protobuf:"bytes,3,opt,name=requestBuffer,proto3" json:"requestBuffer,omitempty"`
	Headers       map[string]string `protobuf:"bytes,4,rep,name=headers" json:"headers,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

func (m *RPCInput) Reset()                    { *m = RPCInput{} }
func (m *RPCInput) String() string            { return proto.CompactTextString(m) }
func (*RPCInput) ProtoMessage()               {}
func (*RPCInput) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *RPCInput) GetServiceName() string {
	if m != nil {
		return m.ServiceName
	}
	return ""
}

func (m *RPCInput) GetMethodName() string {
	if m != nil {
		return m.MethodName
	}
	return ""
}

func (m *RPCInput) GetRequestBuffer() []byte {
	if m != nil {
		return m.RequestBuffer
	}
	return nil
}

func (m *RPCInput) GetHeaders() map[string]string {
	if m != nil {
		return m.Headers
	}
	return nil
}

type RPCOutput struct {
	RetCode        int32             `protobuf:"zigzag32,1,opt,name=retCode" json:"retCode,omitempty"`
	ResponseBuffer []byte            `protobuf:"bytes,2,opt,name=responseBuffer,proto3" json:"responseBuffer,omitempty"`
	Headers        map[string]string `protobuf:"bytes,3,rep,name=headers" json:"headers,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	RetDesc        string            `protobuf:"bytes,4,opt,name=retDesc" json:"retDesc,omitempty"`
	ServiceName    string            `protobuf:"bytes,5,opt,name=serviceName" json:"serviceName,omitempty"`
	MethodName     string            `protobuf:"bytes,6,opt,name=methodName" json:"methodName,omitempty"`
}

func (m *RPCOutput) Reset()                    { *m = RPCOutput{} }
func (m *RPCOutput) String() string            { return proto.CompactTextString(m) }
func (*RPCOutput) ProtoMessage()               {}
func (*RPCOutput) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *RPCOutput) GetRetCode() int32 {
	if m != nil {
		return m.RetCode
	}
	return 0
}

func (m *RPCOutput) GetResponseBuffer() []byte {
	if m != nil {
		return m.ResponseBuffer
	}
	return nil
}

func (m *RPCOutput) GetHeaders() map[string]string {
	if m != nil {
		return m.Headers
	}
	return nil
}

func (m *RPCOutput) GetRetDesc() string {
	if m != nil {
		return m.RetDesc
	}
	return ""
}

func (m *RPCOutput) GetServiceName() string {
	if m != nil {
		return m.ServiceName
	}
	return ""
}

func (m *RPCOutput) GetMethodName() string {
	if m != nil {
		return m.MethodName
	}
	return ""
}

type ServerPush struct {
	MessageType int32             `protobuf:"zigzag32,1,opt,name=messageType" json:"messageType,omitempty"`
	PushBuffer  []byte            `protobuf:"bytes,2,opt,name=pushBuffer,proto3" json:"pushBuffer,omitempty"`
	Headers     map[string]string `protobuf:"bytes,3,rep,name=headers" json:"headers,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	MessageDesc string            `protobuf:"bytes,4,opt,name=messageDesc" json:"messageDesc,omitempty"`
	ServiceName string            `protobuf:"bytes,5,opt,name=serviceName" json:"serviceName,omitempty"`
	MethodName  string            `protobuf:"bytes,6,opt,name=methodName" json:"methodName,omitempty"`
}

func (m *ServerPush) Reset()                    { *m = ServerPush{} }
func (m *ServerPush) String() string            { return proto.CompactTextString(m) }
func (*ServerPush) ProtoMessage()               {}
func (*ServerPush) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *ServerPush) GetMessageType() int32 {
	if m != nil {
		return m.MessageType
	}
	return 0
}

func (m *ServerPush) GetPushBuffer() []byte {
	if m != nil {
		return m.PushBuffer
	}
	return nil
}

func (m *ServerPush) GetHeaders() map[string]string {
	if m != nil {
		return m.Headers
	}
	return nil
}

func (m *ServerPush) GetMessageDesc() string {
	if m != nil {
		return m.MessageDesc
	}
	return ""
}

func (m *ServerPush) GetServiceName() string {
	if m != nil {
		return m.ServiceName
	}
	return ""
}

func (m *ServerPush) GetMethodName() string {
	if m != nil {
		return m.MethodName
	}
	return ""
}

type MultiPush struct {
	Msg     *ServerPush `protobuf:"bytes,1,opt,name=msg" json:"msg,omitempty"`
	UserIDs []int64     `protobuf:"varint,2,rep,packed,name=userIDs" json:"userIDs,omitempty"`
	AppID   int32       `protobuf:"varint,3,opt,name=appID" json:"appID,omitempty"`
}

func (m *MultiPush) Reset()                    { *m = MultiPush{} }
func (m *MultiPush) String() string            { return proto.CompactTextString(m) }
func (*MultiPush) ProtoMessage()               {}
func (*MultiPush) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *MultiPush) GetMsg() *ServerPush {
	if m != nil {
		return m.Msg
	}
	return nil
}

func (m *MultiPush) GetUserIDs() []int64 {
	if m != nil {
		return m.UserIDs
	}
	return nil
}

func (m *MultiPush) GetAppID() int32 {
	if m != nil {
		return m.AppID
	}
	return 0
}

func init() {
	proto.RegisterType((*RPCInput)(nil), "main.RPCInput")
	proto.RegisterType((*RPCOutput)(nil), "main.RPCOutput")
	proto.RegisterType((*ServerPush)(nil), "main.ServerPush")
	proto.RegisterType((*MultiPush)(nil), "main.MultiPush")
}

func init() { proto.RegisterFile("tars.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 389 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xc4, 0x93, 0x4b, 0x4f, 0xea, 0x50,
	0x14, 0x85, 0xd3, 0x96, 0xc7, 0x65, 0xc3, 0xbd, 0xe1, 0x9e, 0xdc, 0x41, 0x73, 0x7d, 0xa4, 0x21,
	0xc6, 0x74, 0xd4, 0x01, 0xc6, 0x47, 0x18, 0x0a, 0x26, 0x32, 0x50, 0xc9, 0xd1, 0xb9, 0xa9, 0xb0,
	0xa1, 0x44, 0xfa, 0xf0, 0x3c, 0x48, 0xf8, 0x31, 0x8e, 0xfd, 0x65, 0xfe, 0x0f, 0x73, 0x4e, 0xa9,
	0x1c, 0x1a, 0x8d, 0x03, 0x07, 0xce, 0xd8, 0x6b, 0x6f, 0xb2, 0xd6, 0xfa, 0x9a, 0x03, 0x20, 0x42,
	0xc6, 0x83, 0x8c, 0xa5, 0x22, 0x25, 0x95, 0x38, 0x9c, 0x27, 0x9d, 0x57, 0x0b, 0x7e, 0xd1, 0x51,
	0x7f, 0x98, 0x64, 0x52, 0x10, 0x0f, 0x9a, 0x1c, 0xd9, 0x72, 0x3e, 0xc6, 0xeb, 0x30, 0x46, 0xd7,
	0xf2, 0x2c, 0xbf, 0x41, 0x4d, 0x89, 0xec, 0x03, 0xc4, 0x28, 0xa2, 0x74, 0xa2, 0x0f, 0x6c, 0x7d,
	0x60, 0x28, 0xe4, 0x00, 0x7e, 0x33, 0x7c, 0x92, 0xc8, 0xc5, 0xb9, 0x9c, 0x4e, 0x91, 0xb9, 0x8e,
	0x67, 0xf9, 0x2d, 0xba, 0x2d, 0x92, 0x63, 0xa8, 0x47, 0x18, 0x4e, 0x90, 0x71, 0xb7, 0xe2, 0x39,
	0x7e, 0xb3, 0xbb, 0x13, 0xa8, 0x30, 0x41, 0x11, 0x24, 0xb8, 0xcc, 0xb7, 0x17, 0x89, 0x60, 0x2b,
	0x5a, 0xdc, 0xfe, 0xef, 0x41, 0xcb, 0x5c, 0x90, 0x36, 0x38, 0x8f, 0xb8, 0x5a, 0xc7, 0x54, 0x3f,
	0xc9, 0x3f, 0xa8, 0x2e, 0xc3, 0x85, 0x2c, 0x92, 0xe5, 0x43, 0xcf, 0x3e, 0xb3, 0x3a, 0xcf, 0x36,
	0x34, 0xe8, 0xa8, 0x7f, 0x23, 0x85, 0x2a, 0xea, 0x42, 0x9d, 0xa1, 0xe8, 0xa7, 0x93, 0xbc, 0xe4,
	0x5f, 0x5a, 0x8c, 0xe4, 0x10, 0xfe, 0x30, 0xe4, 0x59, 0x9a, 0x70, 0x5c, 0x37, 0xb0, 0x75, 0x83,
	0x92, 0x4a, 0x4e, 0x36, 0x15, 0x1c, 0x5d, 0x61, 0xf7, 0xbd, 0x42, 0xee, 0xf1, 0x71, 0x87, 0xb5,
	0xf3, 0x00, 0xf9, 0xd8, 0xad, 0xe8, 0x8c, 0xc5, 0x58, 0x86, 0x5f, 0xfd, 0x0a, 0x7e, 0xad, 0x0c,
	0xff, 0x5b, 0x7c, 0x5e, 0x6c, 0x80, 0x5b, 0x64, 0x4b, 0x64, 0x23, 0xc9, 0x23, 0x15, 0x26, 0x46,
	0xce, 0xc3, 0x19, 0xde, 0xad, 0xb2, 0x02, 0x92, 0x29, 0xa9, 0x30, 0x99, 0xe4, 0xd1, 0x16, 0x24,
	0x43, 0x21, 0xa7, 0x65, 0x40, 0x7b, 0x39, 0xa0, 0x8d, 0xc9, 0x27, 0x84, 0x36, 0xd6, 0x06, 0x25,
	0x53, 0xfa, 0x61, 0x52, 0xf7, 0xd0, 0xb8, 0x92, 0x0b, 0x31, 0xd7, 0x9c, 0x3a, 0xe0, 0xc4, 0x7c,
	0xa6, 0xff, 0xd8, 0xec, 0xb6, 0xcb, 0x0d, 0xa9, 0x5a, 0xaa, 0x4f, 0x2e, 0x39, 0xb2, 0xe1, 0x80,
	0xbb, 0xb6, 0xe7, 0xf8, 0x0e, 0x2d, 0x46, 0x65, 0x12, 0x66, 0xd9, 0x70, 0xa0, 0x5f, 0x49, 0x95,
	0xe6, 0xc3, 0x43, 0x4d, 0xbf, 0xcf, 0xa3, 0xb7, 0x00, 0x00, 0x00, 0xff, 0xff, 0xdc, 0x60, 0x5f,
	0x2e, 0xad, 0x03, 0x00, 0x00,
}
