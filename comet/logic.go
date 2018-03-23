package main

import (
	inet "goim/libs/net"
	"goim/libs/net/xrpc"
	"goim/libs/proto"
	"time"

	log "github.com/thinkboy/log4go"
	"strings"
)

var (
	logicRpcClient *xrpc.Clients
	logicRpcQuit   = make(chan struct{}, 1)

	logicService           = "RPC"
	logicServicePing       = "RPC.Ping"
	logicServiceConnect    = "RPC.Connect"
	logicServiceDisconnect = "RPC.Disconnect"
	logicServiceChangeRoom = "RPC.ChangeRoom"
	logicServiceUpdate     = "RPC.Update"
	logicServiceRegister   = "RPC.Register"
)

func InitLogicRpc(addrs []string) (err error) {
	var (
		bind          string
		network, addr string
		rpcOptions    []xrpc.ClientOptions
	)
	for _, bind = range addrs {
		if network, addr, err = inet.ParseNetwork(bind); err != nil {
			log.Error("inet.ParseNetwork() error(%v)", err)
			return
		}
		options := xrpc.ClientOptions{
			Proto: network,
			Addr:  addr,
		}
		rpcOptions = append(rpcOptions, options)
	}
	// rpc clients
	logicRpcClient = xrpc.Dials(rpcOptions)
	// ping & reconnect
	logicRpcClient.Ping(logicServicePing)
	log.Info("init logic rpc: %v", rpcOptions)
	return
}

func connect(p *proto.Proto) (key string, rid int64, heartbeat time.Duration, err error) {
	var (
		arg   = proto.ConnArg{Token: string(p.Body), Server: Conf.ServerId}
		reply = proto.ConnReply{}
	)
	if err = logicRpcClient.Call(logicServiceConnect, &arg, &reply); err != nil {
		log.Error("c.Call(\"%s\", \"%v\", &ret) error(%v)", logicServiceConnect, arg, err)
		return
	}
	key = reply.Key
	rid = reply.RoomId
	heartbeat = 5 * 60 * time.Second
	return
}

func disconnect(key string, roomId int64) (has bool, err error) {
	var (
		arg   = proto.DisconnArg{Key: key, RoomId: roomId}
		reply = proto.DisconnReply{}
	)
	if err = logicRpcClient.Call(logicServiceDisconnect, &arg, &reply); err != nil {
		log.Error("c.Call(\"%s\", \"%v\", &ret) error(%v)", logicServiceDisconnect, arg, err)
		return
	}
	has = reply.Has
	return
}

func changeRoom(key string, orid int64, rid int64) (has bool, err error) {
	var (
		arg   = proto.ChangeRoomArg{Key: key, OldRoomId: orid, RoomId: rid}
		reply = proto.ChangeRoomReply{}
	)
	if err = logicRpcClient.Call(logicServiceChangeRoom, &arg, &reply); err != nil {
		log.Error("c.Call(\"%s\", \"%v\", &ret) error(%v)", logicServiceChangeRoom, arg, err)
	}
	has = reply.Has
	return
}

func update(key string, roomId int64) (err error) {
	var (
		arg   = proto.UpdateArg{Key: key, RoomId: roomId, Server: Conf.ServerId}
		reply = proto.NoReply{}
	)
	if err = logicRpcClient.Call(logicServiceUpdate, &arg, &reply); err != nil {
		log.Error("c.Call(\"%s\", \"%v\", &ret) error(%v)", logicServiceUpdate, arg, err)
	}
	return
}

func register() (err error) {
	var (
		arg   = proto.RegisterArg{Server: Conf.ServerId, Info: strings.Join(Conf.WebsocketTLSBind, ",")}
		reply = proto.NoReply{}
	)
	if err = logicRpcClient.Call(logicServiceRegister, &arg, &reply); err != nil {
		log.Error("c.Call(\"%s\", \"%v\", &ret) error(%v)", logicServiceRegister, arg, err)
	}
	return
}