package main

import (
	"goim/libs/define"
	"strconv"
	log "github.com/thinkboy/log4go"
	"strings"
)

// developer could implement "Auth" interface for decide how get userId, or roomId
type Auther interface {
	Auth(token string) (userId int64, roomId int64)
}

type DefaultAuther struct {
}

func NewDefaultAuther() *DefaultAuther {
	return &DefaultAuther{}
}

func (a *DefaultAuther) Auth(token string) (userId int64, roomId int64) {
	log.Info("Auth token is %s", token)
	var err error
	var appId int64 = 0
	userId = 0
	roomId = define.NoRoom
	defer func() {
		log.Info("Auth appId is %v, userId is %v, roomId is %v", appId, userId, roomId)
	}()
	if len(token) < 2 {
		return
	}
	token = token[1 : len(token)-1]
	triple := strings.Split(token, "|")
	if len(triple) < 3 {
		return
	}
	if appId, err = strconv.ParseInt(triple[0], 10, 16); err != nil {
		return
	}
	if userId, err = strconv.ParseInt(triple[1], 10, 48); err != nil {
		return
	}
	if roomId, err = strconv.ParseInt(triple[2], 10, 48); err != nil {
		return
	}
	userId = (appId << 48) | userId
	roomId = (appId << 48) | roomId
	return
}
