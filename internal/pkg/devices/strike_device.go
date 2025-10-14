package devices

import (
	"fmt"
	"net"
	"time"
	"uav_defender/internal/pkg/config"
	"uav_defender/internal/pkg/global"
	"uav_defender/internal/pkg/protocol"
	"uav_defender/internal/pkg/utils"

	"go.uber.org/zap"
)

type StrikeDevice struct {
	tcpServer *utils.TCPServer
}

func NewStrikeDevice() *StrikeDevice {
	return &StrikeDevice{}
}

func (s *StrikeDevice) Run() {
	// 获取打击设备TCP服务器地址
	address := config.AppConfig.Configuration.StrikerTcpServer
	if address == "" {
		global.Logger.Error("打击设备TCP服务器地址未配置")
		return
	}

	// 创建TCP服务器配置，使用自定义的打击设备协议解析器
	serverConfig := utils.TCPServerConfig{
		Address:                address,
		HeartbeatTimeout:       30 * time.Second,
		HeartbeatCheckInterval: 10 * time.Second,
		ReadBufferSize:         4096,
		EnableHeartbeat:        true,
		Parser:                 protocol.NewStrikeProtocolParser(), // 使用打击设备协议解析器
	}

	// 创建TCP服务器，传入自己作为消息处理器
	s.tcpServer = utils.NewTCPServer(serverConfig, s)

	// 启动服务器
	err := s.tcpServer.Start()
	if err != nil {
		global.Logger.Error(fmt.Sprintf("启动打击设备TCP服务失败: %v", err))
		return
	}
}

// OnMessage 实现MessageHandler接口 - 处理接收到的消息
func (s *StrikeDevice) OnMessage(data []byte, conn net.Conn, remoteAddr string) []byte {
	global.Logger.Info(fmt.Sprintf("收到来自 %s 的原始数据，长度: %d bytes", remoteAddr, len(data)))

	// 解析协议包
	packet, err := protocol.ParseStrikePacket(data)
	if err != nil {
		global.Logger.Error(fmt.Sprintf("解析打击设备协议包失败: %v，来自: %s", err, remoteAddr))
		return nil
	}

	global.Logger.Info(fmt.Sprintf("收到来自 %s 的消息: %s", remoteAddr, packet.String()))

	// 处理业务逻辑
	response := s.handleStrikePacket(packet, remoteAddr)

	// 返回响应（如果有）
	return response
}

// handleStrikePacket 处理打击设备协议包
func (s *StrikeDevice) handleStrikePacket(packet *protocol.StrikePacket, remoteAddr string) []byte {
	global.Logger.Info(fmt.Sprintf("处理命令: %s (0x%02X)，设备ID: %d，来自: %s",
		packet.GetCmdName(), packet.Cmd, packet.ID, remoteAddr))

	// 根据命令类型进行处理
	switch packet.Cmd {
	case protocol.CmdSetPower:
		return s.handleSetPower(packet, remoteAddr)
	case protocol.CmdQueryModules:
		return s.handleQueryModules(packet, remoteAddr)
	case protocol.CmdDownloadAddr:
		return s.handleDownloadAddr(packet, remoteAddr)
	case protocol.CmdQueryAddr:
		return s.handleQueryAddr(packet, remoteAddr)
	default:
		global.Logger.Warn(fmt.Sprintf("未知的命令类型: 0x%02X，来自: %s", packet.Cmd, remoteAddr))
		return nil
	}
}

// handleSetPower 处理设置N路开关功率命令
func (s *StrikeDevice) handleSetPower(packet *protocol.StrikePacket, remoteAddr string) []byte {
	global.Logger.Debug("处理设置功率命令: ", zap.Any("packet", packet), zap.String("from", remoteAddr))

	// TODO: 实现具体的业务逻辑

	// 示例：返回确认响应
	// return protocol.BuildStrikePacket(packet.ID, protocol.CmdSetPower, []byte{0x01}) // 0x01表示成功
	return nil
}

