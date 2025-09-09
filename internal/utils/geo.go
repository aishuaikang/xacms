package utils

import (
	"fmt"
	"math"
)

/**
地址位置的转换
**/
// 常量定义
const PI = 3.1415926535897932384626
const a = 6378245.0               // 长半轴
const ee = 0.00669342162296594323 // 扁率

// Wgs84ToGcj02  WGS-84 转 GCJ-02 (GPS -> 火星坐标)
func Wgs84ToGcj02(lng, lat float64) (float64, float64) {
	if outOfChina(lng, lat) {
		fmt.Printf("Out of China. No transformation needed. Returning: %.6f, %.6f\n", lng, lat)
		return lng, lat
	}

	//fmt.Printf("Input WGS-84: %.6f, %.6f\n", lng, lat)

	dLat := transformLat(lng-105.0, lat-35.0)
	dLng := transformLon(lng-105.0, lat-35.0)
	radLat := lat / 180.0 * PI
	magic := math.Sin(radLat)
	magic = 1 - ee*magic*magic
	sqrtMagic := math.Sqrt(magic)
	dLat = (dLat * 180.0) / ((a * (1 - ee)) / (magic * sqrtMagic) * PI)
	dLng = (dLng * 180.0) / (a / sqrtMagic * math.Cos(radLat) * PI)

	return truncate(lng+dLng, 6), truncate(lat+dLat, 6)
}

// 判断是否在中国境外
func outOfChina(lng, lat float64) bool {
	return !(lng > 73.66 && lng < 135.05 && lat > 3.86 && lat < 53.55)
}

// 纬度转换
func transformLat(lng, lat float64) float64 {
	ret := -100.0 + 2.0*lng + 3.0*lat + 0.2*lat*lat + 0.1*lng*lat + 0.2*math.Sqrt(math.Abs(lng))
	ret += (20.0*math.Sin(6.0*lng*PI) + 20.0*math.Sin(2.0*lng*PI)) * 2.0 / 3.0
	ret += (20.0*math.Sin(lat*PI) + 40.0*math.Sin(lat/3.0*PI)) * 2.0 / 3.0
	ret += (160.0*math.Sin(lat/12.0*PI) + 320.0*math.Sin(lat*PI/30.0)) * 2.0 / 3.0
	return ret
}

// 经度转换
func transformLon(lng, lat float64) float64 {
	ret := 300.0 + lng + 2.0*lat + 0.1*lng*lng + 0.1*lng*lat + 0.1*math.Sqrt(math.Abs(lng))
	ret += (20.0*math.Sin(6.0*lng*PI) + 20.0*math.Sin(2.0*lng*PI)) * 2.0 / 3.0
	ret += (20.0*math.Sin(lng*PI) + 40.0*math.Sin(lng/3.0*PI)) * 2.0 / 3.0
	ret += (150.0*math.Sin(lng/12.0*PI) + 300.0*math.Sin(lng*PI/30.0)) * 2.0 / 3.0
	return ret
}

// truncate 直接截取到指定的小数位数
func truncate(val float64, precision int) float64 {
	pow := math.Pow(10, float64(precision))
	return math.Floor(val*pow) / pow
}
