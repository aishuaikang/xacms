package models

import (
	"github.com/dromara/carbon/v2"
	"gorm.io/gorm"
)

type CommonModel struct {
	CreatedAt carbon.DateTime `json:"created_at" gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt carbon.DateTime `json:"updated_at" gorm:"autoUpdateTime;comment:更新时间"`
	DeletedAt gorm.DeletedAt  `json:"-" gorm:"index:idx_deleted_at;comment:删除时间"`
}

type CommonNotDeletedModel struct {
	CreatedAt carbon.DateTime `json:"created_at" gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt carbon.DateTime `json:"updated_at" gorm:"autoUpdateTime;comment:更新时间"`
}

// 状态
type Status uint8

const (
	StatusDisabled Status = iota // 禁用
	StatusEnabled                // 启用
)

// IsEnabled 检查状态是否为启用
func (s Status) IsEnabled() bool {
	return s == StatusEnabled
}

// IsDisabled 检查状态是否为禁用
func (s Status) IsDisabled() bool {
	return s == StatusDisabled
}

// String 返回状态的字符串表示
func (s Status) String() string {
	switch s {
	case StatusEnabled:
		return "enabled"
	case StatusDisabled:
		return "disabled"
	default:
		return "unknown"
	}
}
