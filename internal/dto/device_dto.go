package dto

// CreateDeviceRequest 创建设备请求结构
type CreateDeviceRequest struct {
	Name      string  `json:"name" validate:"required,min=2,max=64"`
	Longitude float64 `json:"longitude" validate:"required,longitude"`
	Latitude  float64 `json:"latitude" validate:"required,latitude"`

	// 侦测模块
	DetectionID   int    `json:"detection_id" validate:"required"`   // 侦测模块ID
	DetectionIP   string `json:"detection_ip" validate:"required"`   // 侦测模块IP
	DetectionPort int    `json:"detection_port" validate:"required"` // 侦测模块端口

	// 解析模块
	ParseID int    `json:"parse_id" validate:"required"` // 解析模块ID
	ParseIP string `json:"parse_ip" validate:"required"` // 解析模块IP

	// FPV模块
	FPVIP          string `json:"fpv_ip" validate:"required"`           // FPV模块IP
	StreamServerIP string `json:"stream_server_ip" validate:"required"` // 流媒体服务器IP

	// 打击模块
	StrikeIP   string `json:"strike_ip" validate:"required"`   // 打击模块IP
	StrikePort int    `json:"strike_port" validate:"required"` // 打击模块端口
}

// UpdateDeviceRequest 更新设备请求结构
type UpdateDeviceRequest struct {
	Name      *string  `json:"name" validate:"omitempty,min=2,max=64"`
	Longitude *float64 `json:"longitude" validate:"omitempty,longitude"`
	Latitude  *float64 `json:"latitude" validate:"omitempty,latitude"`

	// 侦测模块
	DetectionID   *int    `json:"detection_id" validate:"omitempty"`   // 侦测模块ID
	DetectionIP   *string `json:"detection_ip" validate:"omitempty"`   // 侦测模块IP
	DetectionPort *int    `json:"detection_port" validate:"omitempty"` // 侦测模块端口

	// 解析模块
	ParseID *int    `json:"parse_id" validate:"omitempty"` // 解析模块ID
	ParseIP *string `json:"parse_ip" validate:"omitempty"` // 解析模块IP

	// FPV模块
	FPVIP          *string `json:"fpv_ip" validate:"omitempty"`           // FPV模块IP
	StreamServerIP *string `json:"stream_server_ip" validate:"omitempty"` // 流媒体服务器IP

	// 打击模块
	StrikeIP *string `json:"strike_ip" validate:"omitempty"` // 打击模块IP
}

// // StrikeState 打击状态
// type StrikeState struct {
// 	Mode      int      `json:"mode"`      // 1-宽频 2-无人值守
// 	Status    string   `json:"status"`    // 1-宽频 2-无人值守
// 	Frequency []string `json:"frequency"` // 频段
// }

// // DeviceStatusInfo 设备状态信息
// type DeviceStatusInfo struct {
// 	HeartbeatCount int   `json:"heartbeat_count"` // 心跳计数
// 	Expires        int64 `json:"expires"`         // 过期时间戳
// 	Status         int   `json:"status"`          // 设备状态，0-离线，1-在线
// 	StrikeState    int   `json:"strike_state"`    // 打击状态，0-未打击，1-打击中，2-打击完成
// }
