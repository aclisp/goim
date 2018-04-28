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
	"time"

	"git.apache.org/thrift.git/lib/go/thrift"
	log "github.com/aclisp/log4go"
	pb "github.com/golang/protobuf/proto"
)

type UdbYyAuther struct {
	udbService thriftpool.Pool
}

func NewUdbYyAuther() Auther {
	udbService, err := thriftpool.NewChannelPool(0, 10000, createUDBServiceConn)
	if err != nil {
		log.Crashf("can not create thrift connection pool for udb auth: %v", err)
	}
	a := &UdbYyAuther{
		udbService: udbService,
	}
	go thriftpool.Ping("udb", a.udbService, func(client interface{}) (err error) {
		_, err = client.(*secuserinfo.SecuserinfoServiceClient).LgSecuserinfoPing(context.TODO(), 0)
		return
	}, 10*time.Minute)
	return a
}

func (a *UdbYyAuther) Auth(body []byte) (userId int64, roomId int64, err error) {
	log.Debug("auth enter. body is \n%s", hex.Dump(body))
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
	if userId != 0 {
		if hasToken {
			err = a.verify(input.Opt[define.Token], userId)
		} else {
			err = fmt.Errorf("uid %d must have a token", userId)
		}
	}
	userId = (appId << 48) | userId
	roomId = (appId << 48) | roomId
	return
}

func (a *UdbYyAuther) verify(ticket string, userId int64) (err error) {
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
	err = thriftpool.Invoke("udb", a.udbService, func(client interface{}) (err error) {
		rsp, err = client.(*secuserinfo.SecuserinfoServiceClient).LgSecuserinfoVerifyApptokenEx64(context.TODO(), req)
		return
	})
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
