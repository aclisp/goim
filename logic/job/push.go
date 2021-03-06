package main

import (
	"encoding/json"
	"goim/libs/define"
	"goim/libs/proto"
	"math/rand"

	log "github.com/aclisp/log4go"
)

type pushArg struct {
	ServerId int32
	SubKeys  []string
	Msg      []byte
	RoomId   int64
	Kick     bool
}

var (
	pushChs []chan *pushArg
	bufCh   chan *proto.KafkaMsg
)

func InitPush() {
	pushChs = make([]chan *pushArg, Conf.PushChan)
	bufCh = make(chan *proto.KafkaMsg, Conf.PushBufSize)
	for i := 0; i < Conf.PushChan; i++ {
		pushChs[i] = make(chan *pushArg, Conf.PushChanSize)
		go processPush(pushChs[i])
	}
	go consumePush()
}

// push routine
func processPush(ch chan *pushArg) {
	var arg *pushArg
	for {
		arg = <-ch
		mPushComet(arg.ServerId, arg.SubKeys, arg.Msg, arg.Kick)
	}
}

// push is called by kafka consumer group
func push(msg []byte) (err error) {
	m := &proto.KafkaMsg{}
	if err = json.Unmarshal(msg, m); err != nil {
		log.Error("json.Unmarshal(%s) error(%s)", msg, err)
		return
	}
	err = pushKafkaMsg(m)
	return
}

func pushKafkaMsg(m *proto.KafkaMsg) (err error) {
	switch m.OP {
	case define.KAFKA_MESSAGE_MULTI:
		pushChs[rand.Int()%Conf.PushChan] <- &pushArg{
			ServerId: m.ServerId,
			SubKeys:  m.SubKeys,
			Msg:      m.Msg,
			RoomId:   define.NoRoom,
			Kick:     m.Kick,
		}
	case define.KAFKA_MESSAGE_BROADCAST:
		broadcast(m.Msg)
	case define.KAFKA_MESSAGE_BROADCAST_ROOM:
		room := roomBucket.Get(m.RoomId)
		if m.Ensure {
			go room.EPush(0, define.OP_SEND_SMS_REPLY, m.Msg)
		} else {
			err = room.Push(0, define.OP_SEND_SMS_REPLY, m.Msg)
			if err != nil {
				log.Error("room.Push(%s) roomId:%d error(%v)", m.Msg, err)
			}
		}
	default:
		log.Error("unknown operation:%s", m.OP)
	}
	return
}

func producePush(m *proto.KafkaMsg) {
	bufCh <- m
}

func consumePush() {
	var m *proto.KafkaMsg
	for {
		m = <-bufCh
		pushKafkaMsg(m)
	}
}
