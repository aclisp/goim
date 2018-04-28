package main

import (
	"encoding/json"
	inet "goim/libs/net"
	"goim/libs/proto"
	"net"
	"net/rpc"

	log "github.com/aclisp/log4go"
	pb "github.com/golang/protobuf/proto"
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

// Connect auth and login
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
		// Determine encoding type
		pbenc := true
		input := proto.RPCInput{}
		if err = pb.Unmarshal(arg.Body, &input); err != nil {
			pbenc = false
		}
		// Notify other clients that I just logged in
		subKeys := genSubKey(uid)
		for serverId, keys := range subKeys {
			others := keys[:0]
			for _, x := range keys {
				if reply.Key != x {
					others = append(others, x)
				}
			}
			if len(others) == 0 {
				continue
			}
			msg := ServerPush{MessageType: 9}
			var buf []byte
			if pbenc {
				buf, err = pb.Marshal(&msg)
			} else {
				buf, err = json.Marshal(&msg)
			}
			if err != nil {
				log.Warn("Connect() notify others, marshal error(%v)", err)
				continue
			}
			if err = mpushKafka(serverId, others, buf, false); err != nil {
				log.Warn("Connect() notify others, mpush error(%v)", err)
			}
		}
		err = nil
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
