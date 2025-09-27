package devices

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net"
	"strings"
	"time"
	conn_ "uav_defender/internal/app/devices/conn"
	parse_fsm "uav_defender/internal/app/devices/fms/parse"
	"uav_defender/internal/cache"
	"uav_defender/internal/dto"
	"uav_defender/internal/models"
	"uav_defender/internal/pkg/config"
	"uav_defender/internal/pkg/global"
	"uav_defender/internal/pkg/utils"
	"uav_defender/internal/services"

	"go.uber.org/zap"
)

type ParseDevice struct {
	ctx context.Context

	decryptTokenCache cache.DecryptTokenCache
	parseCache        cache.ParseCache
	devicesCache      cache.DevicesCache

	parseConnection *conn_.ParseConnection

	whitelistService services.WhitelistService
}

func NewParseDevice(ctx context.Context, decryptTokenCache cache.DecryptTokenCache, parseDataCache cache.ParseCache, devicesCache cache.DevicesCache, parseConnection *conn_.ParseConnection, whitelistService services.WhitelistService) *ParseDevice {
	parseDevice := &ParseDevice{
		ctx:               ctx,
		decryptTokenCache: decryptTokenCache,
		devicesCache:      devicesCache,
		parseCache:        parseDataCache,
		parseConnection:   parseConnection,
		whitelistService:  whitelistService,
	}

	return parseDevice
}

func (s *ParseDevice) Start() {
	utils.BuildTcpServer(s.ctx, "解析模块", fmt.Sprintf(":%d", config.AppConfig.Configuration.ParsePort), s.handleConnection)
}

// handleConnection 处理每个连接
func (s *ParseDevice) handleConnection(module string, conn net.Conn) {
	defer conn.Close()
	addr := conn.RemoteAddr().String()
	global.Logger.Info("新的解析连接来自", zap.String("module", module), zap.String("address", addr))

	// 根据连接的IP地址查找对应的设备
	parseIP := strings.Split(addr, ":")[0]
	device, ok := s.devicesCache.GetDeviceByParseIP(parseIP)
	if !ok {
		global.Logger.Warn("未找到匹配的设备，关闭连接", zap.String("module", module), zap.String("address", addr))
		return
	}

	// 如果状态机在离线状态，尝试切换到在线状态
	if device.ParseFsm.FSM.Is(string(parse_fsm.StateOffline)) {
		if err := device.ParseFsm.FSM.Event(s.ctx, string(parse_fsm.EventToOnline)); err != nil {
			global.Logger.Error("状态机切换到在线状态失败，关闭连接", zap.String("module", module), zap.String("address", addr), zap.Error(err))
			return
		}
	}
	defer func() {
		// 连接关闭时，切换状态机到离线状态
		if !device.ParseFsm.FSM.Is(string(parse_fsm.StateOffline)) {
			device.ParseFsm.FSM.Event(s.ctx, string(parse_fsm.EventToOffline))
		}
	}()

	c := conn_.NewConn(device.ID, conn)
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
				global.Logger.Info("连接超时，关闭连接", zap.String("module", module), zap.String("address", addr))
			} else {
				global.Logger.Warn("读取数据失败", zap.String("module", module), zap.Error(err))
			}
			break
		}

		// 判断 buffer 是否超过 10kB，防止内存耗尽攻击
		if buffer.Len() > 10*1024 {
			global.Logger.Warn("缓冲区数据过大，关闭连接", zap.String("module", module), zap.String("address", addr))
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

			parseData, err := s.parseParseData(fullLine, device)
			if err != nil {
				global.Logger.Warn("解析数据失败", zap.String("module", module), zap.String("address", addr), zap.Error(err))
				continue
			}
			s.updateParseDataList(*parseData, device)
		}
	}
}

