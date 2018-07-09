package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	//"encoding/hex"
	"fmt"
	"goim/libs/define"
	"goim/libs/proto"
	"goim/logic/extproto/bilin"
	"strconv"
	"strings"

	"code.yy.com/yytars/goframework/tars/servant"
	log "github.com/aclisp/log4go"
	pb "github.com/golang/protobuf/proto"
)

const (
	TokenTimestamp = "token-timestamp"
)

type BilinAuther struct {
	comm   *servant.Communicator
	client bilin.DubboProxyClient
}

func NewBilinAuther() Auther {
	return &BilinAuther{}
}

func NewBilinAutherEx() Auther {
	var comm = servant.NewPbCommunicator()
	var obj = "bilin.dubboproxy.LongLinkAuthProxy"
	var client = bilin.NewDubboProxyClient(obj, comm)
	return &BilinAuther{
		comm:   comm,
		client: client,
	}
}

func (a *BilinAuther) Auth(body []byte) (userId int64, roomId int64, err error) {
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
	if authBlocked(token, userid) {
		err = fmt.Errorf("token is blocked for uid %s: %s", userid, token)
		return
	}
	items := strings.Split(token, ",")
	if len(items) > 1 && a.comm != nil {
		var ok bool
		accesstoken := items[1]
		token = items[0]
		if ok, err = a.verifyRemotely(accesstoken, userid); err != nil {
			goto Locally
		} else if !ok {
			err = fmt.Errorf("invalid token for uid %s", userid)
			return
		} else {
			return
		}
	}
	if len(items) > 0 {
		token = items[0]
	}
Locally:
	return a.verifyLocally(token, userid, timestamp)
}

func (a *BilinAuther) verifyLocally(token string, userid string, timestamp string) (err error) {
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

func (a *BilinAuther) verifyRemotely(accesstoken string, userid string) (ok bool, err error) {
	rsp, err := a.client.Invoke(context.Background(), &bilin.DPInvokeReq{
		Service: "com.bilin.user.account.service.IUserLoginService",
		Method:  "verifyUserAccessToken",
		Args: []*bilin.DPInvokeArg{
			{
				Type:  "long",
				Value: userid,
			},
			{
				Type:  "java.lang.String",
				Value: accesstoken,
			},
		},
	})
	if err != nil {
		err = fmt.Errorf("error calling dubbo proxy: %v", err)
		log.Error("%v", err)
		return
	}
	if rsp.ThrewException {
		err = fmt.Errorf("error invoking dubbo service: %v", rsp.Type)
		log.Error("%v, %v", err, rsp.Value)
		return
	}
	if rsp.Type == "java.lang.Boolean" && rsp.Value == "true" {
		// success
		ok = true
	} else {
		log.Error("fail to verify access token %q, with user id %q, got %v: %v", accesstoken, userid, rsp.Type, rsp.Value)
		return
	}
	return
}
