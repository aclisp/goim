package main

import (
	inet "goim/libs/net"
	"goim/libs/proto"
	"net"
	"net/rpc"

	log "github.com/aclisp/log4go"
)

func InitRPC(auther Auther) (err error) {
	var (
		network, addr string
		c             = &RPC{auther: auther}
	)
	rpc.Register(c)
	for i := 0; i < len(Conf.RPCAddrs); i++ {
		log.Info("start listen rpc addr: \"%s\"", Conf.RPCAddrs[i])
		if network, addr, err = inet.ParseNetwork(Conf.RPCAddrs[i]); err != nil {
			log.Error("inet.ParseNetwork() error(%v)", err)
			return
		}
		go rpcListen(network, addr)
	}
	return
}

func rpcListen(network, addr string) {
	l, err := net.Listen(network, addr)
	if err != nil {
		log.Error("net.Listen(\"%s\", \"%s\") error(%v)", network, addr, err)
		panic(err)
	}
	// if process exit, then close the rpc bind
	defer func() {
		log.Info("rpc addr: \"%s\" close", addr)
		if err := l.Close(); err != nil {
			log.Error("listener.Close() error(%v)", err)
		}
	}()
	rpc.Accept(l)
}

// RPC
type RPC struct {
	auther Auther
}

func (r *RPC) Ping(arg *proto.NoArg, reply *proto.NoReply) error {
	return nil
}

// Connect auth and registe login
func (r *RPC) Connect(arg *proto.ConnArg, reply *proto.ConnReply) (err error) {
	if arg == nil {
		err = ErrConnectArgs
		log.Error("Connect() error(%v)", err)
		return
	}
	var (
		uid int64
		seq int32
	)
	if uid, reply.RoomId, err = r.auther.Auth(arg.Body); err != nil {
		log.Warn("Connect() auth error(%v)", err)
		return
	}
	if seq, err = connect(uid, arg.Server, reply.RoomId); err == nil {
		reply.Key = encode(uid, seq)
	}
	return
}

// Disconnect notice router offline
func (r *RPC) Disconnect(arg *proto.DisconnArg, reply *proto.DisconnReply) (err error) {
	if arg == nil {
		err = ErrDisconnectArgs
		log.Error("Disconnect() error(%v)", err)
		return
	}
	var (
		uid int64
		seq int32
	)
	if uid, seq, err = decode(arg.Key); err != nil {
		log.Error("decode(\"%s\") error(%s)", arg.Key, err)
		return
	}
	reply.Has, err = disconnect(uid, seq, arg.RoomId)
	return
}

func (r *RPC) ChangeRoom(arg *proto.ChangeRoomArg, reply *proto.ChangeRoomReply) (err error) {
	if arg == nil {
		err = ErrChangeRoomArgs
		log.Error("ChangeRoom() error(%v)", err)
		return
	}
	var (
		uid int64
		seq int32
	)
	if uid, seq, err = decode(arg.Key); err != nil {
		log.Error("decode(\"%s\") error(%s)", arg.Key, err)
		return
	}
	reply.Has, err = changeRoom(uid, seq, arg.OldRoomId, arg.RoomId)
	return
}

func (r *RPC) Update(arg *proto.UpdateArg, reply *proto.NoReply) (err error) {
	if arg == nil {
		err = ErrUpdateArgs
		log.Error("Update() error(%v)", err)
		return
	}
	var (
		uid int64
		seq int32
	)
	if uid, seq, err = decode(arg.Key); err != nil {
		log.Error("decode(\"%s\") error(%s)", arg.Key, err)
		return
	}
	err = update(uid, seq, arg.Server, arg.RoomId)
	return
}

func (r *RPC) Register(arg *proto.RegisterArg, reply *proto.NoReply) (err error) {
	if arg == nil {
		err = ErrRegisterArgs
		log.Error("Register() error(%v)", err)
		return
	}
	err = register(arg, reply)
	return
}