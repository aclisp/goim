package main

import (
	"context"
	"fmt"
	"goim/libs/define"
	"goim/libs/thriftpool"
	"goim/logic/secuserinfo"
	"strconv"
	"strings"
	"time"

	"git.apache.org/thrift.git/lib/go/thrift"
	log "github.com/thinkboy/log4go"
)

// developer could implement "Auth" interface for decide how get userId, or roomId
type Auther interface {
	Auth(token string) (userId int64, roomId int64, err error)
}

type DefaultAuther struct {
	pool thriftpool.Pool
}

func NewDefaultAuther() *DefaultAuther {
	p, err := thriftpool.NewChannelPool(0, 10000, createNewThriftConn)
	if err != nil {
		log.Crashf("can not create thrift connection pool for udb auth: %v", err)
	}
	a := &DefaultAuther{
		pool: p,
	}
	go a.ping()
	return a
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
	conn, err := a.pool.Get()
	if err != nil {
		return
	}
	defer conn.Close()

	client := conn.Client.(*secuserinfo.SecuserinfoServiceClient)
	req := secuserinfo.NewVerifyAppTokenReqEx64()
	req.Context = "nouse"
	req.Yyuid = userId
	req.Token = ticket
	req.Appid = "5060"   // 客户端SDK用这段代码获取票据 String ticket = AuthSDK.getToken("5060");
	req.EncodingType = 2 // BASE64_WITH_URL = 2, // 最外层是URLs编码,其次是base64编码
	r, err := client.LgSecuserinfoVerifyApptokenEx64(context.TODO(), req)
	if err != nil {
		// close the socket that failed
		conn.Conn.Close()
		// reconnect the socket
		if conn.Conn, err = createNewThriftConn(); err != nil {
			conn.MarkUnusable()
			return
		}
		// retry on the newly connected socket
		client = conn.Client.(*secuserinfo.SecuserinfoServiceClient)
		r, err = client.LgSecuserinfoVerifyApptokenEx64(context.TODO(), req)
		if err != nil {
			conn.MarkUnusable()
			return
		}
	}
	if r.Rescode != 101 { // SUI_VERIFY_SUCCESS = 101, // 票据验证成功
		err = fmt.Errorf("got code %d: uid %d verify ticket", r.Rescode, userId)
		return
	}
	if r.Yyuid != req.Yyuid {
		err = fmt.Errorf("uid mismatch: expect %d, got %d", req.Yyuid, r.Yyuid)
		return
	}
	return
}

func createNewThriftConn() (*thriftpool.Conn, error) {
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

func (a *DefaultAuther) ping() {
	for {
		n := a.pool.Len()
		for i := 0; i < n; i++ {
			conn, err := a.pool.Get()
			if err != nil {
				break
			}
			client := conn.Client.(*secuserinfo.SecuserinfoServiceClient)
			_, err = client.LgSecuserinfoPing(context.TODO(), 0)
			if err != nil {
				conn.MarkUnusable()
			}
			conn.Close()
		}
		time.Sleep(time.Minute * 10)
	}
}
