package main

// Start Commond eg: ./client 1 5000 localhost:8080
// first parameter：beginning userId
// second parameter: amount of clients
// third parameter: comet server ip

import (
	"bufio"
	"encoding/binary"
	"flag"
	"fmt"
	mrand "math/rand"
	"encoding/json"
	"os"
	"runtime"
	"strconv"
	"sync/atomic"
	"time"

	"bilin/protocol"
	log "github.com/aclisp/log4go"
	pb "github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	"net"
	"net/url"
)

const (
	OP_HANDSHARE         = int32(0)
	OP_HANDSHARE_REPLY   = int32(1)
	OP_HEARTBEAT         = int32(2)
	OP_HEARTBEAT_REPLY   = int32(3)
	OP_SEND_SMS          = int32(4)
	OP_SEND_SMS_REPLY    = int32(5)
	OP_DISCONNECT_REPLY  = int32(6)
	OP_AUTH              = int32(7)
	OP_AUTH_REPLY        = int32(8)
	OP_ROOM_CHANGE       = int32(15)
	OP_ROOM_CHANGE_REPLY = int32(16)
	OP_TEST              = int32(254)
	OP_TEST_REPLY        = int32(255)
)

const (
	rawHeaderLen = int32(20)
	heart        = 30 * time.Second //s
)

type Proto struct {
	Ver       int16           `json:"ver"`  // protocol version
	Operation int32           `json:"op"`   // operation for request
	SeqId     int32           `json:"seq"`  // sequence number chosen by client
	Body      json.RawMessage `json:"body"` // binary body bytes(json.RawMessage is []byte)
}

func (p *Proto) String() string {
	b, _ := json.Marshal(p)
	return string(b)
}

var (
	countDown int64
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	log.Global = log.NewDefaultLogger(log.DEBUG)
	flag.Parse()
	defer log.Close()
	begin, err := strconv.Atoi(os.Args[1])
	if err != nil {
		panic(err)
	}

	num, err := strconv.Atoi(os.Args[2])
	if err != nil {
		panic(err)
	}

	go result()

	for i := begin; i < begin+num; i++ {
		key := fmt.Sprintf("%d", i)
		//go websocketClient(key)
		go tcpClient(key)
	}

	var exit chan bool
	<-exit
}

func result() {
	var (
		lastTimes int64
		diff      int64
		nowCount  int64
		timer     = int64(30)
	)

	for {
		nowCount = atomic.LoadInt64(&countDown)
		diff = nowCount - lastTimes
		lastTimes = nowCount
		fmt.Println(fmt.Sprintf("%s down:%d down/s:%d", time.Now().Format("2006-01-02 15:04:05"), nowCount, diff/timer))
		time.Sleep(time.Duration(timer) * time.Second)
	}
}

func websocketClient(key string) {
	for {
		startWebsocketClient(key)
		time.Sleep(3 * time.Second)
	}
}

func tcpClient(key string) {
	for {
		startTcpClient(key)
		time.Sleep(3 * time.Second)
	}
}

