// Code generated by protoc-gen-tars. DO NOT EDIT.
// source: dubboproxy.proto

package bilin

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "context"
	"code.yy.com/yytars/goframework/tars/servant"
	"code.yy.com/yytars/goframework/tars/servant/model"
	"code.yy.com/yytars/goframework/jce/taf"
	"errors"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type DPInvokeReq struct {
	Service string         `protobuf:"bytes,1,opt,name=service" json:"service,omitempty"`
	Method  string         `protobuf:"bytes,2,opt,name=method" json:"method,omitempty"`
	Args    []*DPInvokeArg `protobuf:"bytes,3,rep,name=args" json:"args,omitempty"`
}

func (m *DPInvokeReq) Reset()                    { *m = DPInvokeReq{} }
func (m *DPInvokeReq) String() string            { return proto.CompactTextString(m) }
func (*DPInvokeReq) ProtoMessage()               {}
func (*DPInvokeReq) Descriptor() ([]byte, []int) { return fileDescriptor3, []int{0} }

func (m *DPInvokeReq) GetService() string {
	if m != nil {
		return m.Service
	}
	return ""
}

func (m *DPInvokeReq) GetMethod() string {
	if m != nil {
		return m.Method
	}
	return ""
}

func (m *DPInvokeReq) GetArgs() []*DPInvokeArg {
	if m != nil {
		return m.Args
	}
	return nil
}

type DPInvokeArg struct {
	Type  string `protobuf:"bytes,1,opt,name=type" json:"type,omitempty"`
	Value string `protobuf:"bytes,2,opt,name=value" json:"value,omitempty"`
}

func (m *DPInvokeArg) Reset()                    { *m = DPInvokeArg{} }
func (m *DPInvokeArg) String() string            { return proto.CompactTextString(m) }
func (*DPInvokeArg) ProtoMessage()               {}
func (*DPInvokeArg) Descriptor() ([]byte, []int) { return fileDescriptor3, []int{1} }

func (m *DPInvokeArg) GetType() string {
	if m != nil {
		return m.Type
	}
	return ""
}

func (m *DPInvokeArg) GetValue() string {
	if m != nil {
		return m.Value
	}
	return ""
}

type DPInvokeRsp struct {
	Type           string `protobuf:"bytes,1,opt,name=type" json:"type,omitempty"`
	Value          string `protobuf:"bytes,2,opt,name=value" json:"value,omitempty"`
	ThrewException bool   `protobuf:"varint,3,opt,name=threw_exception,json=threwException" json:"threw_exception,omitempty"`
}

func (m *DPInvokeRsp) Reset()                    { *m = DPInvokeRsp{} }
func (m *DPInvokeRsp) String() string            { return proto.CompactTextString(m) }
func (*DPInvokeRsp) ProtoMessage()               {}
func (*DPInvokeRsp) Descriptor() ([]byte, []int) { return fileDescriptor3, []int{2} }

func (m *DPInvokeRsp) GetType() string {
	if m != nil {
		return m.Type
	}
	return ""
}

func (m *DPInvokeRsp) GetValue() string {
	if m != nil {
		return m.Value
	}
	return ""
}

func (m *DPInvokeRsp) GetThrewException() bool {
	if m != nil {
		return m.ThrewException
	}
	return false
}

