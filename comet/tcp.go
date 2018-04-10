package main

import (
	"fmt"
	"goim/libs/bufio"
	"goim/libs/bytes"
	"goim/libs/define"
	"goim/libs/proto"
	itime "goim/libs/time"
	"net"
	"strconv"
	"strings"
	"time"

	log "github.com/aclisp/log4go"
)

// InitTCP listen all tcp.bind and start accept connections.
func InitTCP(addrs []string, accept int) (err error) {
	var (
		bind     string
		listener *net.TCPListener
		addr     *net.TCPAddr
	)
	for _, bind = range addrs {
		if addr, err = net.ResolveTCPAddr("tcp4", bind); err != nil {
			log.Error("net.ResolveTCPAddr(\"tcp4\", \"%s\") error(%v)", bind, err)
			return
		}
		if listener, err = net.ListenTCP("tcp4", addr); err != nil {
			log.Error("net.ListenTCP(\"tcp4\", \"%s\") error(%v)", bind, err)
			return
		}
		if Debug {
			log.Debug("start tcp listen: \"%s\"", bind)
		}
		// split N core accept
		for i := 0; i < accept; i++ {
			go acceptTCP(DefaultServer, listener)
		}
	}
	return
}

// Accept accepts connections on the listener and serves requests
// for each incoming connection.  Accept blocks; the caller typically
// invokes it in a go statement.
func acceptTCP(server *Server, lis *net.TCPListener) {
	var (
		conn *net.TCPConn
		err  error
		r    int
	)
	for {
		if conn, err = lis.AcceptTCP(); err != nil {
			// if listener close then return
			log.Error("listener.Accept(\"%s\") error(%v)", lis.Addr().String(), err)
			return
		}
		if err = conn.SetKeepAlive(server.Options.TCPKeepalive); err != nil {
			log.Error("conn.SetKeepAlive() error(%v)", err)
			return
		}
		if err = conn.SetReadBuffer(server.Options.TCPRcvbuf); err != nil {
			log.Error("conn.SetReadBuffer() error(%v)", err)
			return
		}
		if err = conn.SetWriteBuffer(server.Options.TCPSndbuf); err != nil {
			log.Error("conn.SetWriteBuffer() error(%v)", err)
			return
		}
		go serveTCP(server, conn, r)
		if r++; r == maxInt {
			r = 0
		}
	}
}

func serveTCP(server *Server, conn *net.TCPConn, r int) {
	var (
		// timer
		tr = server.round.Timer(r)
		rp = server.round.Reader(r)
		wp = server.round.Writer(r)
		// ip addr
		lAddr = conn.LocalAddr().String()
		rAddr = conn.RemoteAddr().String()
	)
	if Debug {
		log.Debug("%p start tcp serve \"%s\" with \"%s\"", conn, lAddr, rAddr)
	}
	server.serveTCP(conn, rp, wp, tr)
}

