package utils

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2/log"
	"github.com/skip2/go-qrcode"
)

// IsRID 检查数据是否包含 RID 警告的关键字
func IsRID(fullLine []byte) bool {
	str := bytes.ToLower(fullLine)
	return bytes.Contains(str, []byte("rid ")) &&
		bytes.Contains(str, []byte("ssid")) &&
		bytes.Contains(str, []byte("freq")) &&
		bytes.Contains(str, []byte("serial"))
}

// IsDID 如果不是 RID 也不是 Encryption 则认为是DID
func IsDID(fullLine []byte) bool {
	if IsRID(fullLine) || IsEncryption(fullLine) {
		return false
	}
	return true
}

// IsEncryption 检查数据是否包含加密警告的关键字
func IsEncryption(fullLine []byte) bool {
	return bytes.Contains(fullLine, []byte("byte"))
}

type GPS struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// 无人机类型
type DroneType int

const (
	DroneTypeUnknown DroneType = iota // 未知
	DroneTypeRC                       // 遥控器/飞手
	DroneTypeUAV                      // 无人机
	DroneTypeBoth                     // 遥控器和无人机
)

// Trajectory 轨迹点
type Trajectory struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// LdResult 距离关系
type LdResult struct {
	Azimuth     float64 `json:"azimuth"`     // 方位角
	Distance    float64 `json:"distance"`    // 距离
	SensorId    string  `json:"sensor_id"`   // 传感器ID
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

type SignType int8

const (
	SignTypeO2O3   SignType = iota + 1 // O2，O3 飞机的报文格式
	SignTypeO3Plus                     // O3+, O4飞机的报文格式
)

type ParseData struct {
	Device         string       `json:"device"`             // 设备编号
	Model          string       `json:"model"`              // 设备型号
	Freq           float64      `json:"freq"`               // 频率
	RSSI           float64      `json:"rssi"`               // 信号强度
	Expires        int64        `json:"expires"`            // 过期时间
	Height         float64      `json:"height"`             // 高度
	Altitude       float64      `json:"altitude"`           // 海拔
	EastV          float64      `json:"eastv"`              // 东向速度
	NorthV         float64      `json:"northv"`             // 北向速度
	UpV            float64      `json:"upv"`                // 垂直速度
	Distance       float64      `json:"distance"`           // 距离
	Serial         string       `json:"serial"`             // 序列号
	DroneGPS       GPS          `json:"drone_gps"`          // 无人机 GPS 坐标
	HomeGPS        GPS          `json:"return_positioning"` // 家（起飞点）GPS 坐标
	PilotGPS       GPS          `json:"rc_gps"`             // 飞行员 GPS 坐标
	TrajectoryList []Trajectory `json:"trajectory_list"`    // 轨迹
	MType          MType        `json:"m_type"`             // 1 -zk ,2-上海l板
	TargetId       string       `json:"target_id"`          // 目标ID
	LdResult       LdResult     `json:"ld_result"`          // 距离关系
	DroneType      DroneType    `json:"drone_type"`         // 目标类别:[0-未知,1-遥控器/飞手,2-无人机]
	InWhiteList    bool         `json:"in_white_list"`      // 是否在白名单内
	Png            string       `json:"png"`                // base64 编码的图片 二维码
	Sign           SignType     `json:"sign"`               // 1 O2，O3 飞机的报文格式；O3+, O4飞机的报文格式 2.RID
	Mac            string       `json:"mac"`                // MAC 地址
	UpdateTime     int64        `json:"update_time"`        // 更新时间
	Speed          float64      `json:"speed"`              // 速度
}

func ParseRID(fullLine []byte, parseData *ParseData) error {

	fields := splitFields(fullLine)
	for _, field := range fields {
		key, value := parseKV(field)
		if key == nil || value == nil {
			continue
		}
		parseRIDFieldValue(string(key), string(value), parseData)
	}

	if parseData.DroneGPS.Longitude == 0 && parseData.PilotGPS.Longitude != 0 {
		parseData.DroneType = DroneTypeRC
	} else if parseData.DroneGPS.Longitude != 0 && parseData.PilotGPS.Longitude == 0 {
		parseData.DroneType = DroneTypeUAV
	} else if parseData.DroneGPS.Longitude != 0 && parseData.PilotGPS.Longitude != 0 {
		parseData.DroneType = DroneTypeBoth
	}

	parseData.TargetId = parseData.Serial
	parseData.Expires = time.Now().Unix()
	parseData.Png, _ = generateQRCodeBase64(parseData.PilotGPS.Longitude, parseData.PilotGPS.Latitude)
	parseData.Sign = SignTypeO3Plus

	return nil
}

// splitFields 按逗号分割字段，同时处理字段值中可能包含的逗号
func splitFields(fullLine []byte) [][]byte {
	var fields [][]byte
	var current []byte
	for _, part := range bytes.Split(fullLine, []byte{','}) {
		part = bytes.TrimSpace(part)
		if len(part) == 0 {
			continue
		}
		if bytes.Contains(part, []byte{'='}) {
			if len(current) > 0 {
				fields = append(fields, current)
			}
			current = part
		} else if len(current) > 0 {
			current = append(append(current, ','), part...)
		}
	}
	if len(current) > 0 {
		fields = append(fields, current)
	}
	return fields
}

// parseKV 解析键值对，返回小写的键和去除空格的值
func parseKV(field []byte) ([]byte, []byte) {
	kv := bytes.SplitN(bytes.TrimSpace(field), []byte{'='}, 2)
	if len(kv) != 2 {
		return nil, nil
	}
	key := bytes.ToLower(bytes.TrimSpace(kv[0]))
	value := bytes.TrimSpace(kv[1])
	return key, value
}

// parseRIDFieldValue 解析 RID 字段值并赋值给 ParseData 结构体
func parseRIDFieldValue(key string, value string, parseData *ParseData) {
	if key == "rid ssid" {
		key = "ssid"
	}

	switch key {
	case "serial":
		if len(value) > 4 {
			parseData.Serial = value[4:]
		}
	case "model":
		parseData.Model = value
	case "drone_gps":
		parseGPS(value, &parseData.DroneGPS)
	case "pilot_gps":
		parseGPS(value, &parseData.PilotGPS)
	case "height_agl":
		parseData.Height = parseFloat(value)
	case "altitude":
		parseData.Altitude = parseFloat(value)
	case "speed":
		parseData.EastV = parseFloat(value)
	case "vspeed":
		parseData.UpV = parseFloat(value)
	case "rssi":
		parseData.RSSI = parseFloat(value)
	case "freq":
		parseData.Freq = parseFloat(value)
	case "ua_type":
		parseData.MType = MType(parseFloat(value))
	case "mac":
		parseData.Mac = value
	}
}

// parseGPS 解析 GPS 字符串并赋值给 GPS 结构体
func parseGPS(value string, gps *GPS) {
	if gps == nil {
		return
	}

	value = strings.TrimSpace(value)
	if value == "" {
		gps.Longitude, gps.Latitude = 0, 0
		return
	}

	coords := strings.Split(value, ",")
	if len(coords) != 2 {
		gps.Longitude, gps.Latitude = 0, 0
		log.Warnf("无法解析 GPS 坐标: %s", value)
		return
	}

	gps.Longitude = parseFloat(coords[0])
	gps.Latitude = parseFloat(coords[1])
}

// parseFloat 安全地解析浮点数，失败时返回 0.0
func parseFloat(value string) float64 {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0.0
	}

	v, err := strconv.ParseFloat(value, 64)
	if err != nil {
		log.Warnf("无法解析浮点数 '%s': %v", value, err)
		return 0.0
	}
	return v
}

