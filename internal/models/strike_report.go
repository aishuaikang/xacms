package models

import (
	"bytes"
	"database/sql/driver"
	"fmt"

	"github.com/bytedance/sonic"
)

type StrikeMode int

const (
	StrikeModeWideband   StrikeMode = 1 // 宽频打击
	StrikeModeUnattended StrikeMode = 2 // 无人值守
)

type StrikeResult int

const (
	StrikeResultSuccess StrikeResult = 1 // 成功
	StrikeResultFailure StrikeResult = 2 // 失败
)

type StringSlice []string

// 为 TrajectorySlice 实现 GORM 的 Scanner 和 Valuer 接口
func (ths *StringSlice) Scan(value any) error {
	switch v := value.(type) {
	case string:
		if v == "" || v == "null" {
			*ths = nil
			return nil
		}
		bytes := []byte(v)
		return sonic.Unmarshal(bytes, ths)
	case []byte:

		if bytes.Equal(v, []byte("")) || bytes.Equal(v, []byte("null")) {
			*ths = nil
			return nil
		}

		return sonic.Unmarshal(v, ths)
	case nil:
		*ths = nil
		return nil
	default:
		return fmt.Errorf("unsupported type: %T", v)
	}
}

func (ths StringSlice) Value() (driver.Value, error) {
	return sonic.Marshal(ths)
}

func (ths StringSlice) StringSlice() []string {
	return ths
}

type StrikeReport struct {
	ID         uint64       `json:"id,string" gorm:"primaryKey;autoIncrement"` // 唯一ID
	StrikeMode StrikeMode   `gorm:"column:strike_mode" json:"strike_mode"`     // 打击模式 1-宽频，2-无人值守
	DeviceID   string       `gorm:"column:device_id" json:"device_id"`         // 设备ID
	Frequency  StringSlice  `gorm:"column:frequency" json:"frequency"`         // 频率
	Model      *StringSlice `gorm:"column:model" json:"model"`                 // 机型
	TargetID   *StringSlice `gorm:"column:target_id" json:"target_id"`         // 目标ID
	StartTime  CustomTime   `gorm:"column:start_time" json:"start_time"`       // 开始时间
	EndTime    *CustomTime  `gorm:"column:end_time" json:"end_time"`
	Duration   int64        `gorm:"column:duration" json:"duration"` // 持续打击时间
	Result     StrikeResult `gorm:"column:result" json:"result"`     // 结果 1-成功，2-失败，3-进行中
	CommonModel
}

func (StrikeReport) TableName() string {
	return "strike_reports"
}
