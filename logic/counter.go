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
	UserRoomCountMap = make(map[int64]map[int64]int32) // roomid->userid count
	ServerCountMap = make(map[int32]int32) // server:count
)

func MergeCount() {
	var (
		c                     *xrpc.Clients
		err                   error
		roomId, userId        int64
		server, count         int32
		roomCounter           map[int64]int32
		roomCount             = make(map[int64]int32)
		userRoomCounter       map[int64]map[int64]int32
		userRoomCount         = make(map[int64]map[int64]int32)
		users, rm             map[int64]int32
		ok                    bool
		serverCounter         map[int32]int32
		serverCount           = make(map[int32]int32)
	)
	// all router nodes
	for _, c = range routerServiceMap {
		if c != nil {
			if roomCounter, err = allRoomCount(c); err != nil {
				continue
			}
			for roomId, count = range roomCounter {
				roomCount[roomId] += count
			}
			if userRoomCounter, err = allUserRoomCount(c); err != nil {
				continue
			}
			for roomId, rm = range userRoomCounter {
				if users, ok = userRoomCount[roomId]; !ok {
					users = make(map[int64]int32)
					userRoomCount[roomId] = users
				}
				for userId, count = range rm {
					users[userId] += count
				}
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
	UserRoomCountMap = userRoomCount
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
