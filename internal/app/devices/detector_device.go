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
	"uav_defender/internal/pkg/config"
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
	ctx           context.Context
	devicesCache  cache.DevicesCache
	connectionCh  chan struct{} // 用于通知连接
	detectorCache cache.DetectorCache
	parseCache    cache.ParseCache
}

func NewDetectorDevice(ctx context.Context, devicesCache cache.DevicesCache, detectorCache cache.DetectorCache, parseCache cache.ParseCache) *DetectorDevice {
	return &DetectorDevice{
		ctx:           ctx,
		devicesCache:  devicesCache,
		connectionCh:  make(chan struct{}),
		detectorCache: detectorCache,
		parseCache:    parseCache,
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

	if elapsedTime > deviceHeartbeatTimeout {
		dd.handleDeviceOffline(device, elapsedTime)
	} else {
		dd.handleDeviceOnline(device)
	}
}

// handleDeviceOffline 处理设备离线逻辑
func (dd *DetectorDevice) handleDeviceOffline(device *cache.DeviceInfo, elapsedTime int64) {
	if device.DetectorFsm.Is(string(detector_fsm.StateOffline)) {
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
func (dd *DetectorDevice) handleDeviceOnline(device *cache.DeviceInfo) {
	if device.DetectorFsm.Is(string(detector_fsm.StateOffline)) {
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
		if conn.DetectorConnPool.CheckDeviceConnected(device.ID) {
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
	conn.DetectorConnPool.AddConnection(detectorConn)

	// 启动读取协程
	go dd.handleConnection(detectorConn, device)

	// 发送start命令
	conn.DetectorConnPool.SendCommandToDevice(device.ID, "start")

	// 等待设备响应
	if _, err = conn.DetectorConnPool.WaitResponseFromDevice(device.ID); err != nil {
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
func (dd *DetectorDevice) handleConnection(c conn.Conn, device *cache.DeviceInfo) {
	defer func() {
		conn.DetectorConnPool.RemoveConnection(c)
		global.Logger.Info("侦测器连接处理结束", zap.Uint("deviceID", c.GetDeviceID()))
	}()

	global.Logger.Info("处理连接，读取数据")

	handlerChan := make(chan Packet, 100)
	defer close(handlerChan)
	// 处理侧向channel
	handlerOrientationChan := make(chan struct{})
	defer close(handlerOrientationChan)

	// 启动侧向处理协程
	go dd.handleOrientationTrigger(handlerOrientationChan, device)
	// 启动数据处理协程
	go dd.handleDeviceReportData(handlerChan, handlerOrientationChan, device)

	buffer := make([]byte, 10000)

	for {
		select {
		case <-dd.ctx.Done():
			return
		default:
			// 设置读取超时
			c.GetConn().SetReadDeadline(time.Now().Add(20 * time.Second))

			n, err := c.GetConn().Read(buffer)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					global.Logger.Warn("侦测器UDP读取超时",
						zap.Uint("deviceID", c.GetDeviceID()))
				} else {
					global.Logger.Error("读取侦测器UDP数据失败",
						zap.Uint("deviceID", c.GetDeviceID()),
						zap.Error(err))
				}

				return
			}

			if n <= 0 {
				continue
			}

			message := buffer[:n]
			global.Logger.Debug("收到侦测器UDP数据",
				zap.Uint("deviceID", c.GetDeviceID()),
				zap.String("message", string(message)))

			// 处理设备响应
			if conn.DetectorConnPool.IsResponseOrReport(c.GetDeviceID(), message) {
				global.Logger.Debug("识别为命令响应",
					zap.Uint("deviceID", c.GetDeviceID()),
					zap.String("command", string(message)))

				// 处理命令响应
				select {
				case c.GetResponseChannel() <- string(message):
					conn.DetectorConnPool.ClearCommandRecord(c.GetDeviceID())
				default:
					// 通道满了，丢弃消息
					global.Logger.Warn("侦测器响应通道满，丢弃响应消息",
						zap.Uint("deviceID", c.GetDeviceID()))
				}

				continue
			}

			// 判断是否是心跳
			if conn.DetectorConnPool.IsHeartbeatData(message) {
				// 更新状态机中的心跳更新时间
				device.DetectorFsm.SetDetectorHeartbeatUpdateTime(time.Now().Unix())
				continue
			}

			// 判断是否是频谱数据
			if conn.DetectorConnPool.IsSpectrumAnalysisData(len(message)) && device.DetectorFsm.Current() == string(detector_fsm.StateSpectrumAnalyzer) {
				global.Logger.Debug("收到频谱数据包", zap.Uint("deviceID", c.GetDeviceID()), zap.Int("size", len(message)))
				// go processSpectrumData(device, buf[:n])
				continue
			}

			// 处理设备上报数据
			select {
			case handlerChan <- Packet{Conn: c, Message: message}:
			default:
				global.Logger.Debug("侦测器数据处理通道满，丢弃数据",
					zap.Uint("deviceID", c.GetDeviceID()),
					zap.String("message", string(message)))
			}

		}
	}
}

// handleDeviceReportData 处理设备上报的数据
func (dd *DetectorDevice) handleDeviceReportData(handlerChan chan Packet, handlerOrientationChan chan struct{}, device *cache.DeviceInfo) {

	for packet := range handlerChan {

		detectorData, err := dd.ParseDetectorData(packet.Message, device)
		if err != nil {
			global.Logger.Warn("解析侦测器数据失败",
				zap.Uint("deviceID", packet.Conn.GetDeviceID()),
				zap.String("data", string(packet.Message)),
				zap.Error(err))
			continue
		}

		// 判断当前报文是不是大疆无人机
		if utils.IsDJIDrone(detectorData.Model) {
			if dd.parseCache.HasDIDData() {
				global.Logger.Debug("当前报文为大疆无人机，且已存在DID数据，跳过本次报文处理",
					zap.Uint("deviceID", packet.Conn.GetDeviceID()),
					zap.String("data", string(packet.Message)),
					zap.Any("parsedData", detectorData))
				continue
			}
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

		// 判断是否已关闭，避免向已关闭的channel发送数据
		select {
		case handlerOrientationChan <- struct{}{}:
			global.Logger.Debug("触发定向处理", zap.Uint("deviceID", device.ID))
		case <-dd.ctx.Done():
			return
		default:
			// channel已满或已关闭，记录日志但不阻塞
			global.Logger.Debug("定向处理通道忙碌，跳过本次触发",
				zap.Uint("deviceID", device.ID))
		}
	}
}

// handleOrientationTrigger 触发处理侧向
func (dd *DetectorDevice) handleOrientationTrigger(handlerOrientationChan chan struct{}, device *cache.DeviceInfo) {
	for range handlerOrientationChan {
		if device.DetectorFsm.Is(string(detector_fsm.StateDirection)) {
			continue
		}

		global.Logger.Info("侦测器设备状态机已切换到定向侦测状态", zap.Uint("deviceID", device.ID))

		targets := dd.detectorCache.GetDirectionalDataTargets()

		global.Logger.Info("获取到定向侦测目标", zap.Int("targetCount", len(targets)), zap.Uint("deviceID", device.ID))

		for _, tgt := range targets {
			// 检查当前状态是否仍然是定向侦测
			if device.DetectorFsm.Is(string(detector_fsm.StateOffline)) {
				global.Logger.Warn("状态已切换，停止当前定向侦测操作")
				break
			}

			// 只有当前状态是全向侦测，才切换到定向侦测
			if device.DetectorFsm.Is(string(detector_fsm.StateOmniDirection)) {
				if err := device.DetectorFsm.Event(dd.ctx, string(detector_fsm.EventToDirect), device); err != nil {
					global.Logger.Error("侦测器设备状态机切换到定向侦测状态失败",
						zap.Uint("deviceID", device.ID),
						zap.Error(err))
					continue
				}
			}

			global.Logger.Info("开始对目标频点进行定向侦测", zap.Float64("freq", tgt.Freq), zap.Uint("deviceID", device.ID))

			// 	// // 清空条件成立的对应频点的GPIO数据
			// 	// utils.ClearGpioDataForTargets(targets)
			// TODO: 这里可以添加一些逻辑来避免频繁对同一频点进行定向侦测

			// 构建定向侦测命令
			cmd := utils.BuildDirectionDetectionCommand(tgt.Freq)
			if err := conn.DetectorConnPool.SendCommandToDevice(device.ID, cmd); err != nil {
				global.Logger.Warn("发送定向侦测命令失败", zap.Error(err))
				break
			}

			// 等待设备响应
			if _, err := conn.DetectorConnPool.WaitResponseFromDevice(device.ID); err != nil {
				global.Logger.Warn("等待定向侦测命令响应失败", zap.Error(err))
				break
			}

			global.Logger.Info("发送定向侦测命令成功", zap.Float64("freq", tgt.Freq), zap.Uint("deviceID", device.ID))

			// 等待一段时间，确保定向侦测完成
			time.Sleep(time.Duration(config.AppConfig.Configuration.LockFrequency) * time.Second)

			if err := dd.detectorCache.UpdateOrientationByFreq(tgt.Freq); err != nil {
				continue
			}

			global.Logger.Info("定向侦测完成，更新方向成功", zap.Float64("freq", tgt.Freq), zap.Uint("deviceID", device.ID))
		}

		// 定向侦测完成，切换回全向侦测状态
		// 判断当前状态 不是定向侦测状态 则跳过
		if !device.DetectorFsm.Is(string(detector_fsm.StateDirection)) {
			global.Logger.Warn("状态已切换，跳过切换回全向侦测操作")
			continue
		}
		offCmd := utils.BuildStopDirectionDetectionCommand()
		if err := conn.DetectorConnPool.SendCommandToDevice(device.ID, offCmd); err != nil {
			global.Logger.Warn("发送停止定向侦测命令失败", zap.Error(err))
			continue
		}

		// 等待设备响应
		if _, err := conn.DetectorConnPool.WaitResponseFromDevice(device.ID); err != nil {
			global.Logger.Warn("等待停止定向侦测命令响应失败", zap.Error(err))
			continue
		}

		if err := device.DetectorFsm.Event(dd.ctx, string(detector_fsm.EventToOmni)); err != nil {
			global.Logger.Error("侦测器设备状态机切换到全向侦测状态失败",
				zap.Uint("deviceID", device.ID),
				zap.Error(err))
			continue
		}
		global.Logger.Info("定向侦测完成，已切换回全向侦测状态", zap.Uint("deviceID", device.ID))

		time.Sleep(10 * time.Second)
	}
}

// ParseDetectorData 解析侦测器数据
func (dd *DetectorDevice) ParseDetectorData(data []byte, device *cache.DeviceInfo) (*dto.DetectorData, error) {
	if !utils.IsValidDetectorData(data) {
		global.Logger.Warn("无效的侦测器数据", zap.String("data", string(data)))
		return nil, fmt.Errorf("无效的侦测器数据")
	}

	detectorData := &dto.DetectorData{}

	detectorData.DeviceID = device.ID

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

	// DetectionID
	if device, ok := fieldMap["device"]; ok {
		id, err := strconv.ParseUint(string(device), 10, 32)
		if err != nil {
			return nil, fmt.Errorf("解析device字段失败: %w", err)
		}
		detectorData.DetectionID = uint(id)
	}

	// Model 和 UAV
	if model, ok := fieldMap["model"]; ok {
		detectorData.Model = string(model)
		// 根据model字段映射无人机型号
		if droneModel, found := utils.GetDroneModelByModelSource(detectorData.Model); found {
			detectorData.UAV = droneModel
		} else {
			detectorData.UAV = detectorData.Model
		}
	}

	// Freq
	if freq, ok := fieldMap["freq"]; ok {
		if f, err := strconv.ParseFloat(string(freq), 64); err == nil {
			detectorData.Freq = f
		}
	}

	// RSSI
	if rssi, ok := fieldMap["rssi"]; ok {
		if f, err := strconv.ParseFloat(string(rssi), 64); err == nil {
			detectorData.RSSI = f
		}
	}

	// Seq
	if seq, ok := fieldMap["seq"]; ok {
		if s, err := strconv.ParseInt(string(seq), 10, 64); err == nil {
			detectorData.Seq = s
		}
	}

	// Gpio
	if gpio, ok := fieldMap["gpio"]; ok {
		if g, err := strconv.ParseInt(string(gpio), 10, 64); err == nil {
			detectorData.Gpio = g
		}
	}

	detectorData.LastTime = models.CustomTime(time.Now())
	detectorData.ID = utils.GenerateRandomID(8)
	// detectorData.GpioS = []int64{}
	detectorData.Orientation = 500
	// detectorData.OrientationTS = ??
	// detectorData.TS = ??
	detectorData.IsCracked = false
	detectorData.FirstSeen = models.CustomTime(time.Now()) // 入库时入侵时间要用
	// detectorData.GpiosData = [8]dto.GpioData{}

	return detectorData, nil
}

// ClearGpioDataForTargets 清空条件成立的对应频点的GPIO数据
// func ClearGpioDataForTargets(targets []*dto.Alert) {

// 	// cache.DroneTargetAlertsLock.RLock()
// 	// if cache.LateralStrategy == 1 {
// 	// 	// cache.LateralStrategy == 1  条件成立清空对应频点的 GPIO 数据（条件成立时）
// 	// 	for i := range cache.DroneTargetAlerts {
// 	// 		// 将筛选出来的目标对象，与实际缓存中的目标对象进行匹配，如果Freq相同 GpiosData清空
// 	// 		if cache.DroneTargetAlerts[i].Freq == tgt.Freq {
// 	// 			cache.DroneTargetAlerts[i].GpiosData = [8]dto.GpioData{}
// 	// 		}
// 	// 	}
// 	// }
// 	// cache.DroneTargetAlertsLock.RUnlock()

// 	// 先判断 cache.LateralStrategy 是否为1
// 	if cache.LateralStrategy != 1 {
// 		return
// 	}

// 	// 在无锁状态下构建需要清空的频率集合，减少持锁时间并避免在持锁时调用外部代码
// 	freqSet := make(map[float64]struct{}, len(targets))
// 	for _, tgt := range targets {
// 		if tgt == nil {
// 			continue
// 		}
// 		freqSet[tgt.Freq] = struct{}{}
// 	}

// 	cache.DroneTargetAlertsLock.Lock()
// 	defer cache.DroneTargetAlertsLock.Unlock()

// 	for _, alert := range cache.DroneTargetAlerts {
// 		if alert == nil {
// 			continue
// 		}
// 		if _, ok := freqSet[alert.Freq]; ok {
// 			alert.GpiosData = [8]dto.GpioData{}
// 		}
// 	}
// }