// // GenerateQRCodeBase64 生成飞手位置二维码
// func generateQRCodeBase64(lon, lat float64) (string, error) {
// 	name := url.QueryEscape("飞手位置")
// 	qrURL := fmt.Sprintf(
// 		"https://m.amap.com/share/index/lnglat=%f,%f&name=%s&src=mypage&callnative=1&innersrc=uriapi",
// 		lon, lat, name,
// 	)

// 	qr, err := qrcode.New(qrURL, qrcode.Medium)
// 	if err != nil {
// 		return "", fmt.Errorf("创建二维码失败: %w", err)
// 	}
// 	pngData, err := qr.PNG(256)
// 	if err != nil {
// 		return "", fmt.Errorf("编码 PNG 失败: %w", err)
// 	}
// 	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(pngData), nil
// }

// generateQRCodeBase64 生成飞手位置二维码
func generateQRCodeBase64(lon, lat float64) (string, error) {
	// 检查经纬度有效性
	if lon == 0 && lat == 0 {
		return "", fmt.Errorf("无效的经纬度: 经度和纬度均为0")
	}

	// 构建高德地图分享链接
	qrURL := fmt.Sprintf(
		"https://m.amap.com/share/index/lnglat=%f,%f&name=%s&src=mypage&callnative=1&innersrc=uriapi",
		lon, lat, url.QueryEscape("飞手位置"),
	)

	// 生成二维码并编码为base64
	pngData, err := qrcode.Encode(qrURL, qrcode.Medium, 256)
	if err != nil {
		return "", fmt.Errorf("生成二维码失败: %w", err)
	}

	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(pngData), nil
}

