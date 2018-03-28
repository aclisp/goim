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
	RetCode        int32  `protobuf:"zigzag32,1,opt,name=retCode" json:"retCode,omitempty"`
	ResponseBuffer []byte `protobuf:"bytes,2,opt,name=responseBuffer,proto3" json:"responseBuffer,omitempty"`
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

func init() {
	proto.RegisterType((*RPCInput)(nil), "main.RPCInput")
	proto.RegisterType((*RPCOutput)(nil), "main.RPCOutput")
}

func init() { proto.RegisterFile("tars.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 242 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x5c, 0x90, 0xcf, 0x4a, 0xc3, 0x40,
	0x10, 0xc6, 0xd9, 0xa4, 0x5a, 0x3b, 0xad, 0xa2, 0x83, 0x87, 0xa0, 0x20, 0xa1, 0x88, 0xe4, 0x94,
	0x83, 0x22, 0x48, 0x8f, 0x06, 0x41, 0x0f, 0x6a, 0xd9, 0x37, 0x58, 0xcd, 0x94, 0x16, 0x4d, 0x36,
	0xce, 0xce, 0x16, 0xfa, 0xc0, 0xbe, 0x87, 0xe4, 0x1f, 0xc4, 0xde, 0x76, 0xbe, 0xef, 0x3b, 0xfc,
	0x7e, 0x0b, 0x20, 0x86, 0x5d, 0x5a, 0xb1, 0x15, 0x8b, 0xa3, 0xc2, 0x6c, 0xca, 0xf9, 0xaf, 0x82,
	0x23, 0xbd, 0xcc, 0x5e, 0xca, 0xca, 0x0b, 0xc6, 0x30, 0x75, 0xc4, 0xdb, 0xcd, 0x27, 0xbd, 0x99,
	0x82, 0x22, 0x15, 0xab, 0x64, 0xa2, 0x87, 0x11, 0x5e, 0x01, 0x14, 0x24, 0x6b, 0x9b, 0x37, 0x83,
	0xa0, 0x19, 0x0c, 0x12, 0xbc, 0x86, 0x63, 0xa6, 0x1f, 0x4f, 0x4e, 0x1e, 0xfd, 0x6a, 0x45, 0x1c,
	0x85, 0xb1, 0x4a, 0x66, 0xfa, 0x7f, 0x88, 0xf7, 0x30, 0x5e, 0x93, 0xc9, 0x89, 0x5d, 0x34, 0x8a,
	0xc3, 0x64, 0x7a, 0x7b, 0x99, 0xd6, 0x30, 0x69, 0x0f, 0x92, 0x3e, 0xb7, 0xed, 0x53, 0x29, 0xbc,
	0xd3, 0xfd, 0xf6, 0x62, 0x01, 0xb3, 0x61, 0x81, 0xa7, 0x10, 0x7e, 0xd1, 0xae, 0xc3, 0xac, 0x9f,
	0x78, 0x0e, 0x07, 0x5b, 0xf3, 0xed, 0x7b, 0xb2, 0xf6, 0x58, 0x04, 0x0f, 0x6a, 0xfe, 0x0a, 0x13,
	0xbd, 0xcc, 0xde, 0xbd, 0xd4, 0x9e, 0x11, 0x8c, 0x99, 0x24, 0xb3, 0x79, 0xeb, 0x78, 0xa6, 0xfb,
	0x13, 0x6f, 0xe0, 0x84, 0xc9, 0x55, 0xb6, 0x74, 0xd4, 0x09, 0x04, 0x8d, 0xc0, 0x5e, 0xfa, 0x71,
	0xd8, 0xfc, 0xe1, 0xdd, 0x5f, 0x00, 0x00, 0x00, 0xff, 0xff, 0x0d, 0x17, 0x93, 0xec, 0x51, 0x01,
	0x00, 0x00,
}
