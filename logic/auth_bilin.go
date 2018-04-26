package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	log "github.com/aclisp/log4go"
	pb "github.com/golang/protobuf/proto"
	"goim/libs/define"
	"goim/libs/proto"
	"strconv"
)

const (
	TokenTimestamp = "token-timestamp"
)

type BilinAuther struct {
}

func NewBilinAuther() Auther {
	return &BilinAuther{}
}

func (a *BilinAuther) Auth(body []byte) (userId int64, roomId int64, err error) {
	log.Debug("auth enter. body is \n%s", hex.Dump(body))
	var appId int64 = 0
	userId = 0
	roomId = define.NoRoom
	defer func() {
		log.Debug("auth return. appid=%v, uid=%v, room=%v", appId, userId, roomId)
	}()
	input := proto.RPCInput{}
	if err = pb.Unmarshal(body, &input); err != nil {
		log.Warn("auth body is not a valid protobuf: %v", err)
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
	_, hasToken := input.Opt[define.Token]
	_, hasTime := input.Opt[TokenTimestamp]
	if userId != 0 {
		token := input.Opt[define.Token]
		userid := input.Opt[define.UID]
		timestamp := input.Opt[TokenTimestamp]
		log.Debug("auth verify: %s=%s %s=%s %s=%s", define.Token, token, define.UID, userid, TokenTimestamp, timestamp)
		if hasToken && hasTime {
			err = a.verify(token, userid, timestamp)
		} else {
			err = fmt.Errorf("uid %d must have token and timestamp", userId)
		}
	} else {
		err = fmt.Errorf("uid must not be zero")
	}
	userId = (appId << 48) | userId
	roomId = (appId << 48) | roomId
	return
}

func (a *BilinAuther) verify(token string, userid string, timestamp string) (err error) {
	mac := hmac.New(sha1.New, []byte(Conf.AuthKey))
	mac.Write([]byte(userid + timestamp))
	expected := mac.Sum(nil)
	expectedToken := base64.StdEncoding.EncodeToString(expected)
	if token == expectedToken {
		return
	}
	err = fmt.Errorf("invalid token for uid %s", userid)
	return
}
