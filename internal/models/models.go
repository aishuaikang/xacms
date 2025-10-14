package models

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type CommonModel struct {
	CreatedAt CustomTime `json:"created_at" gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt CustomTime `json:"updated_at" gorm:"autoUpdateTime;comment:更新时间"`
}

// 自定义时间类型
type CustomTime time.Time

func (ct CustomTime) Time() time.Time {
	return time.Time(ct)
}

// MarshalJSON 自定义时间的 JSON 序列化
func (ct CustomTime) MarshalJSON() ([]byte, error) {
	formatted := time.Time(ct).Format(time.DateTime)
	return []byte(`"` + formatted + `"`), nil
}

func (ct *CustomTime) UnmarshalJSON(data []byte) error {
	s := string(data)
	s = strings.Trim(s, "\"")

	// 1. 尝试解析为时间戳（秒）
	if ts, err := strconv.ParseInt(s, 10, 64); err == nil {
		*ct = CustomTime(time.Unix(ts, 0))
		return nil
	}

	// 2. 尝试多种时间格式，使用本地时区
	formats := []string{
		time.DateTime, // "2006-01-02 15:04:05"
		time.DateOnly, // "2006-01-02"
		time.RFC3339,  // "2006-01-02T15:04:05Z07:00"
	}

	for _, format := range formats {
		if t, err := time.ParseInLocation(format, s, time.Local); err == nil {
			*ct = CustomTime(t)
			return nil
		}
	}

	return fmt.Errorf("无法解析时间: %s", s)
}

// Value 实现 driver.Valuer 接口，用于数据库存储
func (ct CustomTime) Value() (driver.Value, error) {
	t := time.Time(ct)
	if t.IsZero() {
		return nil, nil
	}

	return t, nil
}

// IsZero 检查 CustomTime 是否为零值
func (ct CustomTime) IsZero() bool {
	return time.Time(ct).IsZero()
}

// Scan 从数据库扫描时间值
func (ct *CustomTime) Scan(value any) error {
	if value == nil {
		*ct = CustomTime(time.Time{})
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		*ct = CustomTime(v)
		return nil
	default:
		return fmt.Errorf("unsupported scan type: %T", v)
	}

}

// // 状态
// type Status uint8

// const (
// 	StatusDisabled Status = iota // 禁用
// 	StatusEnabled                // 启用
// )

// // IsEnabled 检查状态是否为启用
// func (s Status) IsEnabled() bool {
// 	return s == StatusEnabled
// }

// // IsDisabled 检查状态是否为禁用
// func (s Status) IsDisabled() bool {
// 	return s == StatusDisabled
// }
