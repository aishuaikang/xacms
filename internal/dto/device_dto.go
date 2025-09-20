package dto

import (
	fpv_fsm "uav_defender/internal/app/devices/fms/fpv"
	parse_fsm "uav_defender/internal/app/devices/fms/parse"
	"uav_defender/internal/models"
)

// CreateDeviceRequest 创建设备请求结构
type CreateDeviceRequest struct {
	Name      string   `json:"name" validate:"required,min=2,max=64"`
	Longitude *float64 `json:"longitude" validate:"omitempty,longitude"`
	Latitude  *float64 `json:"latitude" validate:"omitempty,latitude"`

	// 侦测模块
	DetectionID   int    `json:"detection_id" validate:"required"`   // 侦测模块ID
	DetectionIP   string `json:"detection_ip" validate:"required"`   // 侦测模块IP
	DetectionPort int    `json:"detection_port" validate:"required"` // 侦测模块端口

	// 解析模块
	ParseID int    `json:"parse_id" validate:"required"` // 解析模块ID
	ParseIP string `json:"parse_ip" validate:"required"` // 解析模块IP

	// FPV模块
	FPVIP  string `json:"fpv_ip" validate:"required"`  // FPV模块IP
	RTSPIP string `json:"rtsp_ip" validate:"required"` // RTSPIP

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
	FPVIP  *string `json:"fpv_ip" validate:"omitempty"`  // FPV模块IP
	RTSPIP *string `json:"rtsp_ip" validate:"omitempty"` // RTSPIP

	// 打击模块
	StrikeIP *string `json:"strike_ip" validate:"omitempty"` // 打击模块IP
}

// DeviceInfo 设备信息
type DeviceInfo struct {
	models.DeviceModel
	FPVFsm   *fpv_fsm.FPVFsm
	ParseFsm *parse_fsm.ParseFsm
}

// DeviceDisplayInfo 设备展示信息
type DeviceDisplayInfo struct {
	models.DeviceModel
	FPVState   fpv_fsm.FPVState     `json:"fpv_state"`   // FPV状态
	ParseState parse_fsm.ParseState `json:"parse_state"` // 解析状态
}