// TODO linger close?
func (server *Server) serveTCP(conn *net.TCPConn, rp, wp *bytes.Pool, tr *itime.Timer) {
	var (
		err   error
		key   string
		white bool
		hb    time.Duration // heartbeat
		p     *proto.Proto
		b     *Bucket
		trd   *itime.TimerData
		rb    = rp.Get()
		wb    = wp.Get()
		ch    = NewChannel(server.Options.CliProto, server.Options.SvrProto, define.NoRoom)
		rr    = &ch.Reader
		wr    = &ch.Writer
		opt   = make(map[string]string, 16)
	)
	ch.Reader.ResetBuffer(conn, rb.Bytes())
	ch.Writer.ResetBuffer(conn, wb.Bytes())
	// handshake
	trd = tr.Add(server.Options.HandshakeTimeout, func() {
		conn.Close()
	})
	// must not setadv, only used in auth
	if p, err = ch.CliProto.Set(); err == nil {
		if key, ch.RoomId, hb, err = server.authTCP(rr, wr, p, conn, opt); err == nil {
			b = server.Bucket(key)
			err = b.Put(key, ch)
		}
	}
	if err != nil {
		conn.Close()
		rp.Put(rb)
		wp.Put(wb)
		tr.Del(trd)
		log.Error("%p key: %s handshake failed error(%v)", conn, key, err)
		return
	}
	trd.Key = key
	tr.Set(trd, hb)
	white = DefaultWhitelist.Contains(key)
	if white {
		DefaultWhitelist.Log.Info("key: %s[%d] auth", key, ch.RoomId)
	}
	// hanshake ok start dispatch goroutine
	go server.dispatchTCP(key, conn, wr, wp, wb, ch)
	for {
		if p, err = ch.CliProto.Set(); err != nil {
			break
		}
		if white {
			DefaultWhitelist.Log.Info("key: %s start read proto", key)
		}
		if err = p.ReadTCP(rr); err != nil {
			break
		}
		log.Debug("%p Rx %s tcp serve %+v", conn, key, p)
		if white {
			DefaultWhitelist.Log.Info("key: %s read proto: %v", key, p)
		}
		if p.Operation == define.OP_AUTH {
			p.Body = nil
			p.Operation = define.OP_AUTH_REPLY
			if Debug {
				log.Debug("%p key: %s receive auth", conn, key)
			}
		} else if p.Operation == define.OP_HEARTBEAT {
			tr.Set(trd, hb)
			server.operator.Update(key, ch.RoomId)
			p.Body = nil
			p.Operation = define.OP_HEARTBEAT_REPLY
			if Debug {
				log.Debug("%p key: %s receive heartbeat", conn, key)
			}
		} else if p.Operation == define.OP_ROOM_CHANGE {
			var (
				ret    int
				msg    string
				rid    int64
				appid  int64
				input  proto.RPCInput
				output proto.RPCOutput
			)
			if rid, appid, input, err = tcpParseRoomId(p.Body); err != nil {
				ret = 1
				msg = fmt.Sprintf("invalid roomid: %v", err)
			} else if orid, err := b.Change(key, rid); err != nil {
				ret = 2
				msg = fmt.Sprintf("change roomid %d->%d err: %v", orid, rid, err)
			} else {
				ret = 0
				msg = fmt.Sprintf("change roomid %d->%d ok", orid, rid)
				server.operator.ChangeRoom(key, orid, rid)
				// update conn opt once change roomid ok
				opt[define.AppID] = strconv.FormatInt(appid, 10)
				opt[define.SubscribeRoom] = strconv.FormatInt(int64(int32(rid&0xFFFFFFFFFFFF)), 10)
			}
			if len(input.Obj) > 0 {
				output, err = server.operator.Direct(input, TCPConn, opt)
				if err != nil {
					ret = 3
					msg = fmt.Sprintf("call downstream service %q.%q err: %v", input.Obj, input.Func, err)
				}
			}
			output.Ret = int32(ret)
			output.Desc = msg
			tag := TCPToRPC{}
			p.Body, _ = tag.Encode(output)
			p.Operation = define.OP_ROOM_CHANGE_REPLY
		} else {
			if err = server.operator.Operate(p, TCPConn, opt); err != nil {
				break
			}
		}
		if white {
			DefaultWhitelist.Log.Info("key: %s processed proto: %v", key, p)
		}
		ch.CliProto.SetAdv()
		ch.Signal()
		if white {
			DefaultWhitelist.Log.Info("key: %s signal", key)
		}
	}
	if white {
		DefaultWhitelist.Log.Info("key: %s server tcp error(%v)", key, err)
	}
	if err != nil {
		log.Error("%p key: %s server tcp failed error(%v)", conn, key, err)
	}
	b.Del(key)
	tr.Del(trd)
	rp.Put(rb)
	conn.Close()
	ch.Close()
	if err = server.operator.Disconnect(key, ch.RoomId); err != nil {
		log.Error("%p key: %s operator do disconnect error(%v)", conn, key, err)
	}
	if white {
		DefaultWhitelist.Log.Info("key: %s disconnect error(%v)", key, err)
	}
	if Debug {
		log.Debug("%p key: %s server tcp goroutine exit", conn, key)
	}
	return
}

