package proto

type PutArg struct {
	UserId int64
	Server int32
	RoomId int64
	Seq    int32
}

type PutReply struct {
	Seq int32
}

type DelArg struct {
	UserId int64
	Seq    int32
	RoomId int64
}

type DelReply struct {
	Has bool
}

type MovArg struct {
	UserId    int64
	Seq       int32
	OldRoomId int64
	RoomId    int64
}

type MovReply struct {
	Has bool
}

type DelServerArg struct {
	Server int32
}

type GetArg struct {
	UserId int64
}

type GetReply struct {
	Seqs    []int32
	Servers []int32
}

type GetAllReply struct {
	UserIds  []int64
	Sessions []*GetReply
}

type MGetArg struct {
	UserIds []int64
}

type MGetReply struct {
	UserIds  []int64
	Sessions []*GetReply
}

type CountReply struct {
	Count int32
}

type RoomCountArg struct {
	RoomId int64
}

type RoomCountReply struct {
	Count int32
}

type AllRoomCountReply struct {
	Counter map[int64]int32
}

type AllUserRoomCountReply struct {
	Counter map[int64]map[int64]int32
}

type AllServerCountReply struct {
	Counter map[int32]int32
}

type UserCountArg struct {
	UserId int64
}

type UserCountReply struct {
	Count int32
}

type UserSessionArg struct {
	UserId int64
}

type UserSession struct {
	UserId  int64
	Count   int32
	Seq     int32
	Servers map[int32]struct {
		Comet     int32
		Birth     string
		Heartbeat string
	}
	Rooms   map[int64]map[int32]int32
}

type UserSessionReply struct {
	*UserSession
}

type GetAllServerReply struct {
	Servers map[int32]struct {
		Info      string
		Birth     string
		Heartbeat string
	}
}