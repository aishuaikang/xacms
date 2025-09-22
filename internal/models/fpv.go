package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FPVModel struct {
	ID        uuid.UUID `json:"id" gorm:"primaryKey;type:char(36);comment:唯一ID"` // 唯一ID
	DeviceID  uuid.UUID `json:"device_id" gorm:"type:char(36);comment:设备ID"`     // 设备ID
	Frequency int       `json:"frequency" gorm:"not null;comment:频点"`            // 频点
	FileName  string    `json:"file_name" gorm:"size:255;not null;comment:文件名"`  // 文件名

	CommonModel
}

// TableName 设置表名
func (FPVModel) TableName() string {
	return "fpvs"
}

// BeforeCreate GORM钩子，在创建记录之前调用
func (d *FPVModel) BeforeCreate(tx *gorm.DB) (err error) {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	// 如果 Status 是零值（StatusDisabled = 0），但我们想要明确设置它
	// 可以根据业务需求决定是否需要默认值
	return
}