// handleQueryModules 处理查询N路模块信息命令
func (s *StrikeDevice) handleQueryModules(packet *protocol.StrikePacket, remoteAddr string) []byte {
	global.Logger.Debug("处理查询模块信息命令: ", zap.Any("packet", packet), zap.String("from", remoteAddr))

	// TODO: 实现具体的业务逻辑
	// 示例：返回模块信息
	// moduleInfo := []byte{0x01, 0x02, 0x03, 0x04} // 模块信息数据
	// return protocol.BuildStrikePacket(packet.ID, protocol.CmdQueryModules, moduleInfo)
	return nil
}

// handleDownloadAddr 处理下载地址列表命令
func (s *StrikeDevice) handleDownloadAddr(packet *protocol.StrikePacket, remoteAddr string) []byte {
	global.Logger.Debug("处理下载地址列表命令: ", zap.Any("packet", packet), zap.String("from", remoteAddr))

	// TODO: 实现具体的业务逻辑
	return nil
}

// handleQueryAddr 处理查询地址列表命令
func (s *StrikeDevice) handleQueryAddr(packet *protocol.StrikePacket, remoteAddr string) []byte {
	global.Logger.Debug("处理查询地址列表命令: ", zap.Any("packet", packet), zap.String("from", remoteAddr))

	// TODO: 实现具体的业务逻辑
	// 示例：返回地址列表
	// addrList := []byte{...} // 地址列表数据
	// return protocol.BuildStrikePacket(packet.ID, protocol.CmdQueryAddr, addrList)
	return nil
}

// OnConnect 实现MessageHandler接口 - 客户端连接时调用
func (s *StrikeDevice) OnConnect(conn net.Conn, remoteAddr string) {
	global.Logger.Info(fmt.Sprintf("打击设备客户端已连接: %s", remoteAddr))

	// TODO: 可以在这里进行连接初始化操作
	// 例如：发送欢迎消息、请求设备信息等
}

// OnDisconnect 实现MessageHandler接口 - 客户端断开连接时调用
func (s *StrikeDevice) OnDisconnect(conn net.Conn, remoteAddr string) {
	global.Logger.Info(fmt.Sprintf("打击设备客户端已断开: %s", remoteAddr))

	// TODO: 可以在这里进行清理操作
	// 例如：清除设备状态、通知其他模块等
}

// // SendToDevice 发送消息到指定设备
// func (s *StrikeDevice) SendToDevice(remoteAddr string, data []byte) error {
// 	if s.tcpServer == nil {
// 		return fmt.Errorf("TCP服务器未启动")
// 	}
// 	return s.tcpServer.SendToClient(remoteAddr, data)
// }

// // SendPacketToDevice 发送协议包到指定设备
// func (s *StrikeDevice) SendPacketToDevice(remoteAddr string, id uint16, cmd uint8, data []byte) error {
// 	packet := protocol.BuildStrikePacket(id, cmd, data)
// 	return s.SendToDevice(remoteAddr, packet)
// }

// // BroadcastToAllDevices 广播消息到所有设备
// func (s *StrikeDevice) BroadcastToAllDevices(data []byte) {
// 	if s.tcpServer != nil {
// 		s.tcpServer.Broadcast(data)
// 	}
// }

// // BroadcastPacketToAllDevices 广播协议包到所有设备
// func (s *StrikeDevice) BroadcastPacketToAllDevices(id uint16, cmd uint8, data []byte) {
// 	packet := protocol.BuildStrikePacket(id, cmd, data)
// 	s.BroadcastToAllDevices(packet)
// }

// GetConnectedDeviceCount 获取当前连接的设备数量
func (s *StrikeDevice) GetConnectedDeviceCount() int {
	if s.tcpServer == nil {
		return 0
	}
	return s.tcpServer.GetClientCount()
}

// Stop 停止TCP服务
func (s *StrikeDevice) Stop() {
	if s.tcpServer != nil {
		s.tcpServer.Stop()
		global.Logger.Info("打击设备TCP服务已停止")
	}
}
