package main

import (
	inet "goim/libs/net"
	"goim/libs/proto"
	"net"
	"net/rpc"

	log "github.com/thinkboy/log4go"
)

func InitRPC() (err error) {
	var (
		network, addr string
		c             = &JobRPC{}
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

// JobRPC is used to replace Kafka
type JobRPC struct {
}

func (this *JobRPC) Ping(arg *proto.NoArg, reply *proto.NoReply) error {
	return nil
}

func (this *JobRPC) Push(arg *proto.KafkaMsg, reply *proto.NoReply) error {
	producePush(arg)
	return nil
}
