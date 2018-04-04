package define

// 这里定义所有上行消息 RPCInput 的可选项。内建的名称，只允许小写(a-z)和连字符(-)
const (
	// 长连接鉴权时使用
	UID    = "uid"
	AppID  = "appid"
	ConnID = "connid" // 同一uid下connid唯一
	// 长连接鉴权时使用，切换房间时使用
	SubscribeRoom      = "subscribe-room-push"
	IsAnonymousUser    = "is-anonymous-user"
	HeartbeatThreshold = "heartbeat-threshold"
	// 连接相关信息
	ClientIP        = "client-ip"
	ClientPort      = "client-port"
	AccessPointIP   = "access-point-ip"
	AccessPointPort = "access-point-port"
)