// ParseDID 解析 DID 字段值并赋值给 ParseData 结构体
func ParseDID(fullLine []byte, parseData *ParseData) error {
	fields := splitFields(fullLine)
	for _, field := range fields {
		key, value := parseKV(field)
		if key == nil || value == nil {
			continue
		}
		parseDIDFieldValue(string(key), string(value), parseData)
	}

	parseData.Png, _ = generateQRCodeBase64(parseData.PilotGPS.Longitude, parseData.PilotGPS.Latitude)

	return nil
}

func parseDIDFieldValue(key string, value string, parseData *ParseData) {
	switch key {
	case "device":
		parseData.Device = value
	case "serial":
		parseData.Serial = value
	case "model":
		parseData.Model = parseModel(value)
	case "uuid":
		//目前报文内的uuid 是空的
	case "drone_gps":
		parseGPS(value, &parseData.DroneGPS)
	case "home_gps":
		parseGPS(value, &parseData.HomeGPS)
	case "pilot_gps":
		parseGPS(value, &parseData.PilotGPS)
	case "height":
		parseData.Height = parseFloat(value)
	case "altitude":
		parseData.Altitude = parseFloat(value)
	case "eastv":
		parseData.EastV = parseFloat(value)
	case "nothv":
		parseData.NorthV = parseFloat(value)
	case "upv":
		parseData.UpV = parseFloat(value)
	case "freq":
		parseData.Freq = parseFloat(value)
	case "rssi":
		parseData.RSSI = parseFloat(value)
	case "distance":
		parseData.Distance = parseDistance(value) / 1000
		parseData.Expires = time.Now().Unix()
		parseData.MType = MTypeSHL
		parseData.Sign = SignTypeO2O3
		parseData.TargetId = parseData.Serial
		if parseData.DroneGPS.Longitude == 0 && parseData.PilotGPS.Longitude != 0 {
			parseData.DroneType = 1
		} else if parseData.DroneGPS.Longitude != 0 && parseData.PilotGPS.Longitude == 0 {
			parseData.DroneType = 2
		} else if parseData.DroneGPS.Longitude != 0 && parseData.PilotGPS.Longitude != 0 {
			parseData.DroneType = 3
		}
	}
}

// parseModel 解析飞手型号
func parseModel(val string) string {
	parts := strings.Split(val, "-")
	if len(parts) > 1 {
		return parts[1]
	}
	return val
}

// parseDistance 解析距离，处理可能的单位
func parseDistance(value string) float64 {
	value = strings.TrimSuffix(value, "km") // 去掉 "km" 单位
	return parseFloat(value)
}

