package dto

type FPVWarningData struct {
	DeviceID uint   `json:"device_id"`
	Freq     string `json:"freq"`
	RSSI     string `json:"rssi"`
	Time     int64  `json:"time"`
}

// FPVQueryRequest 查询FPV请求
type FPVQueryRequest struct {
	BaseQueryRequest

	Frequency *int `form:"frequency" validate:"omitempty,min=0"` // 频点
}

// FPVSSERequest 进入和退出凝视模式请求
type FPVSSERequest struct {
	DeviceID  uint `form:"device_id,string" validate:"required"` // 设备ID
	Frequency int  `form:"frequency" validate:"required,min=1"`  // 频点，不能为0
}

// AddFPVRequest 添加FPV请求
type AddFPVRequest struct {
	DeviceID  uint   `json:"device_id" validate:"required"`       // 设备ID
	Frequency int    `json:"frequency" validate:"required,min=1"` // 频点，不能为0
	Filename  string `json:"filename" validate:"required"`        // 文件名
}

// SetFrequencyRequest 设置频点请求
type SetFrequencyRequest struct {
	DeviceID  uint `json:"device_id" validate:"required"`       // 设备ID
	Frequency int  `json:"frequency" validate:"required,min=1"` // 频点，不能为0
}
