package main

import (
	"goim/libs/bytes"
	"goim/libs/define"
	"goim/libs/proto"
	itime "goim/libs/time"
	"sync"
	"time"

	log "github.com/aclisp/log4go"
)

const (
	roomMapCup = 100
	roomIdle   = 53 * time.Minute
)

var roomBucket *RoomBucket

type RoomBucket struct {
	roomNum int
	rooms   map[int64]*Room
	rwLock  sync.RWMutex
	options RoomOptions
	round   *Round
}

func InitRoomBucket(r *Round, options RoomOptions) {
	roomBucket = &RoomBucket{
		roomNum: 0,
		rooms:   make(map[int64]*Room, roomMapCup),
		rwLock:  sync.RWMutex{},
		options: options,
		round:   r,
	}
}

func (b *RoomBucket) Get(roomId int64) (r *Room) {
	b.rwLock.Lock()
	room, ok := b.rooms[roomId]
	if !ok {
		room = NewRoom(roomId, b.round.Timer(b.roomNum), b.options)
		b.rooms[roomId] = room
		b.roomNum++
		log.Debug("new room:%d, num:%d", roomId, b.roomNum)
	}
	b.rwLock.Unlock()
	return room
}

func (b *RoomBucket) Del(roomId int64) {
	b.rwLock.Lock()
	delete(b.rooms, roomId)
	b.rwLock.Unlock()
}

func (b *RoomBucket) Size() int {
	b.rwLock.RLock()
	n := len(b.rooms)
	b.rwLock.RUnlock()
	return n
}

type RoomOptions struct {
	BatchNum   int
	SignalTime time.Duration
}

type Room struct {
	id    int64
	proto chan *proto.Proto
}

var (
	roomReadyProto = &proto.Proto{Operation: define.OP_ROOM_READY}
)

// NewRoom new a room struct, store channel room info.
func NewRoom(id int64, t *itime.Timer, options RoomOptions) (r *Room) {
	r = new(Room)
	r.id = id
	r.proto = make(chan *proto.Proto, options.BatchNum*2)
	go r.pushproc(t, options.BatchNum, options.SignalTime)
	return
}

// Push push msg to the room, if chan full discard it.
func (r *Room) Push(ver int16, operation int32, msg []byte) (err error) {
	var p = &proto.Proto{Ver: ver, Operation: operation, Body: msg}
	select {
	case r.proto <- p:
	default:
		err = ErrRoomFull
	}
	return
}

// EPush ensure push msg to the room.
func (r *Room) EPush(ver int16, operation int32, msg []byte) {
	var p = &proto.Proto{Ver: ver, Operation: operation, Body: msg}
	r.proto <- p
	return
}

// pushproc merge proto and push msgs in batch.
func (r *Room) pushproc(timer *itime.Timer, batch int, sigTime time.Duration) {
	var (
		n   int
		p   *proto.Proto
		td  *itime.TimerData
		buf = bytes.NewWriterSize(int(proto.MaxBodySize))
	)
	log.Debug("start room:%d goroutine, total:%d", r.id, roomBucket.Size())
	td = timer.Add(roomIdle, func() {
		select {
		case r.proto <- roomReadyProto:
		default:
		}
	})
	for {
		if p = <-r.proto; p != roomReadyProto {
			// merge buffer ignore error, always nil
			p.WriteTo(buf)
			// batch
			if n++; n == 1 {
				timer.Set(td, sigTime)
				continue
			} else if n < batch {
				continue
			}
		} else if n == 0 {
			// idle
			// before quit, check again if there is another push
			select {
			case p = <-r.proto:
				r.proto <- p
				continue
			default:
			}
			break
		}
		broadcastRoomBytes(r.id, buf.Buffer())
		n = 0
		timer.Set(td, roomIdle)
		// TODO use reset buffer
		// after push to room channel, renew a buffer, let old buffer gc
		buf = bytes.NewWriterSize(buf.Size())
	}
	timer.Del(td)
	roomBucket.Del(r.id)
	log.Debug("end room:%d goroutine exit, total:%d", r.id, roomBucket.Size())
}