// func parseDistance(value []byte) float64 {
// 	distanceStr := string(value)
// 	if strings.HasSuffix(distanceStr, "km") {
// 		distanceStr = strings.TrimSuffix(distanceStr, "km") // 去掉 "km" 单位
// 	}
// 	return parseFloat([]byte(distanceStr))
// }

// fillTrajectoryAndCoords 坐标转换及轨迹
// func fillTrajectoryAndCoords(parseData *ParseData) {
// 	ji, err := service.FindDroneTrajectory(parseData.Serial)
// 	if err != nil {
// 		log.Errorf("获取轨迹失败: %v", err)
// 		return
// 	}
// 	parseData.TrajectoryList = ji
// 	if cache.Map != 3 {
// 		parseData.DroneGPS.Longitude, parseData.DroneGPS.Latitude = Wgs84ToGcj02(parseData.DroneGPS.Longitude, parseData.DroneGPS.Latitude)
// 		parseData.HomeGPS.Longitude, parseData.HomeGPS.Latitude = Wgs84ToGcj02(parseData.HomeGPS.Longitude, parseData.HomeGPS.Latitude)
// 		parseData.PilotGPS.Longitude, parseData.PilotGPS.Latitude = Wgs84ToGcj02(parseData.PilotGPS.Longitude, parseData.PilotGPS.Latitude)
// 	} else {
// 		//alert.DroneGPS.Longi = math.Round(alert.DroneGPS.Longi*1e3) / 1e3
// 		//alert.DroneGPS.Lati = math.Round(alert.DroneGPS.Lati*1e3) / 1e3
// 		//alert.HomeGPS.Longi = math.Round(alert.HomeGPS.Longi*1e3) / 1e3
// 		//alert.HomeGPS.Lati = math.Round(alert.HomeGPS.Lati*1e3) / 1e3
// 		//alert.PilotGPS.Longi = math.Round(alert.PilotGPS.Longi*1e3) / 1e3
// 		//alert.PilotGPS.Lati = math.Round(alert.PilotGPS.Lati*1e3) / 1e3

// 	}

// }

// 连续解密失败计数器
var decryptFailCount int32

func ParseEncryption(fullLine []byte, parseData *ParseData, token *string, isHasSerial *bool) error {
	freq, rssi, hexStr, id, err := extractFreqRssiAndHexString(fullLine)
	if err != nil {
		log.Errorf("提取频率、RSSI 和十六进制字符串失败: %v", err)
		return err
	}

	pd, err := decryptWithAPI(hexStr, *token)
	if err == nil && pd != nil && pd.Serial != "" {
		// 解密成功，重置计数器
		atomic.StoreInt32(&decryptFailCount, 0)

		// merge 解密内容
		parseData.Serial = pd.Serial
		parseData.Model = pd.Model
		parseData.DroneGPS = pd.DroneGPS
		parseData.HomeGPS = pd.HomeGPS
		parseData.PilotGPS = pd.PilotGPS
		parseData.Height = pd.Height
		parseData.Altitude = pd.Altitude
		parseData.EastV = pd.EastV
		parseData.NorthV = pd.NorthV
		parseData.UpV = pd.UpV
		parseData.TargetId = parseData.Serial
		*isHasSerial = true
		parseData.Expires = time.Now().Unix()
		parseData.MType = MTypeSHL
		parseData.Sign = SignTypeO2O3
		parseData.Freq = freq
		parseData.RSSI = rssi
		if parseData.DroneGPS.Longitude == 0 && parseData.PilotGPS.Longitude != 0 {
			parseData.DroneType = DroneTypeRC
		} else if parseData.DroneGPS.Longitude != 0 && parseData.PilotGPS.Longitude == 0 {
			parseData.DroneType = DroneTypeUAV
		} else if parseData.DroneGPS.Longitude != 0 && parseData.PilotGPS.Longitude != 0 {
			parseData.DroneType = DroneTypeBoth
		}
		parseData.Speed = calculateFlightSpeed(parseData.EastV, parseData.NorthV, parseData.UpV)
	} else {
		// 解密失败，增加计数器
		failCount := atomic.AddInt32(&decryptFailCount, 1)

		// 仅当连续失败达到20次时，启用降级方案
		if failCount >= 20 {
			log.Warnf("解密失败(%d/30): %v，启用降级方案", failCount, err)

			parseData.Freq = freq
			parseData.RSSI = rssi
			parseData.Model = "DJI-Drone"
			parseData.Serial = id
			parseData.Expires = time.Now().Unix()
			parseData.Sign = SignTypeO2O3
			parseData.MType = MTypeSHL
			*isHasSerial = true

			// 重置计数器以便后续重新计数
			atomic.StoreInt32(&decryptFailCount, 0)
		} else {
			log.Warnf("解密失败(%d/30): %v", failCount, err)
		}
	}
	return nil
}

