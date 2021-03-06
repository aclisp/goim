package main

import (
	"context"
	"crypto/tls"
	"goim/libs/define"
	"goim/libs/proto"
	itime "goim/libs/time"
	"math/rand"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/aclisp/log4go"
	"strconv"
	"fmt"
	"strings"
	"errors"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var httpServers []*http.Server

func InitWebsocket(addrs []string) (err error) {
	var (
		bind         string
		listener     *net.TCPListener
		addr         *net.TCPAddr
		httpServeMux = http.NewServeMux()
		server       *http.Server
	)
	httpServeMux.HandleFunc("/sub", ServeWebSocket)

	for _, bind = range addrs {
		if addr, err = net.ResolveTCPAddr("tcp4", bind); err != nil {
			log.Error("net.ResolveTCPAddr(\"tcp4\", \"%s\") error(%v)", bind, err)
			return
		}
		if listener, err = net.ListenTCP("tcp4", addr); err != nil {
			log.Error("net.ListenTCP(\"tcp4\", \"%s\") error(%v)", bind, err)
			return
		}
		server = &http.Server{Handler: httpServeMux}
		log.Info("start websocket listen: \"%s\"", bind)
		go func(host string) {
			if err = server.Serve(listener); err != nil {
				log.Error("server.Serve(\"%s\") error(%v)", host, err)
				if err != http.ErrServerClosed {
					panic(err)
				}
			}
		}(bind)
		httpServers = append(httpServers, server)
	}
	return
}

func InitWebsocketWithTLS(addrs []string, cert, priv string) (err error) {
	var (
		httpServeMux = http.NewServeMux()
	)
	httpServeMux.HandleFunc("/sub", ServeWebSocket)
	config := &tls.Config{}
	config.Certificates = make([]tls.Certificate, 1)
	if config.Certificates[0], err = tls.LoadX509KeyPair(cert, priv); err != nil {
		return
	}
	for _, bind := range addrs {
		server := &http.Server{Addr: bind, Handler: httpServeMux}
		server.SetKeepAlivesEnabled(true)
		log.Info("start websocket wss listen: \"%s\"", bind)
		go func(host string) {
			ln, err := net.Listen("tcp", host)
			if err != nil {
				panic(err)
			}

			tlsListener := tls.NewListener(ln, config)
			if err = server.Serve(tlsListener); err != nil {
				log.Error("server.Serve(\"%s\") error(%v)", host, err)
				if err != http.ErrServerClosed {
					panic(err)
				}
			}
		}(bind)
		httpServers = append(httpServers, server)
	}
	return
}

func ShutdownWebsocket() {
	for _, s := range httpServers {
		s.Shutdown(context.TODO())
	}
}

func ServeWebSocket(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		http.Error(w, "Method Not Allowed", 405)
		return
	}
	ws, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Error("Websocket Upgrade error(%v), userAgent(%s)", err, req.UserAgent())
		return
	}
	defer ws.Close()
	var (
		lAddr = ws.LocalAddr()
		rAddr = ws.RemoteAddr()
		tr    = DefaultServer.round.Timer(rand.Int())
	)
	log.Debug("start websocket serve \"%s\" with \"%s\"", lAddr, rAddr)
	DefaultServer.serveWebsocket(ws, tr)
}

