package utils

import (
	"bytes"
	"math"
	"math/rand"
	"strings"
	"uav_defender/internal/dto"
)

// 无人机型号映射
var droneModelMap = map[string]string{
	"Lightbridge_type1": "DJI P4/Spark",
	"Lightbridge_type2": "Autel Evo2V2/Feimi",
	"eWifi_5M":          "DJI Mini1/Air1",
	"eWifi_10M":         "HUBSAN Zino/HUBSAN Eagle",
	"PAL Analog":        "Analog PAL",
	"NTSC Analog":       "Analog NTSC",
	"DJI_OC123_10M":     "DJI-O1/O2/O3 Series Mini 2,3, 4k, Air 2, 3, Mavic 2,3, Avata, P4-2.0",
	"DJI_OC123_20M":     "DJI-O1/O2/O3 Series Mini 2,3, 4k,Air 2, 3, Mavic 2,3, Avata, P4-2.0",
	"DJI_O4_type":       "DJI-O4 Series Mini4, Air3s, Avata2, Neo",
	"Autel_type1":       "Autel nano/nano+/ lite/lite+",
	"Autel_type2":       "Autel Evo2_V3, MAX4T, lite, lite+",
	"Autel_type3":       "Autel Evo2_V3, MAX4T, lite, lite+",
	"Autel_type4":       "Autel Evo2_V3, MAX4T, lite, lite+",
	"Datalink_type1":    "P900/P840/Mavlink",
	"Datalink_type2":    "P900/P840/Mavlink",
	"Datalink_type3":    "DJI/Autel remote",
	"LTE_type0":         "LTE-image feed",
	"LORA":              "LoRa drone",
	"Walksnail":         "Walksnail drone",
	"DJI_O3+":           "Mavic 3 series",
	"O3+_ofdm_datalink": "DJI/Autel remote",
}

// GetDroneModel 根据输入字符串获取无人机型号
func GetDroneModelByModelSource(source string) (string, bool) {
	model, ok := droneModelMap[source]
	return model, ok
}

// IsValidDetectorData 判断是否有效的侦测数据
func IsValidDetectorData(data []byte) bool {
	if len(data) == 0 {
		return false
	}

	expectedKeys := map[string]bool{
		"device": true,
		"model":  true,
		"serial": true,
		"freq":   true,
		"rssi":   true,
		"seq":    true,
		"gpio":   true,
	}

	// 用于检查必需字段是否存在且不为空
	requiredFields := map[string]bool{
		"model": false,
		"freq":  false,
	}

	fields := bytes.Split(data, []byte{','})
	for _, field := range fields {
		field = bytes.TrimSpace(field)
		if len(field) == 0 {
			continue
		}

		kv := bytes.SplitN(field, []byte{'='}, 2)
		if len(kv) != 2 {
			return false
		}

		key := strings.ToLower(string(bytes.TrimSpace(kv[0])))
		value := bytes.TrimSpace(kv[1])

		// 检查key是否在预期的键列表中
		if !expectedKeys[key] {
			return false
		}

		// 检查必需字段是否不为空
		if key == "model" || key == "freq" {
			if len(value) == 0 {
				return false
			}
			requiredFields[key] = true
		}
	}

	// 检查所有必需字段是否都存在
	for _, found := range requiredFields {
		if !found {
			return false
		}
	}

	return true
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

// GenerateRandomID 生成指定长度的随机字符串ID
func GenerateRandomID(length int) string {
	if length <= 0 {
		return ""
	}

	id := make([]rune, length)
	for i := range id {
		id[i] = letters[rand.Intn(len(letters))]
	}
	return string(id)
}

// IsSameDetectorData 判断是否是相同的侦测数据
func IsSameDetectorData(a, b *dto.DetectorData, similarThreshold float64) bool {
	if a.DetectionID != b.DetectionID {
		return false
	}

	// 检测道通无人机（优先级最高）
	isAutelA := isAutelType(a.Model)
	isAutelB := isAutelType(b.Model)
	if isAutelA || isAutelB {
		freqDiff := math.Abs(a.Freq - b.Freq)
		// 道通无人机使用阈值增加25.0
		if freqDiff <= similarThreshold+25.0 {
			// 模型相同或两者都是道通无人机即视为相等
			if a.Model == b.Model || (isAutelA && isAutelB) {
				return true
			}
		}
		return false
	}

	// 处理O3+_ofdm_datalink模型
	isO3A := a.Model == "O3+_ofdm_datalink"
	isO3B := b.Model == "O3+_ofdm_datalink"
	if isO3A || isO3B {
		freqDiff := math.Abs(a.Freq - b.Freq)
		// O3模型使用阈值增加5.0
		if freqDiff <= similarThreshold+5.0 {
			// 模型相同即视为相等
			if a.Model == b.Model {
				return true
			}
		}
		return false
	}

	// 普通模型处理
	freqDiff := math.Abs(a.Freq - b.Freq)
	if freqDiff <= similarThreshold {
		// 同品牌检测（DIJ大疆或其它同型号）
		if (isDIJType(a.Model) && isDIJType(b.Model)) || a.Model == b.Model {
			return true
		}
	}

	return false

}

// isAutelType 是否通道无人机
func isAutelType(model string) bool {
	// return model == "Autel_type1" || model == "Autel_type2" || model == "Autel_type3" || model == "Autel_type4" || model == "Autel_type5"
	return strings.HasPrefix(model, "Autel_type")
}

// isDIJType 是否是大疆类型
func isDIJType(model string) bool {
	return model == "DJI_OC123_10M" || model == "DJI_OC123_20M"
}
