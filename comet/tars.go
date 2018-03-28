package main

import (
	"code.yy.com/yytars/goframework/tars/servant"
	"code.yy.com/yytars/goframework/jce/taf"
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	log "github.com/thinkboy/log4go"
	pb "github.com/golang/protobuf/proto"
)

// setup yytars communicator
// servant package init will be called first!
//     It should read tars config file during init.
/* NEED THIS PATCH
--- a/tars/servant/Application.go
+++ b/tars/servant/Application.go
@@ -1,7 +1,7 @@
package servant

import (
-       "flag"
+       //"flag"
	"code.yy.com/yytars/goframework/kissgo/appzaplog/zap"
	"net/http"
	"os"
@@ -26,9 +26,9 @@ var (
)

func initConfig() {
-       _configFile := (flag.String("config", "", "init config path"))
-       flag.Parse()
-       configFile = *_configFile
+       //_configFile := (flag.String("config", "", "init config path"))
+       //flag.Parse()
+       configFile = "tars-config.conf"
	if len(configFile) == 0 {
			appzaplog.SetLogLevel("info")
			return
 */
var comm = servant.NewPbCommunicator()

const _ = pb.ProtoPackageIsVersion2

// RPCInput has all necessary information when calling downstream service (Now it is Tars)
type RPCInput struct {
	Obj  string            `protobuf:"bytes,1,opt,name=obj" json:"obj,omitempty"`
	Func string            `protobuf:"bytes,2,opt,name=func" json:"func,omitempty"`
	Req  json.RawMessage   `protobuf:"bytes,3,opt,name=req,proto3" json:"req,omitempty"`
	Opt  map[string]string `protobuf:"bytes,4,rep,name=opt" json:"opt,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

func (m *RPCInput) Reset()                    { *m = RPCInput{} }
func (m *RPCInput) String() string            { return pb.CompactTextString(m) }
func (*RPCInput) ProtoMessage()               {}

// RPCOutput is what the downstream service returns
type RPCOutput struct {
	Ret  int32             `protobuf:"zigzag32,1,opt,name=ret" json:"ret,omitempty"`
	Rsp  json.RawMessage   `protobuf:"bytes,2,opt,name=rsp,proto3" json:"rsp,omitempty"`
	Opt  map[string]string `protobuf:"bytes,3,rep,name=opt" json:"opt,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

func (m *RPCOutput) Reset()                    { *m = RPCOutput{} }
func (m *RPCOutput) String() string            { return pb.CompactTextString(m) }
func (*RPCOutput) ProtoMessage()               {}

// RPCInvoker knows how to interpret Proto.Body and invoke downstream service
type RPCInvoker interface {
	Decode(body json.RawMessage) (input RPCInput, err error)
	Invoke(input RPCInput) (output RPCOutput, err error)
	Encode(output RPCOutput) (body json.RawMessage, err error)
}

type WebsocketToRPC struct {
	comm *servant.Communicator
}

func NewWebsocketToRPC() RPCInvoker {
	return WebsocketToRPC{
		comm: comm,
	}
}

func (ws WebsocketToRPC) Decode(body json.RawMessage) (input RPCInput, err error) {
	log.Debug("proto.body = \n%s", hex.Dump(body))
	if err = json.Unmarshal(body, &input); err != nil {
		log.Error("proto.body is not a valid json: %v", err)
		return
	}
	log.Debug("rpc.obj = %s", input.Obj)
	log.Debug("rpc.func = %s", input.Func)
	if len(input.Req) < 2 {
		err = fmt.Errorf("rpc.req is not a json string: %s", input.Req)
		log.Error("%v", err)
		return
	}
	input.Req = input.Req[1:len(input.Req)-1]
	if input.Req, err = base64.StdEncoding.DecodeString(string(input.Req)); err != nil {
		err = fmt.Errorf("rpc.req can not be decode to hex: %s", input.Req)
		log.Error("%v", err)
		return
	}
	log.Debug("rpc.req = \n%s", hex.Dump(input.Req))
	log.Debug("rpc.opt = %v", input.Opt)
	return
}

func invoke(comm *servant.Communicator, input RPCInput) (output RPCOutput, err error) {
	var (
		rpcStub *servant.ServantProxy
		rpcResp *taf.ResponsePacket
	)
	rpcStub = comm.GetServantProxy(input.Obj)
	rpcResp, err = rpcStub.Taf_invoke(context.TODO(), 0, input.Func, input.Req, nil, input.Opt)
	if err != nil {
		log.Error("rpc.invoke error: %v", err)
		return
	}
	output.Ret = rpcResp.IRet
	output.Rsp = rpcResp.SBuffer
	output.Opt = rpcResp.Context
	return
}

func (ws WebsocketToRPC) Invoke(input RPCInput) (output RPCOutput, err error) {
	return invoke(ws.comm, input)
}

func (ws WebsocketToRPC) Encode(output RPCOutput) (body json.RawMessage, err error) {
	log.Debug("rpc.ret = %d", output.Ret)
	log.Debug("rpc.rsp = \n%s", hex.Dump(output.Rsp))
	log.Debug("rpc.opt = %v", output.Opt)
	output.Rsp = []byte(`"` + base64.StdEncoding.EncodeToString(output.Rsp) + `"`)
	if body, err = json.Marshal(output); err != nil {
		log.Error("can not encode rpc output to json: %v", err)
		return
	}
	return
}

type TCPToRPC struct {
	comm *servant.Communicator
}

func NewTCPToRPC() RPCInvoker {
	return TCPToRPC{
		comm: comm,
	}
}

func (t TCPToRPC) Decode(body json.RawMessage) (input RPCInput, err error) {
	log.Debug("proto.body = \n%s", hex.Dump(body))
	if err = pb.Unmarshal(body, &input); err != nil {
		log.Error("proto.body is not a valid protobuf: %v", err)
		return
	}
	log.Debug("rpc.obj = %s", input.Obj)
	log.Debug("rpc.func = %s", input.Func)
	log.Debug("rpc.req = \n%s", hex.Dump(input.Req))
	log.Debug("rpc.opt = %v", input.Opt)
	return
}

func (t TCPToRPC) Invoke(input RPCInput) (output RPCOutput, err error) {
	return invoke(t.comm, input)
}

func (t TCPToRPC) Encode(output RPCOutput) (body json.RawMessage, err error) {
	log.Debug("rpc.ret = %d", output.Ret)
	log.Debug("rpc.rsp = \n%s", hex.Dump(output.Rsp))
	log.Debug("rpc.opt = %v", output.Opt)
	if body, err = pb.Marshal(&output); err != nil {
		log.Error("can not encode rpc output to protobuf: %v", err)
		return
	}
	return
}