func (server *Server) serveWebsocket(conn *websocket.Conn, tr *itime.Timer) {
	var (
		err error
		key string
		hb  time.Duration // heartbeat
		p   *proto.Proto
		b   *Bucket
		trd *itime.TimerData
		ch  = NewChannel(server.Options.CliProto, server.Options.SvrProto, define.NoRoom)
		opt = make(map[string]string, 16)
	)
	// handshake
	trd = tr.Add(server.Options.HandshakeTimeout, func() {
		conn.Close()
	})
	// must not setadv, only used in auth
	if p, err = ch.CliProto.Set(); err == nil {
		if key, ch.RoomId, hb, err = server.authWebsocket(conn, p); err == nil {
			b = server.Bucket(key)
			err = b.Put(key, ch)
		}
	}
	if err != nil {
		conn.Close()
		tr.Del(trd)
		log.Error("handshake failed error(%v)", err)
		return
	}
	trd.Key = key
	tr.Set(trd, hb)
	// hanshake ok start dispatch goroutine
	go server.dispatchWebsocket(key, conn, ch)
	for {
		if p, err = ch.CliProto.Set(); err != nil {
			break
		}
		if err = p.ReadWebsocket(conn); err != nil {
			break
		}
		//p.Time = *globalNowTime
		if p.Operation == define.OP_HEARTBEAT {
			// heartbeat
			tr.Set(trd, hb)
			server.operator.Update(key, ch.RoomId)
			p.Body = nil
			p.Operation = define.OP_HEARTBEAT_REPLY
		} else if p.Operation == define.OP_ROOM_CHANGE {
			var ret int
			var msg string
			if rid, err := websocketParseRoomId(string(p.Body)); err != nil {
				ret = 1
				msg = fmt.Sprintf("invalid roomid: %s", p.Body)
			} else if orid, err := b.Change(key, rid); err != nil {
				ret = 2
				msg = fmt.Sprintf("change roomid %d->%d err: %v", orid, rid, err)
			} else {
				ret = 0
				msg = fmt.Sprintf("change roomid %d->%d ok", orid, rid)
				tr.Set(trd, hb)
				server.operator.ChangeRoom(key, orid, rid)
			}
			p.Body = []byte(fmt.Sprintf(`{"ret":%d,"msg":%q}`, ret, msg))
			p.Operation = define.OP_ROOM_CHANGE_REPLY
		} else {
			// process message
			if err = server.operator.Operate(p, WebsocketConn, opt); err != nil {
				break
			}
			tr.Set(trd, hb)
			server.operator.Update(key, ch.RoomId)
		}
		ch.CliProto.SetAdv()
		ch.Signal()
	}
	if err != nil {
		if websocket.IsCloseError(err, websocket.CloseGoingAway) {
			log.Debug("key: %s serve websocket failed error(%v)", key, err)
		} else {
			log.Error("key: %s serve websocket failed error(%v)", key, err)
		}
	}
	tr.Del(trd)
	conn.Close()
	ch.Close()
	b.Del(key)
	if err = server.operator.Disconnect(key, ch.RoomId); err != nil {
		log.Error("key: %s operator do disconnect error(%v)", key, err)
	}
	if Debug {
		log.Debug("key: %s serve websocket goroutine exit", key)
	}
	return
}

// dispatch accepts connections on the listener and serves requests
// for each incoming connection.  dispatch blocks; the caller typically
// invokes it in a go statement.
func (server *Server) dispatchWebsocket(key string, conn *websocket.Conn, ch *Channel) {
	var (
		p   *proto.Proto
		err error
	)
	if Debug {
		log.Debug("key: %s start dispatch websocket goroutine", key)
	}
	for {
		p = ch.Ready()
		switch p {
		case proto.ProtoFinish:
			if Debug {
				log.Debug("key: %s wakeup exit dispatch goroutine", key)
			}
			goto failed
		case proto.ProtoReady:
			for {
				if p, err = ch.CliProto.Get(); err != nil {
					err = nil // must be empty error
					break
				}
				if err = p.WriteWebsocket(conn); err != nil {
					goto failed
				}
				p.Body = nil // avoid memory leak
				ch.CliProto.GetAdv()
			}
		default:
			// TODO room-push support
			// just forward the message
			if err = p.WriteWebsocket(conn); err != nil {
				goto failed
			}
		}
		if p != nil && p.Operation == define.OP_DISCONNECT_REPLY {
			log.Warn("key: %s kicked", key)
			break
		}
	}
failed:
	if err != nil {
		log.Error("key: %s dispatch websocket error(%v)", key, err)
	}
	conn.Close()
	// must ensure all channel message discard, for reader won't blocking Signal
	for {
		if p == proto.ProtoFinish {
			break
		}
		p = ch.Ready()
	}
	if Debug {
		log.Debug("key: %s dispatch goroutine exit", key)
	}
	return
}

func (server *Server) authWebsocket(conn *websocket.Conn, p *proto.Proto) (key string, rid int64, heartbeat time.Duration, err error) {
	if err = p.ReadWebsocket(conn); err != nil {
		return
	}
	if p.Operation != define.OP_AUTH {
		err = ErrOperation
		return
	}
	key, rid, heartbeat, err = server.operator.Connect(p)
	p.Operation = define.OP_AUTH_REPLY
	if err != nil {
		p.Body = []byte(fmt.Sprintf(`{"ret":%d,"msg":%q}`, 1, err.Error()))
		p.WriteWebsocket(conn)
		p.Body = nil
		return
	}
	p.Body = okJSONBody
	err = p.WriteWebsocket(conn)
	p.Body = nil
	return
}

func websocketParseRoomId(body string) (rid int64, err error) {
	if len(body) < 2 {
		err = errors.New("body is not json string")
		return
	}
	body = body[1 : len(body)-1]
	pair := strings.Split(body, "|")
	if len(pair) < 2 {
		err = errors.New("expect 'appid|roomid'")
		return
	}
	var appid int64
	if appid, err = strconv.ParseInt(pair[0], 10, 16); err != nil {
		return
	}
	if rid, err = strconv.ParseInt(pair[1], 10, 48); err != nil {
		return
	}
	rid = (appid << 48) | rid
	return
}
