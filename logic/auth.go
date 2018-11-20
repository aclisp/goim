package main

import (
	"goim/libs/define"
	"strconv"
	"strings"
)

type AutherBuilder func() Auther

var AutherRegistry = map[string]AutherBuilder{
	"bypass":  NewBypassAuther,
	"bilin":   NewBilinAuther,
	"bilinx":  NewBilinAutherEx,
	//"udbyy":   NewUdbYyAuther,
	"udbintl": NewUdbIntlAuther,
}

// developer could implement "Auth" interface for decide how get userId, or roomId
type Auther interface {
	Auth(body []byte) (userId int64, roomId int64, err error)
}

func authWithString(token string) (userId int64, roomId int64, err error) {
	var appId int64 = 0
	userId = 0
	roomId = define.NoRoom
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

func authBlocked(token, userid string) bool {
	for _, t := range Conf.AuthBlock {
		if t == token {
			return true
		}
	}
	return false
}
