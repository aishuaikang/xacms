package utils

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"sort"
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

// 遥控器map
var remoteControllerMap = map[string]string{
	"Datalink_type1":    "P900/P840/Mavlink",
	"Datalink_type2":    "P900/P840/Mavlink",
	"Datalink_type3":    "DJI/Autel remote",
	"Walksnail":         "Walksnail drone",
	"O3+_ofdm_datalink": "DJI/Autel remote",
}

// GetDroneModel 根据输入字符串获取无人机型号
func GetDroneModelByModelSource(source string) (string, bool) {
	model, ok := droneModelMap[source]
	return model, ok
}

// GetRemoteControllerModel 根据输入字符串获取遥控器型号
func GetRemoteControllerModelByModelSource(source string) (string, bool) {
	model, ok := remoteControllerMap[source]
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

// AntennaData 天线数据结构
type AntennaData struct {
	Gpio    int64
	RssiAvg float64
}

// CalculateOrientationAngle 根据天线数据计算方向角
func CalculateOrientationAngle(antennas []AntennaData) float64 {
	switch len(antennas) {
	case 0:
		return 500 // 无有效数据
	case 1:
		// 单个天线：直接中心角
		return float64(antennas[0].Gpio) * 45.0
	case 2:
		return calculateTwoAntennaAngle(antennas)
	default:
		return calculateMultipleAntennaAngle(antennas)
	}
}

// calculateTwoAntennaAngle 计算两个天线的方向角
func calculateTwoAntennaAngle(antennas []AntennaData) float64 {
	// 按RSSI排序
	sort.Slice(antennas, func(i, j int) bool {
		return antennas[i].RssiAvg > antennas[j].RssiAvg
	})

	top0, top1 := antennas[0], antennas[1]
	rssiDiff := top0.RssiAvg - top1.RssiAvg

	if rssiDiff > 2.0 {
		// 差值>2dB：使用最强天线
		return float64(top0.Gpio) * 45.0
	} else {
		// 差值≤2dB：计算中间角度
		return ComputeMiddleAngle(top0.Gpio, top1.Gpio)
	}
}

// calculateMultipleAntennaAngle 计算多个天线（>=3）的方向角
func calculateMultipleAntennaAngle(antennas []AntennaData) float64 {
	// 按RSSI从大到小排序
	sort.Slice(antennas, func(i, j int) bool {
		return antennas[i].RssiAvg > antennas[j].RssiAvg
	})

	top0, top1, top2 := antennas[0], antennas[1], antennas[2]

	if top0.RssiAvg-top1.RssiAvg >= 2.0 {
		// 规则1: 最高与第二高相差≥2dB
		return float64(top0.Gpio) * 45.0
	} else if top1.RssiAvg-top2.RssiAvg >= 2.0 {
		// 规则2: 第二高与第三高相差≥2dB
		return ComputeMiddleAngle(top0.Gpio, top1.Gpio)
	} else {
		// 规则2.1: 三个天线处理
		return handleThreeAntennaCase(top0, top1, top2)
	}
}

// handleThreeAntennaCase 处理三个天线的情况
func handleThreeAntennaCase(top0, top1, top2 AntennaData) float64 {
	gpios := []int64{top0.Gpio, top1.Gpio, top2.Gpio}

	if AreAdjacent(gpios) {
		// 相邻取中间天线的中心角
		sort.Slice(gpios, func(i, j int) bool { return gpios[i] < gpios[j] })
		return float64(gpios[1]) * 45.0
	} else {
		// 不相邻 - 选择RSSI差值最小的两个天线
		diff01 := math.Abs(top0.RssiAvg - top1.RssiAvg)
		diff12 := math.Abs(top1.RssiAvg - top2.RssiAvg)
		diff02 := math.Abs(top0.RssiAvg - top2.RssiAvg)

		// 找到最小差值
		minDiff := math.Min(diff01, math.Min(diff12, diff02))

		switch minDiff {
		case diff01:
			return ComputeMiddleAngle(top0.Gpio, top1.Gpio)
		case diff12:
			return ComputeMiddleAngle(top1.Gpio, top2.Gpio)
		default: // diff02
			return ComputeMiddleAngle(top0.Gpio, top2.Gpio)
		}
	}
}

// CalculateNewOrientationAngle 根据候选角度和当前角度计算新的方向角
func CalculateNewOrientationAngle(candidateAngle, currentAngle float64) float64 {
	diff := candidateAngle - currentAngle

	// 标准化角度差值到 [-180, 180) 范围内
	for diff >= 180 {
		diff -= 360
	}
	for diff < -180 {
		diff += 360
	}

	absDiff := math.Abs(diff)

	switch {
	case absDiff <= 22.5:
		return currentAngle // 方向角不变
	case absDiff <= 45:
		return candidateAngle // 使用候选方向角
	default:
		var newAngle float64
		if diff > 0 {
			newAngle = currentAngle + 45
		} else {
			newAngle = currentAngle - 45
		}
		// 确保角度在 [0, 360) 范围内
		return math.Mod(newAngle+360, 360)
	}
}

// BuildStopDirectionDetectionCommand 构建停止定向侦测命令
func BuildStopDirectionDetectionCommand() string {
	return "start -set_ant 255,-turn_off_gpio 3\n"
}

// BuildDirectionDetectionCommand 构建定向侦测命令
func BuildDirectionDetectionCommand(freq float64) string {
	return fmt.Sprintf("start -freq %f -set_ant 255,-turn_on_gpio 3\n", freq)
}

// 计算两个天线中心角的中间角度
func ComputeMiddleAngle(gpio1, gpio2 int64) float64 {
	angle1 := float64(gpio1) * 45.0
	angle2 := float64(gpio2) * 45.0

	radian1 := angle1 * math.Pi / 180
	radian2 := angle2 * math.Pi / 180

	// 将角度转为向量并求平均
	x1, y1 := math.Cos(radian1), math.Sin(radian1)
	x2, y2 := math.Cos(radian2), math.Sin(radian2)
	x := (x1 + x2) / 2
	y := (y1 + y2) / 2

	// 处理向量接近零的情况
	if math.Abs(x) < 1e-9 && math.Abs(y) < 1e-9 {
		return 0
	}

	// 计算平均向量的角度
	radian := math.Atan2(y, x)
	angle := radian * 180 / math.Pi
	if angle < 0 {
		angle += 360
	}
	return math.Floor(angle)
}

// 判断三个天线是否相邻
func AreAdjacent(gpios []int64) bool {
	sorted := make([]int64, len(gpios))
	copy(sorted, gpios)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	// 预定义的连续组合（已排序）
	combinations := [][]int64{
		{0, 1, 2}, {1, 2, 3}, {2, 3, 4}, {3, 4, 5},
		{4, 5, 6}, {5, 6, 7}, {0, 6, 7}, {0, 1, 7},
	}

	for _, comb := range combinations {
		if sorted[0] == comb[0] && sorted[1] == comb[1] && sorted[2] == comb[2] {
			return true
		}
	}
	return false
}
