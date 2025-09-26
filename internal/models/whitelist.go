package models

type Whitelist struct {
	ID       uint   `json:"id" gorm:"primaryKey;autoIncrement;comment:唯一ID"`          // 唯一ID
	DeviceID uint   `json:"device_id" gorm:"not null;index;comment:设备ID"`             // 设备ID
	Model    string `json:"model" gorm:"size:64;not null;comment:型号"`                 // 型号
	Serial   string `json:"serial" gorm:"size:64;not null;uniqueIndex;comment:目标序列号"` // 目标序列号，唯一索引

	CommonModel
}

// TableName 设置表名
func (Whitelist) TableName() string {
	return "whitelists"
}
