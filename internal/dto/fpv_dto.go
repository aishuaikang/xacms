package dto

import "github.com/google/uuid"

type FPVWarningData struct {
	// DetectionID int    `json:"device"`
	DeviceID uuid.UUID `json:"device_id"`
	Freq     string    `json:"freq"`
	RSSI     string    `json:"rssi"`
	// IP          string `json:"ip"`
	Time int64 `json:"time"`
}

// FPVQueryRequest 查询FPV请求
type FPVQueryRequest struct {
	BaseQueryRequest

	Frequency *int   `form:"frequency" validate:"omitempty,min=0"`  // 频点
	StartTime *int64 `form:"start_time" validate:"omitempty,min=0"` // 入侵时间起始 (时间戳/秒)
	EndTime   *int64 `form:"end_time" validate:"omitempty,min=0"`   // 入侵时间结束 (时间戳/秒)
}

// FPVSSERequest 进入和退出凝视模式请求
type FPVSSERequest struct {
	DeviceID  string `form:"device_id" validate:"required,min=1,uuid"` // 设备ID
	Frequency int    `form:"frequency" validate:"required,min=1"`      // 频点，不能为0
	// Addr        string `form:"addr" validate:"required"`               // 凝视地址
}

// AddFPVRequest 添加FPV请求
type AddFPVRequest struct {
	DeviceID  uuid.UUID `json:"device_id" validate:"required,min=1"` // 设备ID
	Frequency int       `json:"frequency" validate:"required,min=1"` // 频点，不能为0
	FileName  string    `json:"file_name" validate:"required"`       // 文件名
}

// SetFrequencyRequest 设置频点请求
type SetFrequencyRequest struct {
	DeviceID  uuid.UUID `json:"device_id" validate:"required,min=1"` // 设备ID
	Frequency int       `json:"frequency" validate:"required,min=1"` // 频点，不能为0
	// Addr        string `json:"addr" validate:"required,min=1"`         // 地址
}
