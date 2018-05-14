package proto

type ConnArg struct {
	Body   []byte
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
	Server    int32
	Key       string
	OldRoomId int64
	RoomId    int64
}

type ChangeRoomReply struct {
	Has bool
}

type UpdateArg struct {
	Server int32
	Key    string
	RoomId int64
}

type RegisterArg struct {
	Server int32
	Info   string
}