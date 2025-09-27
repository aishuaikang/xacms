package devices

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
	"uav_defender/internal/app/devices/conn"
	detector_fsm "uav_defender/internal/app/devices/fms/detector"
	"uav_defender/internal/cache"
	"uav_defender/internal/dto"
	"uav_defender/internal/models"
	"uav_defender/internal/pkg/global"
	"uav_defender/internal/pkg/utils"

	"go.uber.org/zap"
)

const (
	// 设备连接检查间隔
	deviceConnectionInterval = 20 * time.Second
	// 设备心跳检查间隔
	deviceHeartbeatInterval = 2 * time.Second
	// 设备心跳超时时间
	deviceHeartbeatTimeout = 20
)

type DetectorDevice struct {
	ctx                context.Context
	devicesCache       cache.DevicesCache
	connectionCh       chan struct{} // 用于通知连接
	detectorConnection *conn.DetectorConnection
	detectorCache      cache.DetectorCache
}

func NewDetectorDevice(ctx context.Context, devicesCache cache.DevicesCache, detectorConnection *conn.DetectorConnection, detectorCache cache.DetectorCache) *DetectorDevice {
	return &DetectorDevice{
		ctx:                ctx,
		devicesCache:       devicesCache,
		connectionCh:       make(chan struct{}),
		detectorConnection: detectorConnection,
		detectorCache:      detectorCache,
	}
}

func (dd *DetectorDevice) Start() {
	// 启动侦测设备连接协程
	go dd.startConnectionMonitor()

	// 启动侦测设备状态机监控协程
	go dd.startHeartbeatMonitor()
}

// startConnectionMonitor 启动设备连接监控
func (dd *DetectorDevice) startConnectionMonitor() {
	defer global.Logger.Info("侦测器设备连接监控协程停止")

	// 启动时先尝试连接所有设备
	dd.connectAllDevices("启动时连接")

	ticker := time.NewTicker(deviceConnectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-dd.ctx.Done():
			return
		case <-dd.connectionCh:
			global.Logger.Debug("接收到连接信号，开始连接设备")
			dd.connectAllDevices("手动触发连接")
		case <-ticker.C:
			global.Logger.Debug("定时器触发，检查设备连接状态")
			dd.connectAllDevices("定时检查连接")
		}
	}
}

// startHeartbeatMonitor 启动设备心跳监控
func (dd *DetectorDevice) startHeartbeatMonitor() {
	defer global.Logger.Info("侦测器设备状态机监控协程停止")

	ticker := time.NewTicker(deviceHeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-dd.ctx.Done():
			return
		case <-ticker.C:
			dd.checkDeviceHeartbeats()
		}
	}
}

// checkDeviceHeartbeats 检查所有设备的心跳状态
func (dd *DetectorDevice) checkDeviceHeartbeats() {
	devices := dd.devicesCache.GetDevices()
	currentTime := time.Now().Unix()

	for _, device := range devices {
		dd.checkSingleDeviceHeartbeat(&device, currentTime)
	}
}

// checkSingleDeviceHeartbeat 检查单个设备的心跳状态
func (dd *DetectorDevice) checkSingleDeviceHeartbeat(device *cache.DeviceInfo, currentTime int64) {
	lastHeartbeat := device.DetectorFsm.GetDetectorHeartbeatUpdateTime()

	// 如果从未收到心跳，不进行离线判断
	if lastHeartbeat == 0 {
		return
	}

	elapsedTime := currentTime - lastHeartbeat
	isOffline := device.DetectorFsm.Is(string(detector_fsm.StateOffline))

	if elapsedTime > deviceHeartbeatTimeout {
		dd.handleDeviceOffline(device, elapsedTime, isOffline)
	} else {
		dd.handleDeviceOnline(device, isOffline)
	}
}

// handleDeviceOffline 处理设备离线逻辑
func (dd *DetectorDevice) handleDeviceOffline(device *cache.DeviceInfo, elapsedTime int64, isOffline bool) {
	// 如果已经是离线状态，无需重复处理
	if isOffline {
		return
	}

	// 切换到离线状态
	if err := device.DetectorFsm.Event(dd.ctx, string(detector_fsm.EventToOffline)); err != nil {
		global.Logger.Error("侦测器设备状态机切换到离线状态失败",
			zap.Uint("deviceID", device.ID),
			zap.Error(err))
		return
	}

	global.Logger.Warn("侦测器设备心跳超时，已切换至离线状态",
		zap.Uint("deviceID", device.ID),
		zap.Int64("elapsedTime", elapsedTime))
}

// handleDeviceOnline 处理设备在线逻辑
func (dd *DetectorDevice) handleDeviceOnline(device *cache.DeviceInfo, isOffline bool) {
	// 如果不是离线状态，无需处理
	if !isOffline {
		return
	}

	// 切换到全向侦测状态
	if err := device.DetectorFsm.Event(dd.ctx, string(detector_fsm.EventToOmni)); err != nil {
		global.Logger.Error("侦测器设备状态机切换到全向侦测状态失败",
			zap.Uint("deviceID", device.ID),
			zap.Error(err))
		return
	}

	global.Logger.Info("侦测器设备心跳恢复，已切换至全向侦测状态",
		zap.Uint("deviceID", device.ID))
}

