package main

import (
	"goim/libs/define"
	"strconv"
	log "github.com/thinkboy/log4go"
	"strings"
	"git.apache.org/thrift.git/lib/go/thrift"
	"goim/logic/secuserinfo"
	"context"
	"fmt"
)

// developer could implement "Auth" interface for decide how get userId, or roomId
type Auther interface {
	Auth(token string) (userId int64, roomId int64, err error)
}

type DefaultAuther struct {
	client *secuserinfo.SecuserinfoServiceClient
}

func NewDefaultAuther() *DefaultAuther {
	var (
		protocolFactory  thrift.TProtocolFactory
		transportFactory thrift.TTransportFactory
		transport        thrift.TTransport
		err              error
		client           *secuserinfo.SecuserinfoServiceClient
	)
	protocolFactory = thrift.NewTBinaryProtocolFactoryDefault()
	transportFactory = thrift.NewTBufferedTransportFactory(8192)
	transportFactory = thrift.NewTFramedTransportFactory(transportFactory)
	transport, err = thrift.NewTSocket("127.0.0.1:12300")
	if err != nil {
		log.Crashf("Error opening thrift socket: %v", err)
	}
	transport, err = transportFactory.GetTransport(transport)
	if err != nil {
		log.Crashf("Error from transportFactory.GetTransport(): %v", err)
	}
	err = transport.Open()
	if err != nil {
		log.Crashf("Error opening transport: %v", err)
	}
	iprot := protocolFactory.GetProtocol(transport)
	oprot := protocolFactory.GetProtocol(transport)
	client = secuserinfo.NewSecuserinfoServiceClient(thrift.NewTStandardClient(iprot, oprot))
	return &DefaultAuther{
		client: client,
	}
}

func (a *DefaultAuther) Auth(token string) (userId int64, roomId int64, err error) {
	log.Info("Auth token is %s", token)
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
	if len(triple) > 3 {
		ticket := triple[3]
		if err = a.verify(ticket, userId); err != nil {
			return
		}
	}
	userId = (appId << 48) | userId
	roomId = (appId << 48) | roomId
	return
}

func (a *DefaultAuther) verify(ticket string, userId int64) (err error) {
	req := secuserinfo.NewVerifyAppTokenReqEx64()
	req.Context = "nouse"
	req.Yyuid = userId
	req.Token = ticket
	req.Appid = "5060"
	req.EncodingType = 2  // BASE64_WITH_URL = 2,      // 最外层是URLs编码,其次是base64编码
	r, err := a.client.LgSecuserinfoVerifyApptokenEx64(context.TODO(), req)
	if err != nil {
		return
	}
	if r.Rescode != 101 {  // SUI_VERIFY_SUCCESS = 101, // 票据验证成功
		err = fmt.Errorf("got code %d: uid %d verify ticket", r.Rescode, userId)
	}
	return
}