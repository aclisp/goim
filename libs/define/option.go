package define

// 这里定义所有上行消息 RPCInput 的可选项
const (
	// 长连接鉴权时使用
	UID    = "uid"
	AppID  = "appid"
	ConnID = "connid"
	// 长连接鉴权时使用，切换房间时使用
	SubscribeRoom      = "subscribe-room-push"
	IsAnonymousUser    = "is-anonymous-user"
	HeartbeatThreshold = "heartbeat-threshold"
)
