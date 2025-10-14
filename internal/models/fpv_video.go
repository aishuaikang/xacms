package models

type FpvVideo struct {
	// TODO:增加租户ID字段可为空
	ID             uint64 `json:"id,string" gorm:"primaryKey;autoIncrement;comment:唯一ID"` // 唯一ID
	DeviceID       string `gorm:"column:device_id" json:"device_id"`                      // 设备ID
	PointFrequency int    `json:"point_frequency" gorm:"column:point_frequency" comment:"频点"`
	FileUrl        string `json:"file_url" gorm:"column:file_url;type:text" comment:"文件URL"`
	CommonModel
}

// TableName 设置对应的表名
func (FpvVideo) TableName() string {
	return "fpv_videos"
}
