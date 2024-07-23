package main

import (
	inet "goim/libs/net"
	"goim/libs/net/xrpc"
	"goim/libs/proto"
	"math/rand"
	"time"

	"strings"

	log "github.com/aclisp/log4go"
)

var (
	logicServiceSet []*xrpc.Clients
)

const (
	logicService           = "RPC"
	logicServicePing       = "RPC.Ping"
	logicServiceConnect    = "RPC.Connect"
	logicServiceDisconnect = "RPC.Disconnect"
	logicServiceChangeRoom = "RPC.ChangeRoom"
	logicServiceUpdate     = "RPC.Update"
	logicServiceRegister   = "RPC.Register"
)

func InitLogicRpc(addrs map[string]struct{}) (err error) {
	var (
		network, addr string
	)
	for bind := range addrs {
		var rpcOptions []xrpc.ClientOptions
		for _, bind = range strings.Split(bind, ",") {
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
		rpcClient := xrpc.Dials(rpcOptions)
		// ping & reconnect
		rpcClient.Ping(logicServicePing)
		logicServiceSet = append(logicServiceSet, rpcClient)
		log.Info("init logic rpc: %v", rpcOptions)
	}
	return
}

func getLogic() (c *xrpc.Clients, err error) {
	n := len(logicServiceSet)
	r := rand.Intn(n)
	for i := 0; i < n; i++ {
		c = logicServiceSet[r%n]
		if err = c.Available(); err == nil {
			break
		}
		r++
	}
	return
}

func connect(p *proto.Proto) (key string, rid int64, heartbeat time.Duration, err error) {
	var (
		arg    = proto.ConnArg{Body: p.Body, Server: Conf.ServerId}
		reply  = proto.ConnReply{}
		client *xrpc.Clients
	)
	if client, err = getLogic(); err != nil {
		return
	}
	if err = client.Call(logicServiceConnect, &arg, &reply); err != nil {
		log.Error("c.Call(\"%s\", \"%v\", &ret) error(%v)", logicServiceConnect, arg, err)
		return
	}
	key = reply.Key
	rid = reply.RoomId
	heartbeat = Conf.IdleTimeout
	return
}

func disconnect(key string, roomId int64) (has bool, err error) {
	var (
		arg    = proto.DisconnArg{Key: key, RoomId: roomId}
		reply  = proto.DisconnReply{}
		client *xrpc.Clients
	)
	if client, err = getLogic(); err != nil {
		return
	}
	if err = client.Call(logicServiceDisconnect, &arg, &reply); err != nil {
		log.Error("c.Call(\"%s\", \"%v\", &ret) error(%v)", logicServiceDisconnect, arg, err)
		return
	}
	has = reply.Has
	return
}

func changeRoom(key string, orid int64, rid int64) (has bool, err error) {
	var (
		arg    = proto.ChangeRoomArg{Key: key, OldRoomId: orid, RoomId: rid, Server: Conf.ServerId}
		reply  = proto.ChangeRoomReply{}
		client *xrpc.Clients
	)
	if client, err = getLogic(); err != nil {
		return
	}
	if err = client.Call(logicServiceChangeRoom, &arg, &reply); err != nil {
		log.Error("c.Call(\"%s\", \"%v\", &ret) error(%v)", logicServiceChangeRoom, arg, err)
	}
	has = reply.Has
	return
}

func update(key string, roomId int64) (err error) {
	var (
		arg    = proto.UpdateArg{Key: key, RoomId: roomId, Server: Conf.ServerId}
		reply  = proto.NoReply{}
		client *xrpc.Clients
	)
	if client, err = getLogic(); err != nil {
		return
	}
	if err = client.Call(logicServiceUpdate, &arg, &reply); err != nil {
		log.Error("c.Call(\"%s\", \"%v\", &ret) error(%v)", logicServiceUpdate, arg, err)
	}
	return
}

func register() (err error) {
	var (
		arg    = proto.RegisterArg{Server: Conf.ServerId, Info: strings.Join(Conf.AdvertisedAddrs, ",")}
		reply  = proto.NoReply{}
		client *xrpc.Clients
	)
	if client, err = getLogic(); err != nil {
		return
	}
	if err = client.Call(logicServiceRegister, &arg, &reply); err != nil {
		log.Error("c.Call(\"%s\", \"%v\", &ret) error(%v)", logicServiceRegister, arg, err)
	}
	return
}
