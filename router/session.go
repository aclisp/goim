package main

import (
	"goim/libs/define"
	"time"
)

type comet struct {
	id        int32
	birth     uint32
	heartbeat uint32
}

type Session struct {
	seq     int32
	servers map[int32]comet           // seq:server
	rooms   map[int64]map[int32]int32 // roomid:seq:server with specified room id
}

// NewSession new a session struct. store the seq and serverid.
func NewSession(server int) *Session {
	s := new(Session)
	s.servers = make(map[int32]comet, server)
	s.rooms = make(map[int64]map[int32]int32)
	s.seq = 0
	return s
}

func (s *Session) nextSeq() int32 {
	s.seq++
	return s.seq
}

// Put put a session according with sub key.
func (s *Session) Put(server int32, seqLast int32) (seq int32) {
	if seqLast == 0 {
		seq = s.nextSeq()
	} else {
		seq = seqLast
	}
	now := uint32(time.Now().Unix())
	if v, has := s.servers[seq]; has {
		s.servers[seq] = comet{
			id:        server,
			birth:     v.birth,
			heartbeat: now,
		}
	} else {
		s.servers[seq] = comet{
			id:        server,
			birth:     now,
			heartbeat: now,
		}
	}
	return
}

// PutRoom put a session in a room according with subkey.
func (s *Session) PutRoom(server int32, roomId int64, seqLast int32) (seq int32) {
	var (
		ok   bool
		room map[int32]int32
	)
	seq = s.Put(server, seqLast)
	if room, ok = s.rooms[roomId]; !ok {
		room = make(map[int32]int32)
		s.rooms[roomId] = room
	}
	room[seq] = server
	return
}

func (s *Session) Servers() (seqs []int32, servers []int32) {
	var (
		i           = len(s.servers)
	)
	seqs = make([]int32, i)
	servers = make([]int32, i)
	for seq, comet := range s.servers {
		i--
		seqs[i] = seq
		servers[i] = comet.id
	}
	return
}

// Del delete the session by sub key.
func (s *Session) Del(seq int32) (has, empty bool, server int32) {
	var v comet
	if v, has = s.servers[seq]; has {
		delete(s.servers, seq)
		server = v.id
	}
	empty = (len(s.servers) == 0)
	return
}

// DelRoom delete the session and room by subkey.
func (s *Session) DelRoom(seq int32, roomId int64) (has, empty bool, server int32) {
	var (
		ok   bool
		room map[int32]int32
	)
	has, empty, server = s.Del(seq)
	if room, ok = s.rooms[roomId]; ok {
		delete(room, seq)
		if len(room) == 0 {
			delete(s.rooms, roomId)
		}
	}
	return
}

// MovRoom keep the session, but move room from old to new.
func (s *Session) MovRoom(seq int32, oldRoomId int64, roomId int64) (has bool, server int32) {
	var (
		ok   bool
		room map[int32]int32
		v comet
	)
	if v, has = s.servers[seq]; has {
		server = v.id
	}
	if oldRoomId == roomId {
		return
	}
	if oldRoomId != define.NoRoom {
		if room, ok = s.rooms[oldRoomId]; ok {
			delete(room, seq)
			if len(room) == 0 {
				delete(s.rooms, oldRoomId)
			}
		}
	}
	if has && roomId != define.NoRoom {
		if room, ok = s.rooms[roomId]; !ok {
			room = make(map[int32]int32)
			s.rooms[roomId] = room
		}
		room[seq] = server
	}
	return
}

func (s *Session) Count() int {
	return len(s.servers)
}
