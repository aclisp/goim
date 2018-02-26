package proto

type ConnArg struct {
	Token  string
	Server int32
}

type ConnReply struct {
	Key    string
	RoomId int64
}

type DisconnArg struct {
	Key    string
	RoomId int64
}

type DisconnReply struct {
	Has bool
}

type ChangeRoomArg struct {
	Key       string
	OldRoomId int64
	RoomId    int64
}

type ChangeRoomReply struct {
	Has bool
}