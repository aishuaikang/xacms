package dto

// WhitelistQueryRequest 白名单查询请求
type WhitelistQueryRequest struct {
	BaseQueryRequest
}

// WhitelistCreateRequest 创建白名单请求
type WhitelistCreateRequest struct {
	DeviceID uint   `json:"device_id" validate:"required"` // 设备ID
	Model    string `json:"model" validate:"required"`     // 型号
	Serial   string `json:"serial" validate:"required"`    // 目标序列号
}
