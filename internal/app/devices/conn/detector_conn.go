package conn

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"uav_defender/internal/pkg/global"

	"go.uber.org/zap"
)

// DetectorConnection 检测器UDP连接池
type DetectorConnection struct {
	Connections      map[uint]Conn      // UDP连接映射，key为设备ID
	ConnectionsMutex sync.RWMutex       // 读写锁保护连接映射
	LastCommands     map[uint]string    // 记录每个设备最后发送的命令，用于区分响应和上报数据
	CommandsMutex    sync.RWMutex       // 保护LastCommands的读写锁
	ctx              context.Context    // 上下文
	cancel           context.CancelFunc // 取消函数
}

// NewDetectorConnection 创建新的检测器UDP连接池
func NewDetectorConnection() *DetectorConnection {
	ctx, cancel := context.WithCancel(context.Background())
	return &DetectorConnection{
		Connections:  make(map[uint]Conn),
		LastCommands: make(map[uint]string),
		ctx:          ctx,
		cancel:       cancel,
	}
}

// GetAllConnections 获取所有连接
func (dc *DetectorConnection) GetAllConnections() map[uint]Conn {
	dc.ConnectionsMutex.RLock()
	defer dc.ConnectionsMutex.RUnlock()

	// 返回连接的副本
	connections := make(map[uint]Conn)
	for id, conn := range dc.Connections {
		connections[id] = conn
	}
	return connections
}

// AddConnection 添加新的连接
func (dc *DetectorConnection) AddConnection(conn Conn) {
	dc.ConnectionsMutex.Lock()
	defer dc.ConnectionsMutex.Unlock()

	deviceID := conn.GetDeviceID()

	// 如果已存在连接，先关闭旧连接
	if oldConn, ok := dc.Connections[deviceID]; ok {
		global.Logger.Info("关闭旧的检测器UDP连接", zap.Uint("deviceID", deviceID))
		oldConn.Close()
		oldConn.CloseResponseChannel()
	}

	dc.Connections[deviceID] = conn
	global.Logger.Info("添加检测器UDP连接到连接池", zap.Uint("deviceID", deviceID))
}

// RemoveConnection 移除连接
func (dc *DetectorConnection) RemoveConnection(conn Conn) {
	dc.ConnectionsMutex.Lock()
	dc.CommandsMutex.Lock()
	defer dc.ConnectionsMutex.Unlock()
	defer dc.CommandsMutex.Unlock()

	deviceID := conn.GetDeviceID()
	if _, ok := dc.Connections[deviceID]; ok {
		conn.Close()
		conn.CloseResponseChannel()
		delete(dc.Connections, deviceID)
		// 同时清理命令记录
		delete(dc.LastCommands, deviceID)
		global.Logger.Info("从连接池移除检测器UDP连接", zap.Uint("deviceID", deviceID))
	}
}

// GetConnection 获取指定设备的连接
func (dc *DetectorConnection) GetConnection(deviceID uint) (Conn, bool) {
	dc.ConnectionsMutex.RLock()
	defer dc.ConnectionsMutex.RUnlock()

	conn, ok := dc.Connections[deviceID]
	return conn, ok
}

// SendCommandToDevice 向指定设备发送命令
func (dc *DetectorConnection) SendCommandToDevice(deviceID uint, command string) error {
	conn, ok := dc.GetConnection(deviceID)
	if !ok {
		return fmt.Errorf("设备 %d 未连接", deviceID)
	}

	if !conn.IsAlive() {
		dc.RemoveConnection(conn)
		return fmt.Errorf("设备 %d 连接已断开", deviceID)
	}

	// 记录最后发送的命令，用于识别响应
	dc.CommandsMutex.Lock()
	dc.LastCommands[deviceID] = command
	dc.CommandsMutex.Unlock()

	return conn.SendCommand(command)
}

