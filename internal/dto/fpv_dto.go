package dto

type FPVWarningData struct {
	DetectionID int    `json:"device"`
	Freq        string `json:"freq"`
	RSSI        string `json:"rssi"`
	IP          string `json:"ip"`
	Time        int64  `json:"time"`
}

// FPVQueryRequest 查询FPV请求
type FPVQueryRequest struct {
	BaseQueryRequest

	Frequency *int   `form:"frequency" validate:"omitempty,min=0"`  // 频点
	StartTime *int64 `form:"start_time" validate:"omitempty,min=0"` // 入侵时间起始 (时间戳/秒)
	EndTime   *int64 `form:"end_time" validate:"omitempty,min=0"`   // 入侵时间结束 (时间戳/秒)
}

// 进入和退出凝视模式请求
type FPVSSERequest struct {
	DetectionID int    `form:"detection_id" validate:"required,min=1"` // 设备ID
	Addr        string `form:"addr" validate:"required"`               // 凝视地址
	Frequency   int    `form:"frequency" validate:"required,min=1"`    // 频点，不能为0
}