func startWebsocketClient(key string) {
	//time.Sleep(time.Duration(mrand.Intn(30)) * time.Second)
	quit := make(chan bool, 1)
	defer close(quit)

	u := url.URL{Scheme: "ws", Host: os.Args[3], Path: "/sub"}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Error("net.Dial(\"%s\") error(%v)", os.Args[3], err)
		return
	}
	defer conn.Close()
	seqId := int32(1)
	proto := new(Proto)
	proto.Ver = 1
	// auth
	// test handshake timeout
	// time.Sleep(time.Second * 31)
	proto.Operation = OP_AUTH
	proto.SeqId = seqId
	proto.Body = []byte(key)
	if err = websocketWriteProto(conn, proto); err != nil {
		log.Error("websocketWriteProto() error(%v)", err)
		return
	}
	plist := make([]Proto, 0, 1)
	if err = websocketReadProto(conn, &plist); err != nil {
		log.Error("websocketReadProto() error(%v)", err)
		return
	}
	log.Debug("key:%s websocket auth ok, proto: %v", key, proto)
	seqId++
	// writer
	go func() {
		proto1 := new(Proto)
		for {
			// heartbeat
			proto1.Operation = OP_HEARTBEAT
			proto1.SeqId = seqId
			proto1.Body = nil
			if err = websocketWriteProto(conn, proto1); err != nil {
				log.Error("key:%s websocketWriteProto() error(%v)", key, err)
				return
			}
			log.Debug("key:%s websocket write heartbeat", key)
			// test heartbeat
			time.Sleep(heart)
			seqId++
			select {
			case <-quit:
				return
			default:
			}
		}
	}()
	// reader
	for {
		plist = plist[:0]
		if err = websocketReadProto(conn, &plist); err != nil {
			log.Error("key:%s websocketReadProto() error(%v)", key, err)
			quit <- true
			return
		}
		for _, proto := range plist {
			if proto.Operation == OP_HEARTBEAT_REPLY {
				log.Debug("key:%s websocket receive heartbeat", key)
				if err = conn.SetReadDeadline(time.Now().Add(heart + 60*time.Second)); err != nil {
					log.Error("conn.SetReadDeadline() error(%v)", err)
					quit <- true
					return
				}
				atomic.AddInt64(&countDown, 1)
			} else if proto.Operation == OP_TEST_REPLY {
				log.Debug("websocket body: %s", string(proto.Body))
			} else if proto.Operation == OP_SEND_SMS_REPLY {
				log.Debug("key:%s websocket msg: %s", key, string(proto.Body))
				atomic.AddInt64(&countDown, 1)
			}
		}
	}
}

