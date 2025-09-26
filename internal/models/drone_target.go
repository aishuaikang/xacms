package models

import (
	"bytes"
	"database/sql/driver"
	"fmt"

	"github.com/bytedance/sonic"
)

// 1是侦测模块 2是解析模块
type DetectionType int

const (
	DetectionTypeDetection DetectionType = iota + 1 // 侦测模块
	DetectionTypeParse                              // 解析模块
)

// Trajectory 轨迹点
type Trajectory struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
	Height    float64 `json:"height"`
}

type Trajectories []Trajectory

// 为 TrajectoryHeightSlice 实现 GORM 的 Scanner 和 Valuer 接口
func (ths *Trajectories) Scan(value interface{}) error {
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

func (ths Trajectories) Value() (driver.Value, error) {
	return sonic.Marshal(ths)
}

type DroneTarget struct {
	ID             uint          `json:"id" gorm:"primaryKey;autoIncrement;comment:唯一ID"`     // 唯一ID
	Serial         string        `json:"serial" gorm:"size:128;index;not null;comment:目标序列号"` // 目标序列号，普通索引
	Model          string        `json:"model" gorm:"size:64;not null;comment:无人机型号"`         // 无人机型号
	ModelSource    string        `json:"model_source" gorm:"size:64;not null;comment:原无人机型号"` // 原无人机型号
	DeviceID       uint          `json:"device_id" gorm:"not null;index;comment:设备ID"`        // 设备ID
	SensorID       uint          `json:"sensor_id" gorm:"not null;index;comment:传感器ID"`       // 传感器ID
	Distance       float64       `json:"distance" gorm:"comment:目标距离"`                        // 目标距离
	Longitude      float64       `json:"longitude" gorm:"comment:无人机经度"`                      // 无人机经度
	Latitude       float64       `json:"latitude" gorm:"comment:无人机纬度"`                       // 无人机纬度
	Height         float64       `json:"height" gorm:"comment:无人机高度"`                         // 无人机高度
	Frequency      float64       `json:"frequency" gorm:"not null;comment:频点"`                // 频点
	Trajectories   Trajectories  `json:"trajectories" gorm:"type:text;comment:轨迹"`            // 轨迹
	PilotLongitude float64       `json:"pilot_longitude" gorm:"comment:飞手经度"`                 // 飞手经度
	PilotLatitude  float64       `json:"pilot_latitude" gorm:"comment:飞手纬度"`                  // 飞手纬度
	DetectionType  DetectionType `json:"detection_type" gorm:"not null;comment:探测类型"`         // 探测类型
	VanishTime     CustomTime    `json:"vanish_time" gorm:"not null;comment:消失时间"`            // 消失时间

	CommonModel
}

// TableName 设置表名
func (DroneTarget) TableName() string {
	return "drone_targets"
}