// WaitResponseFromDevice 等待指定设备的响应
func (dc *DetectorConnection) WaitResponseFromDevice(deviceID uint) (string, error) {
	conn, ok := dc.GetConnection(deviceID)
	if !ok {
		return "", fmt.Errorf("设备 %d 未连接", deviceID)
	}

	return conn.WaitResponse()
}

// CheckDeviceConnected 检查设备是否已连接
func (dc *DetectorConnection) CheckDeviceConnected(deviceID uint) bool {
	conn, ok := dc.GetConnection(deviceID)
	if !ok {
		return false
	}
	return conn.IsAlive()
}

// DisconnectDevice 断开指定设备的连接
func (dc *DetectorConnection) DisconnectDevice(deviceID uint) {
	if conn, ok := dc.GetConnection(deviceID); ok {
		dc.RemoveConnection(conn)
	}
}

// GetConnectedDeviceCount 获取已连接设备数量
func (dc *DetectorConnection) GetConnectedDeviceCount() int {
	dc.ConnectionsMutex.RLock()
	defer dc.ConnectionsMutex.RUnlock()
	return len(dc.Connections)
}

// GetConnectedDeviceIDs 获取所有已连接设备的ID列表
func (dc *DetectorConnection) GetConnectedDeviceIDs() []uint {
	dc.ConnectionsMutex.RLock()
	defer dc.ConnectionsMutex.RUnlock()

	ids := make([]uint, 0, len(dc.Connections))
	for id := range dc.Connections {
		ids = append(ids, id)
	}
	return ids
}

// CheckConnections 检查所有连接的健康状态
func (dc *DetectorConnection) CheckConnections() {
	connections := dc.GetAllConnections()
	for deviceID, conn := range connections {
		if !conn.IsAlive() {
			global.Logger.Warn("检测到设备连接已断开", zap.Uint("deviceID", deviceID))
			dc.RemoveConnection(conn)
		}
	}
}

// Close 关闭连接池
func (dc *DetectorConnection) Close() {
	dc.cancel() // 取消上下文

	dc.ConnectionsMutex.Lock()
	dc.CommandsMutex.Lock()
	defer dc.ConnectionsMutex.Unlock()
	defer dc.CommandsMutex.Unlock()

	// 关闭所有连接
	for deviceID, conn := range dc.Connections {
		global.Logger.Info("关闭检测器UDP连接", zap.Uint("deviceID", deviceID))
		conn.Close()
		conn.CloseResponseChannel()
	}

	// 清空连接映射和命令记录
	dc.Connections = make(map[uint]Conn)
	dc.LastCommands = make(map[uint]string)
	global.Logger.Info("检测器UDP连接池已关闭")
}

// IsResponseOrReport 判断收到的消息是响应还是上报数据
func (dc *DetectorConnection) IsResponseOrReport(deviceID uint, message []byte) bool {
	dc.CommandsMutex.RLock()
	defer dc.CommandsMutex.RUnlock()

	lastCommand, hasCommand := dc.LastCommands[deviceID]
	if hasCommand && string(message) == lastCommand {
		return true
	}
	return false
}

// IsHeartbeatData 判断数据是否是心跳
func (dc *DetectorConnection) IsHeartbeatData(message []byte) bool {
	message = bytes.TrimSpace(message)
	return bytes.HasPrefix(message, []byte("#=")) && bytes.Contains(message, []byte("Heart_Beat"))
}

// IsSpectrumAnalysisData 判断是否是频谱分析数据
func (dc *DetectorConnection) IsSpectrumAnalysisData(messageLen int) bool {
	// 频谱分析数据通常以特定的字节序列开头，例如0xAA 0xBB 0xCC 0xDD
	return messageLen == 64*2+4
}

// 清除命令记录
func (dc *DetectorConnection) ClearCommandRecord(deviceID uint) {
	dc.CommandsMutex.Lock()
	defer dc.CommandsMutex.Unlock()
	delete(dc.LastCommands, deviceID)
}