func startTcpClient(key string) {
	time.Sleep(time.Duration(mrand.Intn(30)) * time.Second)
	quit := make(chan bool, 1)
	defer close(quit)

	// setup tcp conn
	conn, err := net.Dial("tcp", os.Args[3])
	if err != nil {
		log.Error("net.Dial(\"%s\") error(%v)", os.Args[3], err)
		return
	}
	defer conn.Close()
	seqId := int32(1)
	wr := bufio.NewWriter(conn)
	rd := bufio.NewReader(conn)

	// auth user
	authUser := func(uid string) (err error) {
		proto := new(Proto)
		proto.Ver = 1
		// auth
		// test handshake timeout
		// time.Sleep(time.Second * 31)
		proto.Operation = OP_AUTH
		proto.SeqId = seqId
		in := RPCInput{
			Headers: map[string]string{
				"uid": uid,
			},
		}
		proto.Body, err = pb.Marshal(&in)
		if err != nil {
			log.Error("key:%s pb.Marshal(%v) error(%v)", uid, in, err)
			return
		}
		if err = tcpWriteProto(wr, proto); err != nil {
			log.Error("tcpWriteProto() error(%v)", err)
			return
		}
		if err = tcpReadProto(rd, proto); err != nil {
			log.Error("tcpReadProto() error(%v)", err)
			return
		}
		log.Debug("key:%s tcp auth ok, proto: %v", uid, proto)
		seqId++
		return
	}
	if err = authUser(key); err != nil {
		return
	}

	// writer
	go func() {

		heartBeat := func() (err error) {
			proto := new(Proto)
			proto.Operation = OP_HEARTBEAT
			proto.SeqId = seqId
			proto.Body = nil
			if err = tcpWriteProto(wr, proto); err != nil {
				log.Error("key:%s tcpWriteProto() error(%v)", key, err)
				return
			}
			//log.Debug("key:%s tcp write heartbeat", key)
			seqId++
			return
		}

		enterRoom := func(uid, roomid int64) (err error) {
			// inner packet
			req := bilin.EnterBroRoomReq{
				Header: &bilin.Header{
					Userid: uint64(uid),
					Roomid: uint64(roomid),
				},
			}
			reqBuf, err := pb.Marshal(&req)
			if err != nil {
				log.Error("key:%s pb.Marshal(%v) error(%v)", key, req, err)
				return
			}
			// outer packet
			in := RPCInput{
				ServiceName:   "bilin.bcserver2.BCServantObj",
				MethodName:    "EnterBroRoom",
				RequestBuffer: reqBuf,
				Headers: map[string]string{
					"subscribe-room-push": strconv.FormatInt(roomid, 10),
				},
			}
			inBuf, err := pb.Marshal(&in)
			if err != nil {
				log.Error("key:%s pb.Marshal(%v) error(%v)", key, in, err)
				return
			}
			// tcp packet
			proto := new(Proto)
			proto.Operation = OP_ROOM_CHANGE
			proto.SeqId = seqId
			proto.Body = inBuf
			if err = tcpWriteProto(wr, proto); err != nil {
				log.Error("key:%s tcpWriteProto(op=%d) error(%v)", key, proto.Operation, err)
				return
			}
			log.Debug("key:%s tcp write op=%d enter room", key, proto.Operation)
			seqId++
			return
		}

		exitRoom := func(uid, roomid int64) (err error) {
			// inner packet
			req := bilin.ExitBroRoomReq{
				Header: &bilin.Header{
					Userid: uint64(uid),
					Roomid: uint64(roomid),
				},
			}
			reqBuf, err := pb.Marshal(&req)
			if err != nil {
				log.Error("key:%s pb.Marshal(%v) error(%v)", key, req, err)
				return
			}
			// outer packet
			in := RPCInput{
				ServiceName:   "bilin.bcserver2.BCServantObj",
				MethodName:    "ExitBroRoom",
				RequestBuffer: reqBuf,
				Headers: map[string]string{
					"subscribe-room-push": "-1",
				},
			}
			inBuf, err := pb.Marshal(&in)
			if err != nil {
				log.Error("key:%s pb.Marshal(%v) error(%v)", key, in, err)
				return
			}
			// tcp packet
			proto := new(Proto)
			proto.Operation = OP_ROOM_CHANGE
			proto.SeqId = seqId
			proto.Body = inBuf
			if err = tcpWriteProto(wr, proto); err != nil {
				log.Error("key:%s tcpWriteProto(op=%d) error(%v)", key, proto.Operation, err)
				return
			}
			log.Debug("key:%s tcp write op=%d exit room", key, proto.Operation)
			seqId++
			return
		}

		for {
			// send heartbeat
			if err = heartBeat(); err != nil {
				return
			}
			time.Sleep(1 * time.Second)

			// send enter room
			uid, _ := strconv.ParseInt(key, 10, 64)
			roomid := int64(uid / 1000)
			if err = enterRoom(uid, roomid); err != nil {
				return
			}

			// keep user in room
			time.Sleep(heart)

			// send exit room
			if err = exitRoom(uid, roomid); err != nil {
				return
			}
			time.Sleep(10 * time.Second)

			// check quit
			select {
			case <-quit:
				return
			default:
			}
		}
	}()

	// reader
	for {
		proto := new(Proto)
		if err = tcpReadProto(rd, proto); err != nil {
			log.Error("key:%s tcpReadProto() error(%v)", key, err)
			quit <- true
			return
		}
		if proto.Operation == OP_HEARTBEAT_REPLY {
			//log.Debug("key:%s tcp receive heartbeat", key)
			if err = conn.SetReadDeadline(time.Now().Add(heart + 60*time.Second)); err != nil {
				log.Error("conn.SetReadDeadline() error(%v)", err)
				quit <- true
				return
			}
			//atomic.AddInt64(&countDown, 1)
		} else if proto.Operation == OP_TEST_REPLY {
			log.Debug("tcp body: %s", string(proto.Body))
		} else if proto.Operation == OP_SEND_SMS_REPLY && proto.SeqId != 0 {
			// outer packet
			out := RPCOutput{}
			err = pb.Unmarshal(proto.Body, &out)
			if err != nil {
				log.Error("key:%s tcp receive rpc response error(%v)", key, err)
				continue
			}
			if out.RetCode != 0 {
				log.Warn("key:%s tcp receive rpc response code=%d: %s (service=%s method=%s)",
					key, out.RetCode, out.RetDesc, out.ServiceName, out.MethodName)
			}
			// inner packet TODO
			log.Debug("key:%s tcp receive rpc response msg: %+v", key, out)
			atomic.AddInt64(&countDown, 1)
		} else if proto.Operation == OP_SEND_SMS_REPLY {
			push := ServerPush{}
			err = pb.Unmarshal(proto.Body, &push)
			if err != nil {
				log.Error("key:%s tcp receive server push error(%v)", key, err)
				continue
			}
			log.Debug("key:%s tcp receive server push msg: %+v", key, push)
			atomic.AddInt64(&countDown, 1)
		} else if proto.Operation == OP_ROOM_CHANGE_REPLY {
			out := RPCOutput{}
			err = pb.Unmarshal(proto.Body, &out)
			if err != nil {
				log.Error("key:%s tcp receive change room error(%v)", key, err)
				continue
			}
			if out.RetCode != 0 {
				log.Warn("key:%s tcp receive change room code=%d: %s (service=%s method=%s)",
					key, out.RetCode, out.RetDesc, out.ServiceName, out.MethodName)
			}
			log.Debug("key:%s tcp receive change room msg: %+v", key, out)
			atomic.AddInt64(&countDown, 1)
		}
	}
}