// updateParseDataList 更新解析数据列表
func (s *ParseDevice) updateParseDataList(newParseData dto.ParseData, device *cache.DeviceInfo) {
	// 查找符合条件的定位数据
	var parseDataIndex int = -1
	var parseData dto.ParseData
	parseDataLists := s.parseCache.GetParseDataList()
	for i, item := range parseDataLists {
		if item.Serial == newParseData.Serial {
			parseDataIndex = i
			parseData = item
			break
		}
	}

	if parseDataIndex != -1 {

		if parseData.ParseID == 0 {
			if newParseData.ParseID != 0 {
				parseData.ParseID = newParseData.ParseID
			}
		}

		parseData.DroneGPS = newParseData.DroneGPS
		parseData.HomeGPS = newParseData.HomeGPS
		parseData.PilotGPS = newParseData.PilotGPS
		parseData.Height = newParseData.Height
		parseData.Speed = newParseData.Speed
		parseData.Altitude = newParseData.Altitude
		parseData.EastV = newParseData.EastV
		parseData.NorthV = newParseData.NorthV
		parseData.UpV = newParseData.UpV
		parseData.Freq = newParseData.Freq
		parseData.RSSI = newParseData.RSSI
		parseData.Distance = newParseData.Distance
		parseData.Png = newParseData.Png
		parseData.HasInWhiteList = newParseData.HasInWhiteList

		// 这是新增的更新字段
		parseData.Mac = newParseData.Mac
		parseData.Sign = newParseData.Sign
		parseData.TargetId = newParseData.TargetId
		// parseData.MType = newParseData.MType
		parseData.Serial = newParseData.Serial

		now := models.CustomTime(time.Now())

		// 更新过期时间
		parseData.Expires = now

		parseData.Model = newParseData.Model
		parseData.DroneType = newParseData.DroneType

		isValidDroneGPS := utils.IsValidCoord(parseData.DroneGPS.Longitude, parseData.DroneGPS.Latitude)

		if isValidDroneGPS {
			// 检查是否需要添加新轨迹点
			newPoint := models.Trajectory{
				Latitude:  newParseData.DroneGPS.Latitude,
				Longitude: newParseData.DroneGPS.Longitude,
				Height:    newParseData.Height,
			}
			if len(parseData.Trajectories) == 0 {
				parseData.Trajectories = models.Trajectories{newPoint}
			} else {
				lastTrajectory := parseData.Trajectories[len(parseData.Trajectories)-1]
				// 只有当新点与最后一个点不同才添加，避免重复点
				if newPoint.Latitude != lastTrajectory.Latitude || newPoint.Longitude != lastTrajectory.Longitude {
					parseData.Trajectories = append(parseData.Trajectories, newPoint)
				}
			}

		}

		// 更新已有的定位数据
		s.parseCache.UpdateParseDataAtIndex(parseDataIndex, parseData)
	} else {
		// 添加新的定位数据
		s.parseCache.AddParseData(newParseData)
	}

	// 只对非DJI-Drone模型排序
	parseDataListLength := s.parseCache.GetParseDataListLength()
	if parseDataListLength > 1 && newParseData.Model != "DJI-Drone" {
		s.parseCache.SortParseDataListByExpires()
	}
}

