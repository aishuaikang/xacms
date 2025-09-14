package models

import (
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2/log"
)

type CommonModel struct {
	CreatedAt CustomTime `json:"created_at" gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt CustomTime `json:"updated_at" gorm:"autoUpdateTime;comment:更新时间"`
	// DeletedAt gorm.DeletedAt  `json:"-" gorm:"index;comment:删除时间"`
}

type CommonNotDeletedModel struct {
	CreatedAt CustomTime `json:"created_at" gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt CustomTime `json:"updated_at" gorm:"autoUpdateTime;comment:更新时间"`
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

// 自定义时间类型
type CustomTime time.Time

// MarshalJSON 自定义时间的 JSON 序列化
func (ct CustomTime) MarshalJSON() ([]byte, error) {
	formatted := time.Time(ct).Format(time.DateTime)
	log.Infof("格式化时间: %s", formatted)
	return []byte(formatted), nil
}

// UnmarshalJSON 自定义时间的 JSON 反序列化
func (ct *CustomTime) UnmarshalJSON(data []byte) error {
	t, err := time.Parse(time.DateTime, string(data))
	if err != nil {
		log.Errorf("反序列化时间失败: %v", err)
		return err
	}
	*ct = CustomTime(t)
	log.Infof("反序列化时间: %s", t)
	return nil
}

// Value 实现 driver.Valuer 接口，用于数据库存储
func (ct CustomTime) Value() (driver.Value, error) {
	t := time.Time(ct)
	if t.IsZero() {
		log.Warn("时间为零值，返回 nil")
		return nil, nil
	}
	log.Infof("存储时间: %s", t)
	return t, nil
}

// IsZero 检查 CustomTime 是否为零值
func (ct CustomTime) IsZero() bool {
	return time.Time(ct).IsZero()
}

// Scan 从数据库扫描时间值
func (ct *CustomTime) Scan(value interface{}) error {
	if value == nil {
		*ct = CustomTime(time.Time{})
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		*ct = CustomTime(v)
		log.Infof("扫描时间: %s", v)
		return nil
	default:
		log.Errorf("不支持的扫描类型: %T", v)
		return fmt.Errorf("unsupported scan type: %T", v)
	}
}
