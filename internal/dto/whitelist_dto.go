package dto

// WhitelistQueryRequest 白名单查询请求
type WhitelistQueryRequest struct {
	BaseQueryRequest
}

// WhitelistCreateRequest 创建白名单请求
type WhitelistCreateRequest struct {
	ParseID   int    `json:"parse_id" validate:"required"`         // 解析设备ID
	Model     string `json:"model" validate:"required"`            // 型号
	Serial    string `json:"serial" validate:"required"`           // 目标序列号
	StartTime int64  `json:"start_time" validate:"required,min=0"` // 开始时间 (时间戳/秒)
	EndTime   int64  `json:"end_time" validate:"required,min=0"`   // 结束时间 (时间戳/秒)
}