func tcpWriteProto(wr *bufio.Writer, proto *Proto) (err error) {
	// write
	if err = binary.Write(wr, binary.BigEndian, rawHeaderLen); err != nil {
		return
	}
	if err = binary.Write(wr, binary.BigEndian, int32(proto.Ver)); err != nil {
		return
	}
	if err = binary.Write(wr, binary.BigEndian, proto.Operation); err != nil {
		return
	}
	if err = binary.Write(wr, binary.BigEndian, proto.SeqId); err != nil {
		return
	}
	if err = binary.Write(wr, binary.BigEndian, int32(len(proto.Body))); err != nil {
		return
	}
	if proto.Body != nil {
		//log.Debug("cipher body: %v", proto.Body)
		if err = binary.Write(wr, binary.BigEndian, proto.Body); err != nil {
			return
		}
	}
	err = wr.Flush()
	return
}

func tcpReadProto(rd *bufio.Reader, proto *Proto) (err error) {
	var (
		bodyLen int32
		headLen int32
		version int32
	)
	// read
	if err = binary.Read(rd, binary.BigEndian, &headLen); err != nil {
		return
	}
	//log.Debug("headLen: %d", headLen)
	if headLen != 20 {
		err = fmt.Errorf("header length must be 20")
		return
	}
	if err = binary.Read(rd, binary.BigEndian, &version); err != nil {
		return
	}
	proto.Ver = int16(version)
	//log.Debug("ver: %d", proto.Ver)
	if err = binary.Read(rd, binary.BigEndian, &proto.Operation); err != nil {
		return
	}
	//log.Debug("operation: %d", proto.Operation)
	if err = binary.Read(rd, binary.BigEndian, &proto.SeqId); err != nil {
		return
	}
	//log.Debug("seqId: %d", proto.SeqId)
	if err = binary.Read(rd, binary.BigEndian, &bodyLen); err != nil {
		return
	}
	var (
		n = int(0)
		t = int(0)
	)
	//log.Debug("read body len: %d", bodyLen)
	if bodyLen > 0 {
		proto.Body = make([]byte, bodyLen)
		for {
			if t, err = rd.Read(proto.Body[n:]); err != nil {
				return
			}
			if n += t; n == int(bodyLen) {
				break
			} else if n < int(bodyLen) {
			} else {
			}
		}
	} else {
		proto.Body = nil
	}
	return
}

func websocketReadProto(conn *websocket.Conn, plist *[]Proto) error {
	err := conn.ReadJSON(plist)
	if err == nil {
		b, _ := json.Marshal(plist)
		log.Debug("websocketReadProto: %s", string(b))
	}
	return err
}

func websocketWriteProto(conn *websocket.Conn, p *Proto) error {
	if p.Body == nil {
		p.Body = []byte("{}")
	}
	log.Debug("websocketWriteProto: %s", p)
	return conn.WriteJSON(p)
}
