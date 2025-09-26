package dto

import (
	"uav_defender/internal/models"
)

// GPS 结构体表示 GPS 坐标
type GPS struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// LdResult 距离关系
type LdResult struct {
	Azimuth     float64 `json:"azimuth"`     // 方位角
	Distance    float64 `json:"distance"`    // 距离
	SensorId    uint    `json:"sensor_id"`   // 传感器ID
	Orientation float64 `json:"orientation"` // 方向
	DeviceLat   float64 `json:"device_lat"`  // 设备纬度
	DeviceLon   float64 `json:"device_lon"`  // 设备经度
	Height      float64 `json:"height"`      // 目标出现的高度
}

type MType uint

const (
	MTypeZK  MType = iota + 1 // 中科
	MTypeSHL                  // 上海L板
)

// 无人机类型
type DroneType int

const (
	DroneTypeUnknown DroneType = iota // 未知
	DroneTypeRC                       // 遥控器/飞手
	DroneTypeUAV                      // 无人机
	DroneTypeBoth                     // 遥控器和无人机
)

type SignType int8

const (
	SignTypeO2O3   SignType = iota + 1 // O2，O3 飞机的报文格式
	SignTypeO3Plus                     // O3+, O4飞机的报文格式
)

type ParseData struct {
	DeviceID       uint                `json:"device_id"`         // 设备ID
	ParseID        uint                `json:"parse_id"`          // 解析ID
	Model          string              `json:"model"`             // 设备型号
	Freq           float64             `json:"freq"`              // 频率
	RSSI           float64             `json:"rssi"`              // 信号强度
	Height         float64             `json:"height"`            // 高度
	Altitude       float64             `json:"altitude"`          // 海拔
	EastV          float64             `json:"eastv"`             // 东向速度
	NorthV         float64             `json:"northv"`            // 北向速度
	UpV            float64             `json:"upv"`               // 垂直速度
	Distance       float64             `json:"distance"`          // 距离
	Serial         string              `json:"serial"`            // 序列号
	DroneGPS       GPS                 `json:"drone_gps"`         // 无人机 GPS 坐标
	HomeGPS        GPS                 `json:"home_gps"`          // 返航点 GPS 坐标
	PilotGPS       GPS                 `json:"rc_gps"`            // 飞行员 GPS 坐标
	Trajectories   models.Trajectories `json:"trajectories"`      // 轨迹点
	TargetId       string              `json:"target_id"`         // 目标ID
	LdResult       LdResult            `json:"ld_result"`         // 距离关系
	DroneType      DroneType           `json:"drone_type"`        // 目标类别:[0-未知,1-遥控器/飞手,2-无人机]
	HasInWhiteList bool                `json:"has_in_white_list"` // 是否在白名单内
	Png            string              `json:"png"`               // base64 编码的图片 二维码
	Sign           SignType            `json:"sign"`              // 1 O2，O3 飞机的报文格式；O3+, O4飞机的报文格式 2.RID
	Mac            string              `json:"mac"`               // MAC 地址
	IntrusionTime  models.CustomTime   `json:"intrusion_time"`    // 入侵时间
	Speed          float64             `json:"speed"`             // 速度
	Expires        models.CustomTime   `json:"expires"`           // 过期时间
}
