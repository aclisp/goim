package main

import (
	inet "goim/libs/net"
	"goim/libs/net/xrpc"
	"goim/libs/proto"

	log "github.com/thinkboy/log4go"
)

var (
	jobRpcClient *xrpc.Clients
)

const (
	jobServicePing = "JobRPC.Ping"
	jobServicePush = "JobRPC.Push"
)

func InitJobRpc(addrs []string) (err error) {
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
	jobRpcClient = xrpc.Dials(rpcOptions)
	// ping & reconnect
	jobRpcClient.Ping(jobServicePing)
	log.Info("init job rpc: %v", rpcOptions)
	return
}

func push(arg *proto.KafkaMsg) (err error) {
	var reply = proto.NoReply{}
	if err = jobRpcClient.Call(jobServicePush, arg, &reply); err != nil {
		log.Error("c.Call(\"%s\", \"%+v\") error(%v)", jobServicePush, *arg, err)
		return
	}
	return
}
