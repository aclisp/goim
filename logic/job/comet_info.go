package main

import (
	"time"
)

const (
	syncRoomServersDelay = 1 * time.Second
)

var (
	RoomServersMap = make(map[int64]map[int32]struct{}) // roomid:servers
)

func MergeRoomServers() {
	var (
		c           *Comet
		ok          bool
		roomId      int64
		serverId    int32
		roomIds     map[int64]struct{}
		servers     map[int32]struct{}
		roomServers = make(map[int64]map[int32]struct{})
	)
	// all comet nodes
	for serverId, c = range cometServiceMap {
		if c.rpcClient != nil {
			if roomIds = roomsComet(c.rpcClient); roomIds != nil {
				// merge room's servers
				for roomId, _ = range roomIds {
					if servers, ok = roomServers[roomId]; !ok {
						servers = make(map[int32]struct{})
						roomServers[roomId] = servers
					}
					servers[serverId] = struct{}{}
				}
			}
		}
	}
	RoomServersMap = roomServers
}

func SyncRoomServers() {
	for {
		MergeRoomServers()
		time.Sleep(syncRoomServersDelay)
	}
}
