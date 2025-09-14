package devices

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net"
	"strings"
	"time"
	conn_ "xacms/internal/app/devices/conn"
	"xacms/internal/cache"
	"xacms/internal/dto"
	"xacms/internal/models"
	"xacms/internal/pkg/config"
	"xacms/internal/pkg/utils"

	"github.com/gofiber/fiber/v2/log"
)

type ParseDevice struct {
	ctx               context.Context
	config            *config.Config
	decryptTokenCache cache.DecryptTokenCache
	parseDataCache    cache.ParseDataCache
	devicesCache      cache.DevicesCache
	parseConnection   *conn_.ParseConnection
}

func NewParseDevice(ctx context.Context, config *config.Config, decryptTokenCache cache.DecryptTokenCache, parseDataCache cache.ParseDataCache, devicesCache cache.DevicesCache, parseConnection *conn_.ParseConnection) *ParseDevice {
	parseDevice := &ParseDevice{
		ctx:               ctx,
		config:            config,
		decryptTokenCache: decryptTokenCache,
		devicesCache:      devicesCache,
		parseDataCache:    parseDataCache,
		parseConnection:   parseConnection,
	}

	return parseDevice
}

func (s *ParseDevice) Start() {
	utils.BuildTcpServer(s.ctx, "解析模块", fmt.Sprintf(":%d", s.config.Configuration.ParsePort), s.handleConnection)
}

// handleConnection 处理每个连接
func (s *ParseDevice) handleConnection(module string, conn net.Conn) {
	addr := conn.RemoteAddr().String()
	log.Infof("[%s] 新的解析连接来自: %s", module, addr)

	// 根据连接的IP地址查找对应的设备
	parseIP := strings.Split(addr, ":")[0]
	device, ok := s.devicesCache.GetDeviceByParseIP(parseIP)
	if !ok {
		log.Warnf("[%s] 未找到匹配的设备，关闭连接: %s", module, addr)
		conn.Close()
		return
	}

	c := conn_.NewConn(conn)
	s.parseConnection.AddConnection(c)
	defer s.parseConnection.RemoveConnection(c)

	scanner := bufio.NewReader(conn)
	var buffer bytes.Buffer

	for {
		// 设置读取超时
		conn.SetReadDeadline(time.Now().Add(30 * time.Second))

		// 读取数据直到遇到换行符
		line, err := scanner.ReadBytes('\n')
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				log.Infof("[%s] 连接超时，关闭连接: %s", module, conn.RemoteAddr().String())
			} else {
				log.Errorf("[%s] 读取数据失败: %v", module, err)
			}
			break
		}

		buffer.Write(line)

		for {
			index := bytes.Index(buffer.Bytes(), []byte("\r\n"))
			if index == -1 {
				// 没有找到完整的行，继续读取
				break
			}

			// 提取完整的行
			fullLine := bytes.TrimSpace(buffer.Next(index + 2)) // 包括 \r\n

			var parseData dto.ParseData

			isHasSerial := false

			if utils.IsRID(fullLine) {
				if err := utils.ParseRID(fullLine, &parseData); err != nil {
					log.Errorf("解析 RID 数据失败: %v", err)
					continue
				}

				parseData.Device = device.ParseID

				isHasSerial = true
			} else if utils.IsEncryption(fullLine) {

				decryptToken := s.decryptTokenCache.GetDecryptToken()

				if err := utils.ParseEncryption(fullLine, &parseData, decryptToken, &isHasSerial); err != nil {
					log.Errorf("解析 Encryption 数据失败: %v", err)
					continue
				}

			} else if utils.IsDID(fullLine) {
				if err := utils.ParseDID(fullLine, &parseData); err != nil {
					log.Errorf("解析 DID 数据失败: %v", err)
					continue
				}

				if parseData.Serial != "" {
					isHasSerial = true
				}

			}

			// 这里进行报文内容校验，确保数据 hasSerial 是否存在Serial字段
			if !isHasSerial {
				log.Warnf("报文内容无效，缺少 Serial 字段，忽略该报文: %s", string(fullLine))
				continue
			}

			if parseData.Model == "" {
				continue
			}

			// gps 与解析出来的提供的都是 wgs84

			// 验证了这条告警是不是完整的
			if parseData.Serial != "" {
				// TODO: 这里可以把数据存储到数据库或者发送到消息队列
			}

			// TODO: 1、根据设置的map类型设置进行坐标转换
			// TODO: 2、去白名单查询是否在白名单内

			s.updateParseDataList(&parseData, device)
		}
	}
}

