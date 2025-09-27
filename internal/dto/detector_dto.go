package dto

import "uav_defender/internal/models"

type DetectorData struct {
	DeviceID    uint              `json:"device_id"` // 设备ID
	DetectionID uint              `json:"detection_id"`
	Model       string            `json:"model"` // 设备型号
	UAV         string            `json:"drone_model"`
	Freq        float64           `json:"freq"` // 最新频率
	RSSI        float64           `json:"rssi"` // 最新信号强度
	Seq         int64             `json:"seq"`
	Gpio        int64             `json:"gpio"`
	LastTime    models.CustomTime `json:"expires"` // 过期时间戳（UnixNano）
	ID          string            `json:"id"`
	GpioS       []int64           `json:"gpio_s"`
	// 方向角
	Orientation   float64           `json:"orientation"`
	OrientationTS models.CustomTime `json:"orientation_ts"`

	// 更新GpioS的时间
	TS int64 `json:"ts"`
	// 是否已破解标志
	IsCracked bool              `json:"is_cracked"`
	FirstSeen models.CustomTime `json:"first_seen"` // 第一次见到它的时间（Unix）
	// StartTime int64       `json:"start_time"`
	GpiosData [8]GpioData `json:"gpios_data"`
}

type GpioData struct {
	Gpio int64 `json:"gpio"`
	// 总帧数
	Count int64 `json:"count"`
	// rssi的平均值
	RssiAverage float64 `json:"rssi_average"`
	// 最后更新时间
	LastUpdate int64 `json:"last_update"`
}
