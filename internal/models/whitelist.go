package models

type Whitelist struct {
	ID       uint64 `json:"id,string" gorm:"primaryKey;autoIncrement;comment:唯一ID"` // 唯一ID
	DeviceID string `json:"device_id"`
	Model    string `json:"model"`
	UavId    string `json:"uav_id" gorm:"uniqueIndex;not null"` // 无人机ID，唯一索引

	CommonModel
}

func (Whitelist) TableName() string {
	return "whitelists"
}
