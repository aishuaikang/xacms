package parse_fsm

type ParseState string
type ParseEvent string

const (
	// 状态定义
	StateOffline ParseState = "Offline" // 离线
	StateOnline  ParseState = "Online"  // 在线
)

const (
	// 事件定义
	EventToOffline ParseEvent = "to_offline" // 切换到离线
	EventToOnline  ParseEvent = "to_online"  // 切换到在线
)
