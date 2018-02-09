package main

import (
	"fmt"
	"encoding/json"
	"goim/libs/define"
	"goim/libs/proto"
	"time"

	log "github.com/thinkboy/log4go"
	"yytars/tars/servant"
	"context"
	"encoding/base64"
	"encoding/hex"
)

type Operator interface {
	// Operate process the common operation such as send message etc.
	Operate(*proto.Proto) error
	// Connect used for auth user and return a subkey, roomid, hearbeat.
	Connect(*proto.Proto) (string, int32, time.Duration, error)
	// Disconnect used for revoke the subkey.
	Disconnect(string, int32) error
	// ChangeRoom changes from old roomid to new roomid for the subkey.
	ChangeRoom(string, int32, int32) error
}

type DefaultOperator struct {
	Comm *servant.Communicator
}

func (operator *DefaultOperator) Operate(p *proto.Proto) error {
	var (
		body []byte
	)
	if p.Operation == define.OP_SEND_SMS {
		// call yytars api
		type RPC struct {
			Obj string `json:"obj"`
			Func string `json:"func"`
			Req json.RawMessage `json:"req"`
		}
		var rpc RPC
		json.Unmarshal(p.Body, &rpc)
		log.Info("rpc.obj = %s", rpc.Obj)
		log.Info("rpc.func = %s", rpc.Func)
		if len(rpc.Req) < 2 {
			err := fmt.Errorf("rpc.req is not a json string: %s", rpc.Req)
			log.Error("%v", err)
			return err
		}
		rpc.Req = rpc.Req[1:len(rpc.Req)-1]
		rpcReqBuf, err := base64.StdEncoding.DecodeString(string(rpc.Req))
		if err != nil {
			err := fmt.Errorf("rpc.req can not be decode to hex: %s", rpc.Req)
			log.Error("%v", err)
			return err
		}
		log.Info("rpc.req = \n%s", hex.Dump(rpcReqBuf))

		rpcStub := operator.Comm.GetServantProxy(rpc.Obj)
		rpcResp, err := rpcStub.Taf_invoke(context.TODO(), 0, rpc.Func, rpcReqBuf, nil, nil)
		if err != nil {
			log.Error("rpc.invoke error: %v", err)
			return err
		}

		log.Info("rpc.rsp = \n%s", hex.Dump(rpcResp.SBuffer))
		p.Body = []byte(`"` + base64.StdEncoding.EncodeToString(rpcResp.SBuffer) + `"`)
		p.Operation = define.OP_SEND_SMS_REPLY
	} else if p.Operation == define.OP_TEST {
		log.Debug("test operation: %s", body)
		p.Operation = define.OP_TEST_REPLY
		p.Body = []byte("{\"test\":\"come on\"}")
	} else {
		return ErrOperation
	}
	return nil
}

func (operator *DefaultOperator) Connect(p *proto.Proto) (key string, rid int32, heartbeat time.Duration, err error) {
	key, rid, heartbeat, err = connect(p)
	return
}

func (operator *DefaultOperator) Disconnect(key string, rid int32) (err error) {
	var has bool
	if has, err = disconnect(key, rid); err != nil {
		return
	}
	if !has {
		log.Warn("disconnect key: \"%s\" not exists", key)
	}
	return
}

func (operator *DefaultOperator) ChangeRoom(key string, orid int32, rid int32) (err error) {
	var has bool
	if orid == rid {
		return
	}
	if has, err = changeRoom(key, orid, rid); err != nil {
		return
	}
	if !has {
		log.Warn("change room key: \"%s\" not exists", key)
	}
	return
}
