package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"goim/libs/define"
	"goim/libs/proto"
	otp "goim/logic/extproto"
	"strconv"

	"code.yy.com/yytars/goframework/tars/servant"
	log "github.com/aclisp/log4go"
	pb "github.com/golang/protobuf/proto"
)

type UdbIntlAuther struct {
	comm   *servant.Communicator
	client otp.OtpVerifyClient
}

func NewUdbIntlAuther() Auther {
	var comm = servant.NewPbCommunicator()
	var obj = "UdbApp.OtpVerifyServer.OtpVerifyObj"
	var client = otp.NewOtpVerifyClient(obj, comm)
	return &UdbIntlAuther{
		comm:   comm,
		client: client,
	}
}

func (a *UdbIntlAuther) Auth(body []byte) (userId int64, roomId int64, err error) {
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

func (a *UdbIntlAuther) verify(ticket string, userId int64) (err error) {
	var (
		req *otp.OtpVerifyReq = &otp.OtpVerifyReq{}
		rsp *otp.OtpVerifyRsp
	)
	req.Otp = ticket
	if rsp, err = a.client.OtpVerify(context.TODO(), req); err != nil {
		log.Error("error calling OtpVerify: %v", err)
		return
	}
	if rsp.Rescode != int32(otp.Errcode_SUCCESS) {
		err = fmt.Errorf("got code %d: uid %d OtpVerify", rsp.Rescode, userId)
		return
	}
	if rsp.Uid != uint64(userId) {
		err = fmt.Errorf("uid mismatch: expect %d, got %d", userId, rsp.Uid)
		return
	}
	return
}
