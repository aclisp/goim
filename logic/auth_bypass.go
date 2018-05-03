package main

import (
	//"encoding/hex"
	"goim/libs/define"
	"goim/libs/proto"
	"strconv"

	log "github.com/aclisp/log4go"
	pb "github.com/golang/protobuf/proto"
)

type BypassAuther struct {
}

func NewBypassAuther() Auther {
	return &BypassAuther{}
}

func (a *BypassAuther) Auth(body []byte) (userId int64, roomId int64, err error) {
	//log.Debug("auth enter. body is \n%s", hex.Dump(body))
	var appId int64 = 0
	userId = 0
	roomId = define.NoRoom
	defer func() {
		log.Debug("auth return. appid=%v, uid=%v, room=%v", appId, userId, roomId)
	}()
	input := proto.RPCInput{}
	if err = pb.Unmarshal(body, &input); err != nil {
		log.Debug("auth body is not a valid protobuf: %v", err)
		return authWithString(string(body))
	}
	if _, ok := input.Opt[define.AppID]; ok {
		if appId, err = strconv.ParseInt(input.Opt[define.AppID], 10, 16); err != nil {
			return
		}
	}
	if _, ok := input.Opt[define.UID]; ok {
		if userId, err = strconv.ParseInt(input.Opt[define.UID], 10, 48); err != nil {
			return
		}
	}
	if _, ok := input.Opt[define.SubscribeRoom]; ok {
		if roomId, err = strconv.ParseInt(input.Opt[define.SubscribeRoom], 10, 48); err != nil {
			return
		}
	}
	userId = (appId << 48) | userId
	roomId = (appId << 48) | roomId
	return
}
