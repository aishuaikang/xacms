package models

type FPVVideo struct {
	ID        uint   `json:"id" gorm:"primaryKey;autoIncrement;comment:唯一ID"` // 唯一ID
	DeviceID  uint   `json:"device_id" gorm:"not null;index;comment:设备ID"`    // 设备ID
	Frequency int    `json:"frequency" gorm:"not null;comment:频点"`            // 频点
	Filename  string `json:"filename" gorm:"size:255;not null;comment:文件名"`   // 文件名

	CommonModel
}

// TableName 设置表名
func (FPVVideo) TableName() string {
	return "fpv_videos"
}
