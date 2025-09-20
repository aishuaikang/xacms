package models

import (
	"bytes"
	"database/sql/driver"
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	"gorm.io/gorm"
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

type DroneTargetModel struct {
	ID            uuid.UUID     `json:"id" gorm:"primaryKey;type:char(36);comment:唯一ID"`     // 唯一ID
	Serial        string        `json:"serial" gorm:"size:128;index;not null;comment:目标序列号"` // 目标序列号，普通索引
	Model         string        `json:"model" gorm:"size:64;not null;comment:无人机型号"`         // 无人机型号
	Device        int           `json:"device" gorm:"size:64;not null;comment:捕获设备"`         // 捕获设备
	Distance      float64       `json:"distance" gorm:"comment:目标距离"`                        // 目标距离
	DroneLng      float64       `json:"drone_lng" gorm:"comment:无人机经度"`                      // 无人机经度
	DroneLat      float64       `json:"drone_lat" gorm:"comment:无人机纬度"`                      // 无人机纬度
	Height        float64       `json:"height" gorm:"comment:无人机高度"`                         // 无人机高度
	Frequency     float64       `json:"frequency" gorm:"not null;comment:频点"`                // 频点
	Trajectories  Trajectories  `json:"trajectories" gorm:"type:text;comment:轨迹"`            // 轨迹
	PilotLng      float64       `json:"pilot_lng" gorm:"comment:飞手经度"`                       // 飞手经度
	PilotLat      float64       `json:"pilot_lat" gorm:"comment:飞手纬度"`                       // 飞手纬度
	DetectionType DetectionType `json:"detection_type" gorm:"not null;comment:探测类型"`         // 探测类型
	VanishTime    CustomTime    `json:"vanish_time" gorm:"not null;comment:消失时间"`            // 消失时间

	CommonModel
}

// TableName 设置表名
func (DroneTargetModel) TableName() string {
	return "drone_targets"
}

// BeforeCreate GORM钩子，在创建记录之前调用
func (d *DroneTargetModel) BeforeCreate(tx *gorm.DB) (err error) {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	// 如果 Status 是零值（StatusDisabled = 0），但我们想要明确设置它
	// 可以根据业务需求决定是否需要默认值
	return
}