// dispatch accepts connections on the listener and serves requests
// for each incoming connection.  dispatch blocks; the caller typically
// invokes it in a go statement.
func (server *Server) dispatchTCP(key string, conn *net.TCPConn, wr *bufio.Writer, wp *bytes.Pool, wb *bytes.Buffer, ch *Channel) {
	var (
		err    error
		finish bool
		white  = DefaultWhitelist.Contains(key)
	)
	if Debug {
		log.Debug("%p key: %s start dispatch tcp goroutine", conn, key)
	}
	for {
		if white {
			DefaultWhitelist.Log.Info("key: %s wait proto ready", key)
		}
		var p = ch.Ready()
		if white {
			DefaultWhitelist.Log.Info("key: %s proto ready", key)
		}
		if Debug {
			log.Debug("%p key: %s dispatch msg: {Ver:%d Op:%d Seq:%d Body:%d}", conn, key, p.Ver, p.Operation, p.SeqId, len(p.Body))
		}
		switch p {
		case proto.ProtoFinish:
			if white {
				DefaultWhitelist.Log.Info("key: %s receive proto finish", key)
			}
			if Debug {
				log.Debug("%p key: %s wakeup exit dispatch goroutine", conn, key)
			}
			finish = true
			goto failed
		case proto.ProtoReady:
			// fetch message from svrbox(client send)
			for {
				if p, err = ch.CliProto.Get(); err != nil {
					err = nil // must be empty error
					break
				}
				if white {
					DefaultWhitelist.Log.Info("key: %s start write client proto: %v", key, p)
				}
				if err = p.WriteTCP(wr); err != nil {
					goto failed
				}
				log.Debug("%p Tx %s tcp dispatch %+v", conn, key, p)
				if white {
					DefaultWhitelist.Log.Info("key: %s written client proto: %v", key, p)
				}
				p.Body = nil // avoid memory leak
				ch.CliProto.GetAdv()
			}
		default:
			if white {
				DefaultWhitelist.Log.Info("key: %s start write server proto: %v", key, p)
			}
			// server send
			if err = p.WriteTCP(wr); err != nil {
				goto failed
			}
			log.Debug("%p Tx %s tcp dispatch %+v", conn, key, p)
			if white {
				DefaultWhitelist.Log.Info("key: %s written server proto: %v", key, p)
			}
		}
		if white {
			DefaultWhitelist.Log.Info("key: %s start flush", key)
		}
		// only hungry flush response
		if err = wr.Flush(); err != nil {
			break
		}
		if white {
			DefaultWhitelist.Log.Info("key: %s flush", key)
		}
		if p != nil && p.Operation == define.OP_DISCONNECT_REPLY {
			log.Warn("%p key: %s kicked", conn, key)
			break
		}
	}
failed:
	if white {
		DefaultWhitelist.Log.Info("key: %s dispatch tcp error(%v)", key, err)
	}
	if err != nil {
		log.Error("%p key: %s dispatch tcp error(%v)", conn, key, err)
	}
	conn.Close()
	wp.Put(wb)
	// must ensure all channel message discard, for reader won't blocking Signal
	for !finish {
		finish = (ch.Ready() == proto.ProtoFinish)
	}
	if Debug {
		log.Debug("%p key: %s dispatch goroutine exit", conn, key)
	}
	return
}

