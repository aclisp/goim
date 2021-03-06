package main

import (
	"goim/libs/define"
	"goim/libs/proto"
	"sync"
	"time"
)

type Bucket struct {
	bLock             sync.RWMutex
	server            int                       // session server map init num
	session           int                       // bucket session init num
	sessions          map[int64]*Session        // userid->sessions
	roomCounter       map[int64]int32           // roomid->count
	serverCounter     map[int32]int32           // server->count
	userServerCounter map[int32]map[int64]int32 // serverid->userid count
	userRoomCounter   map[int64]map[int64]int32 // roomid->userid count
	cleaner           *Cleaner                  // bucket map cleaner
}

// NewBucket new a bucket struct. store the subkey with im channel.
func NewBucket(session, server, cleaner int) *Bucket {
	b := new(Bucket)
	b.sessions = make(map[int64]*Session, session)
	b.roomCounter = make(map[int64]int32)
	b.serverCounter = make(map[int32]int32)
	b.userServerCounter = make(map[int32]map[int64]int32)
	b.userRoomCounter = make(map[int64]map[int64]int32)
	b.cleaner = NewCleaner(cleaner)
	b.server = server
	b.session = session
	go b.clean()
	return b
}

// counter incr or decr counter.
func (b *Bucket) counter(userId int64, server int32, roomId int64, incr bool) {
	var (
		sm map[int64]int32 // userid->count
		rm map[int64]int32 // userid->count
		v  int32
		ok bool
	)
	if sm, ok = b.userServerCounter[server]; !ok {
		sm = make(map[int64]int32, b.session)
		b.userServerCounter[server] = sm
	}
	if rm, ok = b.userRoomCounter[roomId]; !ok {
		rm = make(map[int64]int32, b.session)
		b.userRoomCounter[roomId] = rm
	}
	if incr {
		sm[userId]++
		rm[userId]++
		b.roomCounter[roomId]++
		b.serverCounter[server]++
	} else {
		// WARN:
		// if decr a userid but key not exists just ignore
		// this may not happen
		if v, _ = sm[userId]; v <= 1 {
			delete(sm, userId)
			if len(sm) == 0 {
				delete(b.userServerCounter, server)
			}
		} else {
			sm[userId] = v - 1
		}
		if v, _ = rm[userId]; v <= 1 {
			delete(rm, userId)
			if len(rm) == 0 {
				delete(b.userRoomCounter, roomId)
			}
		} else {
			rm[userId] = v - 1
		}
		if v, _ = b.roomCounter[roomId]; v <= 1 {
			delete(b.roomCounter, roomId)
		} else {
			b.roomCounter[roomId] = v - 1
		}
		if v, _ = b.serverCounter[server]; v <= 1 {
			delete(b.serverCounter, server)
		} else {
			b.serverCounter[server] = v - 1
		}
	}
}

func (b *Bucket) counterRoom(userId int64, server int32, oldRoomId, roomId int64) {
	var (
		old map[int64]int32
		now map[int64]int32
		v   int32
		ok  bool
	)
	if oldRoomId == roomId {
		return
	}
	if v, _ = b.roomCounter[oldRoomId]; v <= 1 {
		delete(b.roomCounter, oldRoomId)
	} else {
		b.roomCounter[oldRoomId] = v - 1
	}
	b.roomCounter[roomId]++
	if old, ok = b.userRoomCounter[oldRoomId]; !ok {
		old = make(map[int64]int32, b.session)
		b.userRoomCounter[oldRoomId] = old
	}
	if now, ok = b.userRoomCounter[roomId]; !ok {
		now = make(map[int64]int32, b.session)
		b.userRoomCounter[roomId] = now
	}
	if v, _ = old[userId]; v <= 1 {
		delete(old, userId)
		if len(old) == 0 {
			delete(b.userRoomCounter, oldRoomId)
		}
	} else {
		old[userId] = v - 1
	}
	now[userId]++
}

// Put put a channel according with user id. update if seqLast is not zero.
func (b *Bucket) Put(userId int64, server int32, roomId int64, seqLast int32) (seq int32) {
	var (
		s       *Session
		ok, has bool
	)
	b.bLock.Lock()
	if s, ok = b.sessions[userId]; !ok {
		s = NewSession(b.server)
		b.sessions[userId] = s
	}
	if roomId != define.NoRoom {
		seq, has = s.PutRoom(server, roomId, seqLast)
	} else {
		seq, has = s.Put(server, seqLast)
	}
	if seqLast == 0 || !ok || !has {
		b.counter(userId, server, roomId, true)
	}
	b.bLock.Unlock()
	return
}

func (b *Bucket) Get(userId int64) (seqs []int32, servers []int32) {
	b.bLock.RLock()
	if s, ok := b.sessions[userId]; ok {
		seqs, servers = s.Servers()
	}
	b.bLock.RUnlock()
	return
}

func (b *Bucket) GetAll() (userIds []int64, seqs [][]int32, servers [][]int32) {
	b.bLock.RLock()
	i := len(b.sessions)
	userIds = make([]int64, i)
	seqs = make([][]int32, i)
	servers = make([][]int32, i)
	for userId, s := range b.sessions {
		i--
		userIds[i] = userId
		seqs[i], servers[i] = s.Servers()
	}
	b.bLock.RUnlock()
	return
}

