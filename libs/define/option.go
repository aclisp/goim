package define

// 这里定义所有上行消息 RPCInput 的可选项
const (
	// 长连接鉴权时使用
	UID = "uid"
	AppID = "appid"
	// 长连接鉴权时使用，切换房间时使用
	SubscribeRoom = "subscribe-room"
)