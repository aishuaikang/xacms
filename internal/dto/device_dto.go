package dto

import "uav_defender/internal/models"

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

type StrikeMode int

const (
	StrikeModeIdle       StrikeMode = iota // 空闲
	StrikeModeBroadband                    // 宽频打击
	StrikeModeUnattended                   // 无人值守
)

type StrikeStatus string

const (
	StrikeStatusNotStriked StrikeStatus = "NotStriked" // 未打击
	StrikeStatusStriking   StrikeStatus = "Striking"   // 打击中
	StrikeStatusOffline    StrikeStatus = "Offline"    // 离线
)

// StrikeInfo 打击状态
type StrikeInfo struct {
	Mode      StrikeMode   `json:"mode"`      // 0 空闲 1-宽频打击 2-无人值守
	Status    StrikeStatus `json:"status"`    // 打击状态，0-未打击，1-打击中，2-离线 "NotStriked", "Striking", "Offline"
	Frequency []string     `json:"frequency"` // 频段
}

// DeviceInfoStatus 设备状态枚举
type DeviceInfoStatus int

const (
	DeviceInfoStatusOffline DeviceInfoStatus = 0 // 离线
	DeviceInfoStatusOnline  DeviceInfoStatus = 1 // 在线
)

// DeviceInfo 设备信息
type DeviceInfo struct {
	models.DeviceModel
	HeartbeatCount int              `json:"heartbeat_count"` // 心跳计数
	Expires        int64            `json:"expires"`         // 过期时间戳
	Status         DeviceInfoStatus `json:"status"`          // 设备状态，0-离线，1-在线
	StrikeInfo     StrikeInfo       `json:"strike_info"`     // 打击状态信息
}