// extractFreqRssiAndHexString 处理字节数组输入的版本
func extractFreqRssiAndHexString(data []byte) (float64, float64, string, string, error) {
	input := string(data)

	// 1. 提取 freq
	freqPattern := `freq=([0-9]+\.[0-9]+)`
	reFreq := regexp.MustCompile(freqPattern)
	freqMatches := reFreq.FindStringSubmatch(input)
	if len(freqMatches) < 2 {
		return 0, 0, "", "", fmt.Errorf("freq not found")
	}
	freq, err := strconv.ParseFloat(freqMatches[1], 64)
	if err != nil {
		return 0, 0, "", "", fmt.Errorf("invalid freq format: %v", err)
	}

	// 2. 提取 rssi
	rssiPattern := `rssi=(-?[0-9]+(?:\.[0-9]+)?)`
	reRssi := regexp.MustCompile(rssiPattern)
	rssiMatches := reRssi.FindStringSubmatch(input)
	if len(rssiMatches) < 2 {
		return 0, 0, "", "", fmt.Errorf("rssi not found")
	}
	rssi, err := strconv.ParseFloat(rssiMatches[1], 64)
	if err != nil {
		return 0, 0, "", "", fmt.Errorf("invalid rssi format: %v", err)
	}

	// 3. 提取 encryptedID
	encryptedID := ""
	encryptedIDPattern := `Encypted Mavic_O4_ID=([0-9a-fA-F]+)`
	reEncryptedID := regexp.MustCompile(encryptedIDPattern)
	encryptedIDMatches := reEncryptedID.FindStringSubmatch(input)
	if len(encryptedIDMatches) >= 2 {
		encryptedID = strings.ToLower(encryptedIDMatches[1])
	}

	// 4. 提取 byte 数据部分
	idx := strings.Index(input, "byte,")
	if idx == -1 {
		return 0, 0, "", "", fmt.Errorf("'byte,' not found in input")
	}

	bytePart := input[idx+5:]
	bytePart = strings.TrimSpace(bytePart)
	bytePart = strings.TrimRight(bytePart, ",")
	bytePart = strings.ReplaceAll(bytePart, " ", "")
	bytePart = strings.ReplaceAll(bytePart, "\n", "")

	parts := strings.Split(bytePart, ",")
	var hexStr strings.Builder
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if len(part) == 1 {
			hexStr.WriteString("0" + part)
		} else {
			hexStr.WriteString(strings.ToLower(part))
		}
	}

	return freq, rssi, hexStr.String(), encryptedID, nil
}

