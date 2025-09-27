package utils

import (
	"math"
	"uav_defender/internal/dto"
)

// Distance 计算两点之间的距离（米）
func Distance(gps1, gps2 dto.GPS) float64 {
	if gps1.Latitude == 0 && gps1.Longitude == 0 || gps2.Latitude == 0 && gps2.Longitude == 0 {
		return 0
	}
	const EarthRadius = 6371000.0 // 米
	dLat := (gps2.Latitude - gps1.Latitude) * math.Pi / 180
	dLon := (gps2.Longitude - gps1.Longitude) * math.Pi / 180
	lat1Rad := gps1.Latitude * math.Pi / 180
	lat2Rad := gps2.Latitude * math.Pi / 180

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return EarthRadius * c
}

// Bearing 计算两点之间的方位角（度）
func Bearing(gps1, gps2 dto.GPS) float64 {
	lat1Rad := gps1.Latitude * math.Pi / 180
	lat2Rad := gps2.Latitude * math.Pi / 180
	dLon := (gps2.Longitude - gps1.Longitude) * math.Pi / 180

	y := math.Sin(dLon) * math.Cos(lat2Rad)
	x := math.Cos(lat1Rad)*math.Sin(lat2Rad) -
		math.Sin(lat1Rad)*math.Cos(lat2Rad)*math.Cos(dLon)
	theta := math.Atan2(y, x)
	bearing := math.Mod(theta*180/math.Pi+360, 360)
	return bearing
}

// 定义有效坐标阈值
const coordThreshold = 0.001

// IsValidCoord 检查经纬度是否在有效范围内，并排除无效或全0坐标（WGS84）
func IsValidCoord(longitude, latitude float64) bool {
	if math.Abs(longitude) < coordThreshold && math.Abs(latitude) < coordThreshold {
		return false
	}

	if longitude >= -180 && longitude <= 180 && latitude >= -90 && latitude <= 90 {
		return true
	} else {
		return false

	}
}

// IsValidCoordPtr 检查经纬度指针是否为nil，且值在有效范围内，并排除无效或全0坐标（WGS84）
func IsValidCoordPtr(longitude, latitude *float64) bool {
	if longitude == nil || latitude == nil {
		return false
	}

	if math.Abs(*longitude) < coordThreshold && math.Abs(*latitude) < coordThreshold {
		return false
	}

	if *longitude >= -180 && *longitude <= 180 && *latitude >= -90 && *latitude <= 90 {
		return true
	} else {
		return false

	}

}

// CalculateFlightSpeed 计算飞行速度（米/秒）
// 参数为东向、北向、上向速度分量
func CalculateFlightSpeed(eastV, northV, upV float64) float64 {
	return math.Sqrt(eastV*eastV + northV*northV + upV*upV)
}
