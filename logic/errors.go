package main

import (
	"errors"
)

var (
	ErrRouter         = errors.New("router rpc is not available")
	ErrDecodeKey      = errors.New("decode key error")
	ErrNetworkAddr    = errors.New("network addrs error, must network@address")
	ErrConnectArgs    = errors.New("connect rpc args error")
	ErrDisconnectArgs = errors.New("disconnect rpc args error")
	ErrChangeRoomArgs = errors.New("changeroom rpc args error")
	ErrUpdateArgs     = errors.New("update rpc args error")
	ErrRegisterArgs   = errors.New("register rpc args error")
)