// parseParseData
func (s *ParseDevice) parseParseData(message []byte, device *cache.DeviceInfo) (*dto.ParseData, error) {

	var parseData dto.ParseData
	parseData.DeviceID = device.ID

	if utils.IsRID(message) {
		utils.ParseRID(message, &parseData)

		parseData.ParseID = device.ParseID

		parseData.Sign = dto.SignTypeO3Plus
	} else if utils.IsEncryption(message) {
		decryptToken := s.decryptTokenCache.GetDecryptToken()
		if err := utils.ParseEncryption(message, &parseData, decryptToken); err != nil {
			global.Logger.Error("解析加密报文失败", zap.Error(err))
			return nil, err
		}

		parseData.Sign = dto.SignTypeO2O3

	} else if utils.IsDID(message) {
		utils.ParseDID(message, &parseData)
		parseData.Sign = dto.SignTypeO2O3
	}

	if parseData.Model == "" || parseData.Serial == "" {
		return nil, fmt.Errorf("报文内容无效，缺少 Model 或 Serial 字段")
	}

	if parseData.Sign == 1 {
		// serial 增加前缀 1581
		if !strings.HasPrefix(parseData.Serial, "1581") {
			parseData.Serial = "1581" + parseData.Serial
			parseData.TargetId = parseData.Serial
		}

		// model中没有dji字符串，在model的最前面加上DJI和一个空格
		if !utils.IsDJIDrone(parseData.Model) {
			parseData.Model = "DJI " + parseData.Model
		}
	}

	// 目标ID
	parseData.TargetId = parseData.Serial

	// 解析无人机类型
	utils.ParseDroneType(&parseData)

	now := models.CustomTime(time.Now())
	// 更新过期时间
	parseData.Expires = now

	// 记录入侵时间
	parseData.IntrusionTime = now

	// 去白名单查询是否在白名单内
	hasInWhiteList, err := s.whitelistService.IsSerialWhitelisted(parseData.Serial)
	if err != nil {
		global.Logger.Error("查询白名单失败", zap.String("serial", parseData.Serial), zap.Error(err))
	}

	parseData.HasInWhiteList = hasInWhiteList

	isValidPilotGPS := utils.IsValidCoord(parseData.PilotGPS.Longitude, parseData.PilotGPS.Latitude)

	// 判断飞手经纬度是否有效
	if isValidPilotGPS {
		parseData.Png, err = utils.GenerateQRCodeBase64(parseData.PilotGPS.Longitude, parseData.PilotGPS.Latitude)
		if err != nil {
			global.Logger.Error("生成飞手位置二维码失败", zap.String("serial", parseData.Serial), zap.Error(err))
		}
	}

	isValidDeviceGPS := utils.IsValidCoordPtr(device.Longitude, device.Latitude)

	isValidDroneGPS := utils.IsValidCoord(parseData.DroneGPS.Longitude, parseData.DroneGPS.Latitude)

	// 判断设备是否配置了经纬度并且设备经纬度和无人机经纬度是否有效
	if isValidDeviceGPS && isValidDroneGPS {
		targetGPS := dto.GPS{
			Latitude:  *device.Latitude,
			Longitude: *device.Longitude,
		}

		// 计算距离
		distance := utils.Distance(targetGPS, parseData.DroneGPS)

		// 计算方位角
		bearing := utils.Bearing(targetGPS, parseData.DroneGPS)

		// 4. 更新 LdResult（无需方位角）
		parseData.LdResult = dto.LdResult{
			Azimuth:         bearing, // 明确标注未计算方位角
			Distance:        distance,
			DetectionID:     device.DetectionID,
			Orientation:     bearing,
			DeviceLongitude: targetGPS.Longitude,
			DeviceLatitude:  targetGPS.Latitude,
			Height:          parseData.Height,
		}

	} else {
		// 设备未配置经纬度或设备和无人机经纬度无效，LdResult 字段置为默认值
		global.Logger.Warn("设备或无人机经纬度无效，无法计算距离和方位角", zap.Uint("device_id", device.DetectionID), zap.String("device_model", device.Name), zap.Float64("device_longitude", *device.Longitude), zap.Float64("device_latitude", *device.Latitude), zap.Float64("drone_longitude", parseData.DroneGPS.Longitude), zap.Float64("drone_latitude", parseData.DroneGPS.Latitude))
		parseData.LdResult = dto.LdResult{
			Azimuth:         500, // 明确标注未计算方位角
			Distance:        0,
			DetectionID:     device.DetectionID,
			Orientation:     500,
			DeviceLongitude: 0,
			DeviceLatitude:  0,
			Height:          0,
		}
	}

	return &parseData, nil
}