func (b *Bucket) Tidy() {
	now := uint32(time.Now().Unix())
	type deadSession struct {
		userId int64
		seqs   []int32
		rooms  []int64
	}
	var deadList []deadSession
	b.bLock.RLock()
	for userId, s := range b.sessions {
		seqs, rooms := s.Dead(now)
		deadList = append(deadList, deadSession{
			userId: userId,
			seqs:   seqs,
			rooms:  rooms,
		})
	}
	b.bLock.RUnlock()
	for _, v := range deadList {
		for i := range v.seqs {
			b.Del(v.userId, v.seqs[i], v.rooms[i])
		}
	}
}

// Del delete the channel by sub key.
func (b *Bucket) Del(userId int64, seq int32, roomId int64) (ok bool) {
	var (
		s          *Session
		server     int32
		has, empty bool
	)
	b.bLock.Lock()
	if s, ok = b.sessions[userId]; ok {
		// WARN:
		// delete(b.sessions, userId)
		// empty is a dirty data, we use here for try lru clean discard session.
		// when one user flapped connect & disconnect, this also can reduce
		// frequently new & free object, gc is slow!!!
		if roomId != define.NoRoom {
			has, empty, server = s.DelRoom(seq, roomId)
		} else {
			has, empty, server = s.Del(seq)
		}
		if has {
			b.counter(userId, server, roomId, false)
		}
	}
	b.bLock.Unlock()
	// lru
	if empty {
		b.cleaner.PushFront(userId, Conf.SessionExpire)
	}
	return
}

// Mov moves the channel from oldRoomId to roomId
func (b *Bucket) Mov(userId int64, seq int32, oldRoomId int64, roomId int64) (ok bool) {
	var (
		s      *Session
		server int32
		has    bool
	)
	b.bLock.Lock()
	if s, ok = b.sessions[userId]; ok {
		has, server = s.MovRoom(seq, oldRoomId, roomId)
		if has {
			b.counterRoom(userId, server, oldRoomId, roomId)
		}
	}
	b.bLock.Unlock()
	return
}

func (b *Bucket) count(roomId int64) (count int32) {
	b.bLock.RLock()
	count = b.roomCounter[roomId]
	b.bLock.RUnlock()
	return
}

func (b *Bucket) Count() (count int32) {
	count = b.count(define.NoRoom)
	return
}

func (b *Bucket) RoomCount(roomId int64) (count int32) {
	count = b.count(roomId)
	return
}

func (b *Bucket) AllRoomCount() (roomCounter map[int64]int32) {
	var roomId int64
	var count int32
	b.bLock.RLock()
	roomCounter = make(map[int64]int32)
	for roomId, count = range b.roomCounter {
		if count > 0 {
			roomCounter[roomId] = count
		}
	}
	b.bLock.RUnlock()
	return
}

func (b *Bucket) AllUserRoomCount() (userRoomCounter map[int64]map[int64]int32) {
	b.bLock.RLock()
	userRoomCounter = make(map[int64]map[int64]int32)
	for roomId, rm := range b.userRoomCounter {
		if len(rm) == 0 {
			continue
		}
		userCounter := make(map[int64]int32, len(rm))
		for userId, count := range rm {
			if count > 0 {
				userCounter[userId] = count
			}
		}
		if len(userCounter) == 0 {
			continue
		}
		userRoomCounter[roomId] = userCounter
	}
	b.bLock.RUnlock()
	return
}

func (b *Bucket) AllServerCount() (serverCounter map[int32]int32) {
	var server, count int32
	b.bLock.RLock()
	serverCounter = make(map[int32]int32, len(b.serverCounter))
	for server, count = range b.serverCounter {
		serverCounter[server] = count
	}
	b.bLock.RUnlock()
	return
}

func (b *Bucket) UserCount(userId int64) (count int32) {
	b.bLock.RLock()
	if s, ok := b.sessions[userId]; ok {
		count = int32(s.Count())
	}
	b.bLock.RUnlock()
	return
}

func (b *Bucket) UserSession(userId int64) (us *proto.UserSession) {
	b.bLock.RLock()
	if s, ok := b.sessions[userId]; ok {
		servers := make(map[int32]struct {
			Comet     int32
			Birth     string
			Heartbeat string
		}, len(s.servers))
		for seq, comet := range s.servers {
			v := servers[seq]
			v.Comet = comet.id
			v.Birth = time.Unix(int64(comet.birth), 0).Format(time.Stamp)
			v.Heartbeat = time.Unix(int64(comet.heartbeat), 0).Format(time.Stamp)
			servers[seq] = v
		}
		rooms := make(map[int64]map[int32]int32, len(s.rooms))
		for roomid, m := range s.rooms {
			v := make(map[int32]int32, len(m))
			for seq, server := range m {
				v[seq] = server
			}
			rooms[roomid] = v
		}
		us = &proto.UserSession{
			UserId:  userId,
			Count:   int32(s.Count()),
			Seq:     s.seq,
			Servers: servers,
			Rooms:   rooms,
		}
	}
	b.bLock.RUnlock()
	return
}

func (b *Bucket) delEmpty(userId int64) {
	var (
		s  *Session
		ok bool
	)
	if s, ok = b.sessions[userId]; ok {
		if s.Count() == 0 {
			delete(b.sessions, userId)
		}
	}
}

func (b *Bucket) clean() {
	var (
		i       int
		userIds []int64
	)
	for {
		userIds = b.cleaner.Clean()
		if len(userIds) != 0 {
			b.bLock.Lock()
			for i = 0; i < len(userIds); i++ {
				b.delEmpty(userIds[i])
			}
			b.bLock.Unlock()
			continue
		}
		time.Sleep(Conf.BucketCleanPeriod)
		b.Tidy()
	}
}
