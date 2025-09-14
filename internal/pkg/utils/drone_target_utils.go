package utils

import "math"

// IsValidCoord 检查坐标是否有效
func IsValidCoord(lon, lat float64, coordThreshold float64) bool {
	if math.Abs(lon) < coordThreshold && math.Abs(lat) < coordThreshold {
		return false
	}
	return lon != 0 || lat != 0 // 同时排除全0坐标
}
