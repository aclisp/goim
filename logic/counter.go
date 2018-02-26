package main

import (
	"goim/libs/net/xrpc"
	"time"
)

const (
	syncCountDelay = 1 * time.Second
)

var (
	RoomCountMap   = make(map[int64]int32) // roomid:count
	ServerCountMap = make(map[int32]int32) // server:count
)

func MergeCount() {
	var (
		c                     *xrpc.Clients
		err                   error
		roomId                int64
		server, count         int32
		roomCounter           map[int64]int32
		roomCount             = make(map[int64]int32)
		serverCounter         map[int32]int32
		serverCount           = make(map[int32]int32)
	)
	// all comet nodes
	for _, c = range routerServiceMap {
		if c != nil {
			if roomCounter, err = allRoomCount(c); err != nil {
				continue
			}
			for roomId, count = range roomCounter {
				roomCount[roomId] += count
			}
			if serverCounter, err = allServerCount(c); err != nil {
				continue
			}
			for server, count = range serverCounter {
				serverCount[server] += count
			}
		}
	}
	RoomCountMap = roomCount
	ServerCountMap = serverCount
}

/*
func RoomCount(roomId int32) (count int32) {
	count = RoomCountMap[roomId]
	return
}
*/

func SyncCount() {
	for {
		MergeCount()
		time.Sleep(syncCountDelay)
	}
}
