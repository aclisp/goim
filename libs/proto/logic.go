package proto

type ConnArg struct {
	Token  string
	Server int32
}

type ConnReply struct {
	Key    string
	RoomId int32
}

type DisconnArg struct {
	Key    string
	RoomId int32
}

type DisconnReply struct {
	Has bool
}

type ChangeRoomArg struct {
	Key       string
	OldRoomId int32
	RoomId    int32
}

type ChangeRoomReply struct {
	Has bool
}