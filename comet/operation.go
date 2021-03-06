package main

import (
	"goim/libs/define"
	"goim/libs/encoding/binary"
	"goim/libs/proto"
	"time"

	log "github.com/aclisp/log4go"
)

type ConnType int

const (
	WebsocketConn ConnType = iota
	TCPConn
)

var currentTimeMillisBytes [8]byte

type Operator interface {
	// Operate process the common operation such as send message etc.
	Operate(*proto.Proto, ConnType, map[string]string) error
	// Direct is a temp method only used to quick enter room.
	Direct(proto.RPCInput, ConnType, map[string]string) (proto.RPCOutput, error)
	// Connect used for auth user and return a subkey, roomid, hearbeat.
	Connect(*proto.Proto) (string, int64, time.Duration, error)
	// Disconnect used for revoke the subkey.
	Disconnect(string, int64) error
	// ChangeRoom changes from old roomid to new roomid for the subkey.
	ChangeRoom(string, int64, int64) error
	// Update keeps the latest online info for the subkey.
	Update(string, int64) error
	// Register this comet instance
	Register() error
}

type DefaultOperator struct {
	WebsocketToRPC RPCInvoker
	TCPToRPC       RPCInvoker
}

func NewOperator() Operator {
	return &DefaultOperator{
		WebsocketToRPC: NewWebsocketToRPC(),
		TCPToRPC:       NewTCPToRPC(),
	}
}

func (operator *DefaultOperator) Operate(p *proto.Proto, connType ConnType, optMerge map[string]string) error {
	var (
		invoker RPCInvoker
	)
	if connType == TCPConn {
		invoker = operator.TCPToRPC
	} else {
		invoker = operator.WebsocketToRPC
	}
	if p.Operation == define.OP_SEND_SMS {
		input, err := invoker.Decode(p.Body)
		if err != nil {
			return err
		}
		for k, v := range optMerge {
			if input.Opt == nil {
				input.Opt = make(map[string]string, len(optMerge))
			}
			input.Opt[k] = v
		}
		output, err := invoker.Invoke(input)
		if err != nil {
			//return err
		}
		p.Body, err = invoker.Encode(output)
		if err != nil {
			return err
		}
		p.Operation = define.OP_SEND_SMS_REPLY
	} else if p.Operation == define.OP_TEST {
		binary.BigEndian.PutInt64(currentTimeMillisBytes[:], time.Now().UnixNano()/1000000)
		p.Body = currentTimeMillisBytes[:]
		p.Operation = define.OP_TEST_REPLY
	} else {
		return ErrOperation
	}
	return nil
}

func (operator * DefaultOperator) Direct(input proto.RPCInput, connType ConnType, optMerge map[string]string) (output proto.RPCOutput, err error) {
	var (
		invoker RPCInvoker
	)
	if connType == TCPConn {
		invoker = operator.TCPToRPC
	} else {
		invoker = operator.WebsocketToRPC
	}
	for k, v := range optMerge {
		if input.Opt == nil {
			input.Opt = make(map[string]string, len(optMerge))
		}
		input.Opt[k] = v
	}
	output, err = invoker.Invoke(input)
	return
}

func (operator *DefaultOperator) Connect(p *proto.Proto) (key string, rid int64, heartbeat time.Duration, err error) {
	key, rid, heartbeat, err = connect(p)
	return
}

func (operator *DefaultOperator) Disconnect(key string, rid int64) (err error) {
	var has bool
	if has, err = disconnect(key, rid); err != nil {
		return
	}
	if !has {
		log.Warn("disconnect key: \"%s\" not exists", key)
	}
	return
}

func (operator *DefaultOperator) ChangeRoom(key string, orid int64, rid int64) (err error) {
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

func (operator *DefaultOperator) Update(key string, rid int64) (err error) {
	err = update(key, rid)
	return
}

func (Operator *DefaultOperator) Register() (err error ) {
	err = register()
	return
}