func (s *ParseDevice) updateParseDataList(data *dto.ParseData, device *models.DeviceModel) {
	// 查找符合条件的定位数据
	var parseDataIndex int = -1
	var parseData dto.ParseData
	parseDataLists := s.parseDataCache.GetParseDataList()
	for i, item := range parseDataLists {
		if item.Serial == data.Serial && item.Device == data.Device {
			parseDataIndex = i
			parseData = item
			break
		}
	}

	if parseDataIndex != -1 {
		parseData.DroneGPS = data.DroneGPS
		parseData.HomeGPS = data.HomeGPS
		parseData.PilotGPS = data.PilotGPS
		parseData.Height = data.Height
		parseData.Speed = data.Speed
		parseData.Altitude = data.Altitude
		parseData.EastV = data.EastV
		parseData.NorthV = data.NorthV
		parseData.UpV = data.UpV
		parseData.Freq = data.Freq
		parseData.RSSI = data.RSSI
		parseData.Distance = data.Distance
		parseData.Png = data.Png
		parseData.TrajectoryList = data.TrajectoryList
		parseData.InWhiteList = data.InWhiteList

		// 更新过期时间
		parseData.Expires = time.Now().Unix()
		if data.Model != "" {
			parseData.Model = data.Model
		}

		if data.DroneGPS.Longitude == 0 && data.PilotGPS.Longitude != 0 {
			parseData.DroneType = dto.DroneTypeRC
		} else if data.DroneGPS.Longitude != 0 && data.PilotGPS.Longitude == 0 {
			parseData.DroneType = dto.DroneTypeUAV
		} else if data.DroneGPS.Longitude != 0 && data.PilotGPS.Longitude != 0 {
			parseData.DroneType = dto.DroneTypeBoth
		}

		targetLat := device.Latitude
		targetLon := device.Longitude

		if targetLat != 0 && targetLon != 0 {
			distance := utils.Haversine(
				targetLat,                    // 设备纬度
				targetLon,                    // 设备经度
				parseData.DroneGPS.Latitude,  // 无人机纬度
				parseData.DroneGPS.Longitude, // 无人机经度
			)
			if targetLat != 0 && parseData.DroneGPS.Longitude > 0.1 {
				azimuth := utils.CalculateBearing(
					targetLat,                    // 设备纬度
					targetLon,                    // 设备经度
					parseData.DroneGPS.Latitude,  // 无人机纬度
					parseData.DroneGPS.Longitude, // 无人机经度
				)

				// 4. 更新 LdResult（无需方位角）
				parseData.LdResult = dto.LdResult{
					SensorId:  device.DetectionID, // 直接使用字符串类型ID
					Distance:  distance,
					Azimuth:   azimuth,
					DeviceLon: targetLon,
					DeviceLat: targetLat,
				}

			} else {
				// 如果未计算出有效的方位角
				parseData.LdResult = dto.LdResult{
					SensorId: device.DetectionID, // 直接使用字符串类型ID
					Distance: 0,
					Azimuth:  500, // 明确标注未计算方位角
				}
			}
		} else {
			parseData.LdResult = dto.LdResult{
				SensorId: device.DetectionID, // 直接使用字符串类型ID
				Distance: 0,
				Azimuth:  500, // 明确标注未计算方位角
			}
		}

		// 更新已有的定位数据
		s.parseDataCache.UpdateParseDataAtIndex(parseDataIndex, parseData)
	} else {
		// 添加新的定位数据
		s.parseDataCache.AddParseData(*data)
	}

	// 只对非DJI-Drone模型排序
	parseDataListLength := s.parseDataCache.GetParseDataListLength()
	if data.Model != "DJI-Drone" && parseDataListLength >= 2 {
		s.parseDataCache.SortParseDataListByExpires()

	}
}
