package main

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"goim/libs/proto"
	"code.yy.com/yytars/goframework/tars/servant"
	"code.yy.com/yytars/goframework/jce/taf"
	pb "github.com/golang/protobuf/proto"
	log "github.com/thinkboy/log4go"
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

func invoke(comm *servant.Communicator, input proto.RPCInput) (output proto.RPCOutput, err error) {
	var (
		rpcStub *servant.ServantProxy
		rpcResp *taf.ResponsePacket
	)
	rpcStub = comm.GetServantProxy(input.Obj)
	rpcResp, err = rpcStub.Taf_invoke(context.TODO(), 0, input.Func, input.Req, nil, input.Opt)
	if err != nil {
		err = fmt.Errorf("rpc.invoke error: %v", err)
		log.Error("%v", err)
		output.Ret = 1
		output.Desc = err.Error()
		output.Obj = input.Obj
		output.Func = input.Func
		return
	}
	output.Ret = rpcResp.IRet
	output.Desc = rpcResp.SResultDesc
	output.Rsp = rpcResp.SBuffer
	output.Opt = rpcResp.Context
	output.Obj = input.Obj
	output.Func = input.Func
	return
}

// RPCInvoker knows how to interpret Proto.Body and invoke downstream service
type RPCInvoker interface {
	Decode(body json.RawMessage) (input proto.RPCInput, err error)
	Invoke(input proto.RPCInput) (output proto.RPCOutput, err error)
	Encode(output proto.RPCOutput) (body json.RawMessage, err error)
}

type WebsocketToRPC struct {
	comm *servant.Communicator
}

func NewWebsocketToRPC() RPCInvoker {
	return WebsocketToRPC{
		comm: comm,
	}
}

func (ws WebsocketToRPC) Decode(body json.RawMessage) (input proto.RPCInput, err error) {
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

func (ws WebsocketToRPC) Invoke(input proto.RPCInput) (output proto.RPCOutput, err error) {
	return invoke(ws.comm, input)
}

func (ws WebsocketToRPC) Encode(output proto.RPCOutput) (body json.RawMessage, err error) {
	log.Debug("rpc.ret = %d", output.Ret)
	log.Debug("rpc.rsp = \n%s", hex.Dump(output.Rsp))
	log.Debug("rpc.opt = %v", output.Opt)
	log.Debug("rpc.desc = %s", output.Desc)
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

func (t TCPToRPC) Decode(body json.RawMessage) (input proto.RPCInput, err error) {
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

func (t TCPToRPC) Invoke(input proto.RPCInput) (output proto.RPCOutput, err error) {
	return invoke(t.comm, input)
}

func (t TCPToRPC) Encode(output proto.RPCOutput) (body json.RawMessage, err error) {
	log.Debug("rpc.ret = %d", output.Ret)
	log.Debug("rpc.rsp = \n%s", hex.Dump(output.Rsp))
	log.Debug("rpc.opt = %v", output.Opt)
	log.Debug("rpc.desc = %s", output.Desc)
	if body, err = pb.Marshal(&output); err != nil {
		log.Error("can not encode rpc output to protobuf: %v", err)
		return
	}
	return
}