func init() {
	proto.RegisterType((*DPInvokeReq)(nil), "bilin.tars.dubboproxy.DPInvokeReq")
	proto.RegisterType((*DPInvokeArg)(nil), "bilin.tars.dubboproxy.DPInvokeArg")
	proto.RegisterType((*DPInvokeRsp)(nil), "bilin.tars.dubboproxy.DPInvokeRsp")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context

// Client API for DubboProxy service

type DubboProxyClient interface {
	Invoke(ctx context.Context, in *DPInvokeReq, opts ...map[string]string) (*DPInvokeRsp, error)
}

type dubboProxyClient struct {
	s model.Servant
}

func NewDubboProxyClient(objname string, comm servant.ICommunicator) DubboProxyClient {
	if comm == nil || objname == "" {
		return nil
	}
	return &dubboProxyClient{s: comm.GetServantProxy(objname)}
}

func (c *dubboProxyClient) Invoke(ctx context.Context, in *DPInvokeReq, opts ...map[string]string) (*DPInvokeRsp, error) {
	var (
		reply DPInvokeRsp
	)

	pbbuf, err := proto.Marshal(in)
	if err != nil {
		return nil, err
	}

	_resp, err := c.s.Taf_invoke(ctx, 0, "Invoke", pbbuf)
	if err != nil {
		return nil, err
	}

	if err = proto.Unmarshal(_resp.SBuffer, &reply); err != nil {
		return nil, err
	}
	return &reply, nil
}

// Server API for DubboProxy service

type DubboProxyServer interface {
	Invoke(context.Context, *DPInvokeReq) (*DPInvokeRsp, error)
}

type dubboProxyDispatcher struct {
}

func NewDubboProxyDispatcher() servant.Dispatcher {
	return &dubboProxyDispatcher{}
}

func (_obj *dubboProxyDispatcher) Dispatch(ctx context.Context, _val interface{}, req *taf.RequestPacket) (*taf.ResponsePacket, error) {
	var pbbuf []byte
	_imp := _val.(DubboProxyServer)
	switch req.SFuncName {
	case "Invoke":
		var req_ DPInvokeReq
		if err := proto.Unmarshal(req.SBuffer, &req_); err != nil {
			return nil, err
		}

		_ret, err := _imp.Invoke(ctx, &req_)
		if err != nil {
			return nil, err
		}

		if pbbuf, err = proto.Marshal(_ret); err != nil {
			return nil, err
		}

	default:
		return nil, errors.New("unknow func")
	}
	return &taf.ResponsePacket{
		IVersion:   1,
		IRequestId: req.IRequestId,
		SBuffer:    pbbuf,
		Context:    req.Context,
	}, nil
}

func init() { proto.RegisterFile("dubboproxy.proto", fileDescriptor3) }

var fileDescriptor3 = []byte{
	// 262 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x91, 0xc1, 0x4b, 0xf3, 0x40,
	0x10, 0xc5, 0xc9, 0x97, 0x26, 0x9f, 0x4e, 0x41, 0x65, 0x50, 0x59, 0xc4, 0x43, 0xc8, 0xc5, 0x1c,
	0x64, 0x0f, 0x15, 0xf4, 0xdc, 0x52, 0x0f, 0xde, 0xc2, 0x1e, 0x3d, 0xa8, 0x49, 0x3a, 0xa4, 0xc1,
	0x36, 0xbb, 0xdd, 0xdd, 0xa6, 0xcd, 0x7f, 0x2f, 0xd9, 0x36, 0xd4, 0x43, 0xa1, 0xde, 0xf6, 0xcd,
	0x7b, 0xcb, 0xef, 0x0d, 0x03, 0x57, 0xb3, 0x75, 0x9e, 0x4b, 0xa5, 0xe5, 0xb6, 0xe5, 0x4a, 0x4b,
	0x2b, 0xf1, 0x26, 0xaf, 0x16, 0x55, 0xcd, 0x6d, 0xa6, 0x0d, 0x3f, 0x98, 0xf1, 0x06, 0x86, 0xd3,
	0xf4, 0xad, 0x6e, 0xe4, 0x37, 0x09, 0x5a, 0x21, 0x83, 0xff, 0x86, 0x74, 0x53, 0x15, 0xc4, 0xbc,
	0xc8, 0x4b, 0xce, 0x45, 0x2f, 0xf1, 0x16, 0xc2, 0x25, 0xd9, 0xb9, 0x9c, 0xb1, 0x7f, 0xce, 0xd8,
	0x2b, 0x7c, 0x86, 0x41, 0xa6, 0x4b, 0xc3, 0xfc, 0xc8, 0x4f, 0x86, 0xa3, 0x98, 0x1f, 0xc5, 0xf0,
	0x9e, 0x31, 0xd6, 0xa5, 0x70, 0xf9, 0xf8, 0xe5, 0x00, 0x1e, 0xeb, 0x12, 0x11, 0x06, 0xb6, 0x55,
	0x3d, 0xd5, 0xbd, 0xf1, 0x1a, 0x82, 0x26, 0x5b, 0xac, 0x69, 0x4f, 0xdc, 0x89, 0xf8, 0xeb, 0x57,
	0x63, 0xa3, 0xfe, 0xfe, 0x11, 0x1f, 0xe0, 0xd2, 0xce, 0x35, 0x6d, 0x3e, 0x69, 0x5b, 0x90, 0xb2,
	0x95, 0xac, 0x99, 0x1f, 0x79, 0xc9, 0x99, 0xb8, 0x70, 0xe3, 0xd7, 0x7e, 0x3a, 0xfa, 0x00, 0x98,
	0x76, 0xd5, 0xd3, 0xae, 0x3a, 0xa6, 0x10, 0xee, 0x68, 0x78, 0x6a, 0x39, 0x41, 0xab, 0xbb, 0x93,
	0x19, 0xa3, 0x26, 0x8f, 0x70, 0x5f, 0xc8, 0x25, 0x6f, 0xdb, 0xe3, 0xd9, 0x49, 0x90, 0x76, 0x17,
	0x7b, 0x0f, 0x9c, 0x9b, 0x87, 0xee, 0x7e, 0x4f, 0x3f, 0x01, 0x00, 0x00, 0xff, 0xff, 0xc8, 0x74,
	0x9a, 0x57, 0xd3, 0x01, 0x00, 0x00,
}