// connectAllDevices 连接所有未连接的设备
func (dd *DetectorDevice) connectAllDevices(reason string) {
	devices := dd.devicesCache.GetDevices()
	if len(devices) == 0 {
		global.Logger.Debug("没有可连接的设备", zap.String("reason", reason))
		return
	}

	global.Logger.Info("开始连接设备",
		zap.String("reason", reason),
		zap.Int("totalDevices", len(devices)))

	connectedCount := 0
	failedCount := 0

	for _, device := range devices {
		// 跳过已连接的设备
		if dd.detectorConnection.CheckDeviceConnected(device.ID) {
			connectedCount++
			continue
		}

		// 尝试连接设备
		if err := dd.ConnectToDetector(&device, device.DetectionIP, device.DetectionPort); err != nil {
			failedCount++
			global.Logger.Error("连接侦测器设备失败",
				zap.Uint("deviceID", device.ID),
				zap.String("address", fmt.Sprintf("%s:%d", device.DetectionIP, device.DetectionPort)),
				zap.String("reason", reason),
				zap.Error(err))
		} else {
			global.Logger.Info("成功连接侦测器设备",
				zap.Uint("deviceID", device.ID),
				zap.String("address", fmt.Sprintf("%s:%d", device.DetectionIP, device.DetectionPort)),
				zap.String("reason", reason))
		}
	}

	global.Logger.Info("设备连接完成",
		zap.String("reason", reason),
		zap.Int("totalDevices", len(devices)),
		zap.Int("alreadyConnected", connectedCount),
		zap.Int("failed", failedCount))
}

// TriggerConnection 手动触发设备连接
func (dd *DetectorDevice) TriggerConnection() {
	select {
	case dd.connectionCh <- struct{}{}:
		global.Logger.Debug("成功发送连接触发信号")
	default:
		global.Logger.Warn("连接触发信号通道满，忽略本次触发")
	}
}

// ConnectToDetector UDP连接到侦测器设备
func (dd *DetectorDevice) ConnectToDetector(device *cache.DeviceInfo, detectionIP string, detectionPort uint) error {
	var address string
	if net.ParseIP(detectionIP) != nil && net.ParseIP(detectionIP).To4() == nil {
		// IPv6 address
		address = fmt.Sprintf("[%s]:%d", detectionIP, detectionPort)
	} else {
		// IPv4 address or hostname
		address = fmt.Sprintf("%s:%d", detectionIP, detectionPort)
	}

	// UDP连接
	c, err := net.DialTimeout("udp", address, 10*time.Second)
	if err != nil {
		global.Logger.Error("UDP连接侦测器设备失败",
			zap.Uint("deviceID", device.ID),
			zap.String("address", address),
			zap.Error(err))
		return fmt.Errorf("UDP连接侦测器设备失败: %w", err)
	}

	// 创建连接包装器
	detectorConn := conn.NewConn(device.ID, c)

	// 添加到连接池
	dd.detectorConnection.AddConnection(detectorConn)

	// 启动读取协程
	go dd.handleConnection(detectorConn, device)

	// 发送start命令
	dd.detectorConnection.SendCommandToDevice(device.ID, "start")

	// 等待设备响应
	if _, err = dd.detectorConnection.WaitResponseFromDevice(device.ID); err != nil {
		return fmt.Errorf("等待设备响应失败: %w", err)
	}

	global.Logger.Info("成功连接到侦测器设备",
		zap.Uint("deviceID", device.ID),
		zap.String("address", address))

	return nil
}

type Packet struct {
	Conn    conn.Conn
	Message []byte
}

// handleConnection 处理连接，读取数据
func (dd *DetectorDevice) handleConnection(conn conn.Conn, device *cache.DeviceInfo) {
	defer func() {
		dd.detectorConnection.RemoveConnection(conn)
		global.Logger.Info("侦测器连接处理结束", zap.Uint("deviceID", conn.GetDeviceID()))
	}()

	handlerChan := make(chan Packet, 300)
	defer close(handlerChan)

	// 启动数据处理协程
	go dd.handleDeviceReportData(handlerChan, device)

	buffer := make([]byte, 10000)

	for {
		select {
		case <-dd.ctx.Done():
			return
		default:
			// 设置读取超时
			conn.GetConn().SetReadDeadline(time.Now().Add(20 * time.Second))

			n, err := conn.GetConn().Read(buffer)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					global.Logger.Warn("侦测器UDP读取超时",
						zap.Uint("deviceID", conn.GetDeviceID()))
				} else {
					global.Logger.Error("读取侦测器UDP数据失败",
						zap.Uint("deviceID", conn.GetDeviceID()),
						zap.Error(err))
				}

				return
			}

			if n <= 0 {
				continue
			}

			message := buffer[:n]
			global.Logger.Debug("收到侦测器UDP数据",
				zap.Uint("deviceID", conn.GetDeviceID()),
				zap.String("message", string(message)))

			// 处理设备上报数据
			select {
			case handlerChan <- Packet{Conn: conn, Message: message}:
			default:
				global.Logger.Debug("侦测器数据处理通道满，丢弃数据",
					zap.Uint("deviceID", conn.GetDeviceID()),
					zap.String("message", string(message)))
			}

		}
	}
}

