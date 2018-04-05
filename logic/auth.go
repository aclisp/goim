package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"goim/libs/define"
	"goim/libs/proto"
	"goim/libs/thriftpool"
	"goim/logic/secuserinfo"
	"strconv"
	"strings"
	"time"

	"git.apache.org/thrift.git/lib/go/thrift"
	log "github.com/thinkboy/log4go"
	pb "github.com/golang/protobuf/proto"
)

// developer could implement "Auth" interface for decide how get userId, or roomId
type Auther interface {
	Auth(body []byte) (userId int64, roomId int64, err error)
}

type DefaultAuther struct {
	udbService thriftpool.Pool
}

func NewDefaultAuther() *DefaultAuther {
	udbService, err := thriftpool.NewChannelPool(0, 10000, createUDBServiceConn)
	if err != nil {
		log.Crashf("can not create thrift connection pool for udb auth: %v", err)
	}
	a := &DefaultAuther{
		udbService: udbService,
	}
	go thriftpool.Ping("udb", a.udbService, func(client interface{}) (err error){
		_, err = client.(*secuserinfo.SecuserinfoServiceClient).LgSecuserinfoPing(context.TODO(), 0)
		return
	}, 10*time.Minute)
	return a
}

func (a *DefaultAuther) Auth(body []byte) (userId int64, roomId int64, err error) {
	log.Info("Auth enter. body is \n%s", hex.Dump(body))
	var appId int64 = 0
	userId = 0
	roomId = define.NoRoom
	defer func() {
		log.Info("Auth return. appId is %v, userId is %v, roomId is %v", appId, userId, roomId)
	}()
	input := proto.RPCInput{}
	if err =  pb.Unmarshal(body, &input); err != nil {
		log.Warn("Auth body is not a valid protobuf: %v", err)
		return a.authWithString(string(body))
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

func (a *DefaultAuther) authWithString(token string) (userId int64, roomId int64, err error) {
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
	var (
		req *secuserinfo.VerifyAppTokenReqEx64
		rsp *secuserinfo.VerifyAppTokenResEx64
	)
	req = secuserinfo.NewVerifyAppTokenReqEx64()
	req.Context = "nouse"
	req.Yyuid = userId
	req.Token = ticket
	req.Appid = "5060"   // 客户端SDK用这段代码获取票据 String ticket = AuthSDK.getToken("5060");
	req.EncodingType = 2 // BASE64_WITH_URL = 2, // 最外层是URLs编码,其次是base64编码
	err = thriftpool.Invoke("udb", a.udbService, func(client interface{}) (err error){
		rsp, err = client.(*secuserinfo.SecuserinfoServiceClient).LgSecuserinfoVerifyApptokenEx64(context.TODO(), req)
		return
	}, createUDBServiceConn)
	if err != nil {
		log.Error("error calling LgSecuserinfoVerifyApptokenEx64: %v", err)
		return
	}
	if rsp.Rescode != 101 { // SUI_VERIFY_SUCCESS = 101, // 票据验证成功
		err = fmt.Errorf("got code %d: uid %d verify ticket", rsp.Rescode, userId)
		return
	}
	if rsp.Yyuid != req.Yyuid {
		err = fmt.Errorf("uid mismatch: expect %d, got %d", req.Yyuid, rsp.Yyuid)
		return
	}
	return
}

func createUDBServiceConn() (*thriftpool.Conn, error) {
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
		return nil, fmt.Errorf("error new thrift transport: %v", err)
	}
	transport, err = transportFactory.GetTransport(transport)
	if err != nil {
		return nil, fmt.Errorf("error wrap thrift transport: %v", err)
	}
	err = transport.Open()
	if err != nil {
		return nil, fmt.Errorf("error open thrift transport: %v", err)
	}
	iprot := protocolFactory.GetProtocol(transport)
	oprot := protocolFactory.GetProtocol(transport)
	client = secuserinfo.NewSecuserinfoServiceClient(thrift.NewTStandardClient(iprot, oprot))
	return &thriftpool.Conn{
		Socket: transport,
		Client: client,
	}, nil
}
