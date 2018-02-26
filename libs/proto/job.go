package proto

// TODO optimize struct after replace kafka
type KafkaMsg struct {
	OP       string   `json:"op"`
	RoomId   int64    `json:"roomid,omitempty"`
	ServerId int32    `json:"server,omitempty"`
	SubKeys  []string `json:"subkeys,omitempty"`
	Msg      []byte   `json:"msg"`
	Ensure   bool     `json:"ensure,omitempty"`
}