func decryptWithAPI(hexStr, token string) (*ParseData, error) {
	url := fmt.Sprintf("http://101.227.171.238:5000/api/yd/decryptl?hex=%s&token=%s", hexStr, token)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		SN       string  `json:"sn"`
		Model    string  `json:"model"`
		Lon      float64 `json:"lon"`
		Lat      float64 `json:"lat"`
		Alt      float64 `json:"alt"`
		Height   float64 `json:"height"`
		X        float64 `json:"x"`
		Y        float64 `json:"y"`
		Z        float64 `json:"z"`
		PilotLon float64 `json:"pilot_lon"`
		PilotLat float64 `json:"pilot_lat"`
		HomeLon  float64 `json:"home_lon"`
		HomeLat  float64 `json:"home_lat"`
	}
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}
	if err := sonic.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	// 构造 alert 结构
	parseData := &ParseData{
		Serial: result.SN,
		Model:  result.Model,
		DroneGPS: GPS{
			Longitude: result.Lon,
			Latitude:  result.Lat,
		},
		HomeGPS: GPS{
			Longitude: result.HomeLon,
			Latitude:  result.HomeLat,
		},
		PilotGPS: GPS{
			Longitude: result.PilotLon,
			Latitude:  result.PilotLat,
		},
		Altitude: result.Alt,
		Height:   result.Height,
		EastV:    result.X,
		NorthV:   result.Y,
		UpV:      result.Z,
	}
	return parseData, nil
}

// calculateFlightSpeed 计算飞行速度
func calculateFlightSpeed(eastV, northV, upV float64) float64 {
	horizontalSpeed := math.Sqrt(eastV*eastV + northV*northV)
	return math.Sqrt(horizontalSpeed*horizontalSpeed + upV*upV)
}

func MergeParseData(oldData, newData ParseData) (ParseData, error) {
	// 更新其他字段
	oldData.DroneGPS = newData.DroneGPS
	oldData.HomeGPS = newData.HomeGPS
	oldData.PilotGPS = newData.PilotGPS
	oldData.Height = newData.Height
	oldData.Speed = newData.Speed
	oldData.Altitude = newData.Altitude
	oldData.EastV = newData.EastV
	oldData.NorthV = newData.NorthV
	oldData.UpV = newData.UpV
	oldData.Freq = newData.Freq
	oldData.RSSI = newData.RSSI
	oldData.Distance = newData.Distance
	oldData.Png = newData.Png
	oldData.TrajectoryList = newData.TrajectoryList
	oldData.InWhiteList = newData.InWhiteList

	// 更新过期时间
	oldData.Expires = time.Now().Unix()

	if newData.Model != "" {
		oldData.Model = newData.Model
	}

	if newData.DroneGPS.Longitude == 0 && newData.PilotGPS.Longitude != 0 {
		oldData.DroneType = DroneTypeRC
	} else if newData.DroneGPS.Longitude != 0 && newData.PilotGPS.Longitude == 0 {
		oldData.DroneType = DroneTypeUAV
	} else if newData.DroneGPS.Longitude != 0 && newData.PilotGPS.Longitude != 0 {
		oldData.DroneType = DroneTypeBoth
	}

	// // 1. 通过设备编号获取设备注册信息
	// atoi, err := strconv.Atoi(oldData.Device) // 直接转换，不需要循环
	// if err != nil {
	// 	global.Logger.Infof("未找到定位设备")
	// 	return err
	// }

	return oldData, nil
}

// Haversine 计算两点之间的距离，单位：米
func Haversine(lat1, lon1, lat2, lon2 float64) float64 {
	if lat2 == 0 || lon2 == 0 {
		return 0
	}
	const R = 6371000.0
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	φ1 := lat1 * math.Pi / 180
	φ2 := lat2 * math.Pi / 180

	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(φ1)*math.Cos(φ2)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return math.Round(R * c)
}

// CalculateBearing 计算两点之间的方位角，单位：度
func CalculateBearing(lat1, lon1, lat2, lon2 float64) float64 {
	φ1 := lat1 * math.Pi / 180
	φ2 := lat2 * math.Pi / 180
	Δλ := (lon2 - lon1) * math.Pi / 180

	y := math.Sin(Δλ) * math.Cos(φ2)
	x := math.Cos(φ1)*math.Sin(φ2) - math.Sin(φ1)*math.Cos(φ2)*math.Cos(Δλ)
	θ := math.Atan2(y, x)
	return math.Mod(θ*180/math.Pi+360, 360)
}
