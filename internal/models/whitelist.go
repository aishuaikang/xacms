package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WhitelistModel struct {
	ID        uuid.UUID  `json:"id" gorm:"primaryKey;type:char(36);comment:唯一ID"`          // 唯一ID
	ParseID   int        `json:"parse_id" gorm:"size:64;not null;comment:解析设备ID"`          // 解析设备ID
	Model     string     `json:"model" gorm:"size:64;not null;comment:型号"`                 // 型号
	Serial    string     `json:"serial" gorm:"size:64;not null;uniqueIndex;comment:目标序列号"` // 目标序列号，唯一索引
	StartTime CustomTime `json:"start_time" gorm:"not null;comment:开始时间"`                  // 开始时间
	EndTime   CustomTime `json:"end_time" gorm:"not null;comment:结束时间"`                    // 结束时间

	CommonModel
}

// TableName 设置表名
func (WhitelistModel) TableName() string {
	return "whitelists"
}

// BeforeCreate GORM钩子，在创建记录之前调用
func (d *WhitelistModel) BeforeCreate(tx *gorm.DB) (err error) {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	// 如果 Status 是零值（StatusDisabled = 0），但我们想要明确设置它
	// 可以根据业务需求决定是否需要默认值
	return
}
