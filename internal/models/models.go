package models

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"
	"time"
	"uav_defender/internal/pkg/global"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type CommonModel struct {
	CreatedAt CustomTime     `json:"created_at" gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt CustomTime     `json:"updated_at" gorm:"autoUpdateTime;comment:更新时间"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index;comment:删除时间"`
}

type CommonNotDeletedModel struct {
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
	if ts, err := strconv.ParseInt(s, 10, 64); err == nil {
		global.Logger.Info("UnmarshalJSON", zap.String("input", s), zap.Int64("parsed", ts))
		*ct = CustomTime(time.Unix(ts, 0))
		return nil
	}
	global.Logger.Info("UnmarshalJSON", zap.String("input", s), zap.String("info", "not a timestamp"))
	t, err := time.Parse("2006-01-02 15:04:05", s)
	if err != nil {
		return err
	}
	global.Logger.Info("UnmarshalJSON", zap.String("input", s), zap.Time("parsed", t))
	*ct = CustomTime(t)
	return nil
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
func (ct *CustomTime) Scan(value interface{}) error {
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