// auth for goim handshake with client, use rsa & aes.
func (server *Server) authTCP(rr *bufio.Reader, wr *bufio.Writer, p *proto.Proto, conn *net.TCPConn, opt map[string]string) (key string, rid int64, heartbeat time.Duration, err error) {
	for {
		if err = p.ReadTCP(rr); err != nil {
			return
		}
		log.Debug("%p Rx tcp auth %+v", conn, p)
		if p.Operation == define.OP_HEARTBEAT {
			p.Body = nil
			p.Operation = define.OP_HEARTBEAT_REPLY
			p.WriteTCP(wr)
			wr.Flush()
			log.Debug("%p Tx tcp auth %+v", conn, p)
			continue
		}
		if p.Operation != define.OP_AUTH {
			log.Warn("%p auth operation not valid: %d", conn, p.Operation)
			continue
		}
		break
	}
	key, rid, heartbeat, err = server.operator.Connect(p)
	appid, uid, connid := tcpParseKey(key)
	anony := uid == 0
	var laddr, raddr *net.TCPAddr
	var ok bool
	if laddr, ok = conn.LocalAddr().(*net.TCPAddr); ok {
		opt[define.AccessPointIP] = laddr.IP.String()
		opt[define.AccessPointPort] = strconv.Itoa(laddr.Port)
	}
	if raddr, ok = conn.RemoteAddr().(*net.TCPAddr); ok {
		opt[define.ClientIP] = raddr.IP.String()
		opt[define.ClientPort] = strconv.Itoa(raddr.Port)
	}
	opt[define.IsAnonymousUser] = strconv.FormatBool(anony)
	opt[define.UID] = strconv.FormatInt(uid, 10)
	opt[define.AppID] = strconv.FormatInt(appid, 10)
	opt[define.ConnID] = strconv.FormatInt(connid, 10)
	opt[define.SubscribeRoom] = strconv.FormatInt(int64(int32(rid&0xFFFFFFFFFFFF)), 10)
	opt[define.HeartbeatThreshold] = heartbeat.String()
	p.Operation = define.OP_AUTH_REPLY
	if err != nil {
		output := proto.RPCOutput{
			Ret:  1,
			Desc: "auth fail: " + err.Error(),
			Opt:  opt,
		}
		tag := TCPToRPC{}
		p.Body, _ = tag.Encode(output)
		p.WriteTCP(wr)
		wr.Flush()
		log.Debug("%p Tx tcp auth %+v", conn, p)
		p.Body = nil
		return
	}
	output := proto.RPCOutput{
		Ret:  0,
		Desc: "auth ok",
		Opt:  opt,
	}
	tag := TCPToRPC{}
	p.Body, _ = tag.Encode(output)
	if err = p.WriteTCP(wr); err != nil {
		p.Body = nil
		return
	}
	err = wr.Flush()
	log.Debug("%p Tx %s tcp auth %+v", conn, key, p)
	p.Body = nil
	return
}

func tcpParseRoomId(body []byte) (rid, appid int64, input proto.RPCInput, err error) {
	tag := TCPToRPC{}
	input, err = tag.Decode(body)
	if err != nil {
		err = fmt.Errorf("body is not a valid protobuf: %v", err)
		return
	}
	if _, ok := input.Opt[define.SubscribeRoom]; !ok {
		err = fmt.Errorf("need header %q", define.SubscribeRoom)
		return
	}
	if rid, err = strconv.ParseInt(input.Opt[define.SubscribeRoom], 10, 48); err != nil {
		err = fmt.Errorf("header %q is not integer: %v", define.SubscribeRoom, err)
		return
	}
	if _, ok := input.Opt[define.AppID]; ok {
		if appid, err = strconv.ParseInt(input.Opt[define.AppID], 10, 16); err != nil {
			return
		}
	}
	rid = (appid << 48) | rid
	return
}

func tcpParseKey(key string) (appId, userId, connId int64) {
	var (
		idx int
		err error
	)
	if idx = strings.IndexByte(key, '_'); idx == -1 {
		return
	}
	if userId, err = strconv.ParseInt(key[:idx], 10, 64); err != nil {
		return
	}
	if connId, err = strconv.ParseInt(key[idx+1:], 10, 32); err != nil {
		return
	}
	appId = 0xFFFF & (userId >> 48)
	userId = userId & 0xFFFFFFFFFFFF
	return
}