// handleDeviceReportData 处理设备上报的数据
func (dd *DetectorDevice) handleDeviceReportData(handlerChan chan Packet, device *cache.DeviceInfo) {
	for packet := range handlerChan {
		// 处理设备响应
		if dd.detectorConnection.IsResponseOrReport(packet.Conn.GetDeviceID(), packet.Message) {
			global.Logger.Debug("识别为命令响应",
				zap.Uint("deviceID", packet.Conn.GetDeviceID()),
				zap.String("command", string(packet.Message)))

			// 处理命令响应
			select {
			case packet.Conn.GetResponseChannel() <- string(packet.Message):
				dd.detectorConnection.ClearCommandRecord(packet.Conn.GetDeviceID())
			default:
				// 通道满了，丢弃消息
				global.Logger.Warn("侦测器响应通道满，丢弃响应消息",
					zap.Uint("deviceID", packet.Conn.GetDeviceID()))
			}

			continue
		}

		// 判断是否是心跳
		if dd.detectorConnection.IsHeartbeatData(packet.Message) {
			// 更新状态机中的心跳更新时间
			device.DetectorFsm.SetDetectorHeartbeatUpdateTime(time.Now().Unix())
			continue
		}

		// 判断是否是频谱数据
		if dd.detectorConnection.IsSpectrumAnalysisData(len(packet.Message)) && device.DetectorFsm.Current() == string(detector_fsm.StateSpectrumAnalyzer) {
			global.Logger.Debug("收到频谱数据包", zap.Uint("deviceID", packet.Conn.GetDeviceID()), zap.Int("size", len(packet.Message)))
			// go processSpectrumData(device, buf[:n])
			continue
		}

		detectorData, err := dd.ParseDetectorData(packet.Message)
		if err != nil {
			global.Logger.Warn("解析侦测器数据失败",
				zap.Uint("deviceID", packet.Conn.GetDeviceID()),
				zap.String("data", string(packet.Message)),
				zap.Error(err))
			continue
		}

		global.Logger.Debug("解析侦测器数据成功",
			zap.Uint("deviceID", packet.Conn.GetDeviceID()),
			zap.String("data", string(packet.Message)),
			zap.Any("parsedData", detectorData))

		// 更新缓存
		if err := dd.detectorCache.UpdateDetectorDataList(*detectorData); err != nil {
			global.Logger.Error("更新侦测器数据缓存失败",
				zap.Uint("deviceID", packet.Conn.GetDeviceID()),
				zap.Error(err))
		}
	}
}

func (dd *DetectorDevice) ParseDetectorData(data []byte) (*dto.DetectorData, error) {
	if !utils.IsValidDetectorData(data) {
		global.Logger.Warn("无效的侦测器数据", zap.String("data", string(data)))
		return nil, fmt.Errorf("无效的侦测器数据")
	}

	detectorData := &dto.DetectorData{}

	// 预分配map提高性能
	fieldMap := make(map[string][]byte, 8)

	// 解析键值对
	for _, field := range bytes.Split(data, []byte{','}) {
		kv := bytes.SplitN(bytes.TrimSpace(field), []byte{'='}, 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.ToLower(string(bytes.TrimSpace(kv[0])))
		value := bytes.TrimSpace(kv[1])
		fieldMap[key] = value
	}

	// 批量处理字段
	if device, ok := fieldMap["device"]; ok {
		id, err := strconv.ParseUint(string(device), 10, 32)
		if err != nil {
			return nil, fmt.Errorf("解析device字段失败: %w", err)
		}
		detectorData.DetectionID = uint(id)
	}

	if model, ok := fieldMap["model"]; ok {
		detectorData.Model = string(model)
		// 根据model字段映射无人机型号
		if droneModel, found := utils.GetDroneModelByModelSource(detectorData.Model); found {
			detectorData.UAV = droneModel
		} else {
			detectorData.UAV = detectorData.Model
		}
	}

	if freq, ok := fieldMap["freq"]; ok {
		if f, err := strconv.ParseFloat(string(freq), 64); err == nil {
			detectorData.Freq = f
		}
	}

	if rssi, ok := fieldMap["rssi"]; ok {
		if f, err := strconv.ParseFloat(string(rssi), 64); err == nil {
			detectorData.RSSI = f
		}
	}

	if seq, ok := fieldMap["seq"]; ok {
		if s, err := strconv.ParseInt(string(seq), 10, 64); err == nil {
			detectorData.Seq = s
		}
	}

	if gpio, ok := fieldMap["gpio"]; ok {
		if g, err := strconv.ParseInt(string(gpio), 10, 64); err == nil {
			detectorData.Gpio = g
		}
	}

	detectorData.LastTime = models.CustomTime(time.Now())
	detectorData.ID = utils.GenerateRandomID(8)
	detectorData.Orientation = 500
	detectorData.FirstSeen = models.CustomTime(time.Now()) // 入库时入侵时间要用

	return detectorData, nil
}
