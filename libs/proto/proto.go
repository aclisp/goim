package proto

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"goim/libs/bufio"
	"goim/libs/bytes"
	"goim/libs/define"
	"goim/libs/encoding/binary"

	"github.com/gorilla/websocket"
)

// for tcp
const (
	MaxBodySize = int32(1024*8 - 20)
)

// https://github.com/Tencent/mars/blob/master/mars/stn/proto/longlink_packer.cc
/*
#pragma pack(push, 1)
struct __STNetMsgXpHeader {
    uint32_t    head_length;
    uint32_t    client_version;
    uint32_t    cmdid;
    uint32_t    seq;
    uint32_t	body_length;
};
#pragma pack(pop)
*/
const (
	// size
	RawHeaderSize = 20
	MaxPackSize   = MaxBodySize + int32(RawHeaderSize)
	// offset
	HeadLengthOffset    = 0
	ClientVersionOffset = HeadLengthOffset + 4
	CmdIdOffset         = ClientVersionOffset + 4
	SeqOffset           = CmdIdOffset + 4
	BodyLengthOffset    = SeqOffset + 4
	BodyOffset          = BodyLengthOffset + 4
)

var (
	emptyProto    = Proto{}
	emptyJSONBody = []byte("{}")

	ErrProtoPackLen   = errors.New("default server codec pack length error")
	ErrProtoHeaderLen = errors.New("default server codec header length error")
)

var (
	ProtoReady  = &Proto{Operation: define.OP_PROTO_READY}
	ProtoFinish = &Proto{Operation: define.OP_PROTO_FINISH}
)

// Proto is a request&response written before every goim connect.  It is used internally
// but documented here as an aid to debugging, such as when analyzing
// network traffic.
// tcp:
// binary codec
// websocket & http:
// raw codec, with http header stored ver, operation, seqid
type Proto struct {
	Ver       int16           `json:"ver"`  // protocol version
	Operation int32           `json:"op"`   // operation for request
	SeqId     int32           `json:"seq"`  // sequence number chosen by client
	Body      json.RawMessage `json:"body"` // binary body bytes(json.RawMessage is []byte)
}

func (p *Proto) Reset() {
	*p = emptyProto
}

func (p *Proto) String() string {
	return fmt.Sprintf("\n-------- proto --------\nver: %d\nop: %d\nseq: %d\nbody: \n%s-----------------------", p.Ver, p.Operation, p.SeqId, hex.Dump(p.Body))
}

func (p *Proto) WriteTo(b *bytes.Writer) {
	var (
		buf = b.Peek(RawHeaderSize)
	)
	binary.BigEndian.PutInt32(buf[HeadLengthOffset:], RawHeaderSize)
	binary.BigEndian.PutInt32(buf[ClientVersionOffset:], int32(p.Ver))
	binary.BigEndian.PutInt32(buf[CmdIdOffset:], p.Operation)
	binary.BigEndian.PutInt32(buf[SeqOffset:], p.SeqId)
	binary.BigEndian.PutInt32(buf[BodyLengthOffset:], int32(len(p.Body)))
	if p.Body != nil {
		b.Write(p.Body)
	}
}

func (p *Proto) ReadTCP(rr *bufio.Reader) (err error) {
	var (
		bodyLen int32
		headLen int32
		packLen int32
		buf     []byte
	)
	if buf, err = rr.Pop(RawHeaderSize); err != nil {
		return
	}
	headLen = binary.BigEndian.Int32(buf[HeadLengthOffset:ClientVersionOffset])
	bodyLen = binary.BigEndian.Int32(buf[BodyLengthOffset:BodyOffset])
	packLen = headLen + bodyLen

	p.Ver = int16(binary.BigEndian.Int32(buf[ClientVersionOffset:CmdIdOffset]))
	p.Operation = binary.BigEndian.Int32(buf[CmdIdOffset:SeqOffset])
	p.SeqId = binary.BigEndian.Int32(buf[SeqOffset:BodyLengthOffset])
	if packLen > MaxPackSize {
		return ErrProtoPackLen
	}
	if headLen != RawHeaderSize {
		return ErrProtoHeaderLen
	}
	if bodyLen > 0 {
		p.Body, err = rr.Pop(int(bodyLen))
	} else {
		p.Body = nil
	}
	return
}

func (p *Proto) WriteTCP(wr *bufio.Writer) (err error) {
	var (
		buf []byte
	)
	if p.Operation == define.OP_RAW {
		// write without buffer, job concact proto into raw buffer
		_, err = wr.WriteRaw(p.Body)
		return
	}
	if buf, err = wr.Peek(RawHeaderSize); err != nil {
		return
	}
	binary.BigEndian.PutInt32(buf[HeadLengthOffset:], RawHeaderSize)
	binary.BigEndian.PutInt32(buf[ClientVersionOffset:], int32(p.Ver))
	binary.BigEndian.PutInt32(buf[CmdIdOffset:], p.Operation)
	binary.BigEndian.PutInt32(buf[SeqOffset:], p.SeqId)
	binary.BigEndian.PutInt32(buf[BodyLengthOffset:], int32(len(p.Body)))
	if p.Body != nil {
		_, err = wr.Write(p.Body)
	}
	return
}

func (p *Proto) ReadWebsocket(wr *websocket.Conn) (err error) {
	if p.Body != nil {
		panic("memory pointed by p.Body may be overwritten after ReadWebsocket")
	}
	err = wr.ReadJSON(p)
	return
}

func (p *Proto) WriteBodyTo(b *bytes.Writer) (err error) {
	var (
		ph  Proto
		js  []json.RawMessage
		j   json.RawMessage
		jb  []byte
		bts []byte
	)
	offset := int32(0)
	buf := p.Body[:]
	for {
		if (len(buf[offset:])) < RawHeaderSize {
			// should not be here
			break
		}
		headLen := binary.BigEndian.Int32(buf[offset+HeadLengthOffset : offset+ClientVersionOffset])
		bodyLen := binary.BigEndian.Int32(buf[offset+BodyLengthOffset : offset+BodyOffset])
		packLen := headLen + bodyLen
		packBuf := buf[offset : offset+packLen]
		// packet
		ph.Ver = int16(binary.BigEndian.Int32(packBuf[ClientVersionOffset:CmdIdOffset]))
		ph.Operation = binary.BigEndian.Int32(packBuf[CmdIdOffset:SeqOffset])
		ph.SeqId = binary.BigEndian.Int32(packBuf[SeqOffset:BodyLengthOffset])
		ph.Body = packBuf[BodyOffset:]
		if jb, err = json.Marshal(&ph); err != nil {
			return
		}
		j = json.RawMessage(jb)
		js = append(js, j)
		offset += packLen
	}
	if bts, err = json.Marshal(js); err != nil {
		return
	}
	b.Write(bts)
	return
}

func (p *Proto) WriteWebsocket(wr *websocket.Conn) (err error) {
	if p.Body == nil {
		p.Body = emptyJSONBody
	}
	// [{"ver":1,"op":8,"seq":1,"body":{}}, {"ver":1,"op":3,"seq":2,"body":{}}]
	if p.Operation == define.OP_RAW {
		// batch mod
		var b = bytes.NewWriterSize(len(p.Body) + 40*RawHeaderSize)
		if err = p.WriteBodyTo(b); err != nil {
			return
		}
		err = wr.WriteMessage(websocket.TextMessage, b.Buffer())
		return
	}
	err = wr.WriteJSON([]*Proto{p})
	return
}
