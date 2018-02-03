package main

import (
	"goim/libs/define"
	"strconv"
	log "github.com/thinkboy/log4go"
)

// developer could implement "Auth" interface for decide how get userId, or roomId
type Auther interface {
	Auth(token string) (userId int64, roomId int32)
}

type DefaultAuther struct {
}

func NewDefaultAuther() *DefaultAuther {
	return &DefaultAuther{}
}

func (a *DefaultAuther) Auth(token string) (userId int64, roomId int32) {
	log.Info("Auth token is %s", token)
	var err error
	userId = 0
	roomId = define.NoRoom
	defer func() {
		log.Info("Auth userId is %v, roomId is %v", userId, roomId)
	}()
	if len(token) < 2 {
		return
	}
	token = token[1 : len(token)-1]
	if userId, err = strconv.ParseInt(token, 10, 64); err != nil {
		return
	}
	roomId = 1 // only for debug
	return
}
