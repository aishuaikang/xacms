package models

import (
	"bytes"
	"database/sql/driver"
	"fmt"

	"github.com/bytedance/sonic"
)

// Trajectory 无人机轨迹点
type Trajectory struct {
	Lat    float64 `json:"lat"`
	Lon    float64 `json:"lon"`
	Height float64 `json:"height"`
}

type TrajectorySlice []Trajectory

// 为 TrajectorySlice 实现 GORM 的 Scanner 和 Valuer 接口
func (ths *TrajectorySlice) Scan(value any) error {
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

func (ths TrajectorySlice) Value() (driver.Value, error) {
	return sonic.Marshal(ths)
}

type DetectionType int

const (
	DetectionTypeDetect DetectionType = iota + 1 // 侦测模块加入的
	DetectionTypeParse                           // 解析模块加入的
)

type DroneType uint8

const (
	DroneTypeUnknown       DroneType = iota // 0 - 未知
	DroneTypeRemoteControl                  // 1 - 遥控器/飞手
	DroneTypeUAV                            // 2 - 无人机
	DroneTypeCombined                       // 3 - 无人机+遥控器
)

type DroneTarget struct {
	ID               uint64          `json:"id,string" gorm:"primaryKey;autoIncrement"` // 唯一ID
	TargetId         string          `json:"target_id"`                                 // 目标ID,用于关联目标行为
	Model            string          `json:"model"`                                     // 无人机型号,无线侦测时有效
	UAV              string          `json:"drone_model" gorm:"column:drone_model"`     // 无人机型号,无线侦测时有效
	Device           string          `json:"device" gorm:"default:''"`                  // JSON格式的捕获目标的设备类型及ID列表
	RSSI             float64         `json:"rssi"`                                      // 信号强度,RSSI值,单位dBm,无线侦测时有效
	FreqPoint        float64         `json:"freq_point"`                                // 发现目标无线传输链路的当前频点,MHz为单位, 0,0表示未检测到,无线侦测时有效
	DetectionType    DetectionType   `json:"detection_type"`                            // 1是侦测模块加入的 2是解析模块加入的
	Distance         float64         `json:"distance"`                                  // 目标距离,单位为m
	LocationLng      float64         `json:"location_lng"`                              // 目标出现的经度
	LocationLat      float64         `json:"location_lat"`                              // 目标出现的纬度
	Height           float64         `json:"height"`                                    // 目标出现的高度
	Trajectory       TrajectorySlice `json:"trajectory"`                                // JSON格式的无人机航迹,从无人机侦测设备直接获取或根据无人机每次经纬度和高度的上报信息自动填入该字段,第一个经纬度坐标是出现位置,最后一个是消失位置
	RemoteControlLat float64         `json:"remote_control_lat"`                        // 遥控器出现的纬度
	RemoteControlLon float64         `json:"remote_control_lon"`                        // 遥控器出现的经度
	DroneType        DroneType       `json:"drone_type"`                                // 目标类别:[0-未知,1-遥控器/飞手,2-无人机,3-无人机+遥控器]
	Orientation      float64         `json:"orientation"`                               // 目标出现的方位,正北为0°,0-360°
	Description      string          `json:"description"`                               // 目标描述,手动编辑录入

	CommonModel
}

func (DroneTarget) TableName() string {
	return "drone_targets"
}
