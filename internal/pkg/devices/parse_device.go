package devices

import (
	"fmt"
	"net"
	"time"
	"uav_defender/internal/pkg/config"
	"uav_defender/internal/pkg/global"
	"uav_defender/internal/pkg/utils"

	"go.uber.org/zap"
)

type ParseDevice struct {
	tcpServer *utils.TCPServer
}

func NewParseDevice() *ParseDevice {
	return &ParseDevice{}
}

func (p *ParseDevice) Run() {
	// 获取解析设备TCP服务器地址
	address := config.AppConfig.Configuration.ParserTcpServer
	if address == "" {
		global.Logger.Error("解析设备TCP服务器地址未配置")
		return
	}

	// 创建TCP服务器配置
	serverConfig := utils.TCPServerConfig{
		Address:                address,
		HeartbeatTimeout:       30 * time.Second,
		HeartbeatCheckInterval: 10 * time.Second,
		ReadBufferSize:         4096,
		MessageDelimiter:       '\n',
		EnableHeartbeat:        true,
	}

	// 创建TCP服务器，传入自己作为消息处理器
	p.tcpServer = utils.NewTCPServer(serverConfig, p)

	// 启动服务器
	err := p.tcpServer.Start()
	if err != nil {
		global.Logger.Error(fmt.Sprintf("启动解析设备TCP服务失败: %v", err))
		return
	}
}

// OnMessage 实现MessageHandler接口 - 处理接收到的消息
func (p *ParseDevice) OnMessage(data []byte, conn net.Conn, remoteAddr string) []byte {
	global.Logger.Info(fmt.Sprintf("收到来自 %s 的消息: %s", remoteAddr, string(data)))

	// TODO: 在这里实现你的业务逻辑
	// 解析数据，进行业务处理
	p.parseData(data, remoteAddr)

	// 如果需要回复客户端，返回响应数据
	// 例如: return []byte("OK")
	// 如果不需要回复，返回nil
	return nil
}

// OnConnect 实现MessageHandler接口 - 客户端连接时调用
func (p *ParseDevice) OnConnect(conn net.Conn, remoteAddr string) {
	global.Logger.Info(fmt.Sprintf("解析设备客户端已连接: %s", remoteAddr))

	// TODO: 可以在这里进行连接初始化操作
	// 例如：发送欢迎消息、请求设备信息等
}

// OnDisconnect 实现MessageHandler接口 - 客户端断开连接时调用
func (p *ParseDevice) OnDisconnect(conn net.Conn, remoteAddr string) {
	global.Logger.Info(fmt.Sprintf("解析设备客户端已断开: %s", remoteAddr))

	// TODO: 可以在这里进行清理操作
	// 例如：清除设备状态、通知其他模块等
}

// parseData 解析接收到的数据 - 业务逻辑处理
func (p *ParseDevice) parseData(data []byte, remoteAddr string) {
	// TODO: 实现具体的数据解析逻辑
	global.Logger.Debug("解析数据: ", zap.ByteString("data", data), zap.String("from", remoteAddr))

	// 示例：根据数据内容进行不同的处理
	// if bytes.HasPrefix(data, []byte("CMD:")) {
	//     // 处理命令
	// } else if bytes.HasPrefix(data, []byte("DATA:")) {
	//     // 处理数据
	// }
}

// SendToDevice 发送消息到指定设备
func (p *ParseDevice) SendToDevice(remoteAddr string, data []byte) error {
	if p.tcpServer == nil {
		return fmt.Errorf("TCP服务器未启动")
	}
	return p.tcpServer.SendToClient(remoteAddr, data)
}

// BroadcastToAllDevices 广播消息到所有设备
func (p *ParseDevice) BroadcastToAllDevices(data []byte) {
	if p.tcpServer != nil {
		p.tcpServer.Broadcast(data)
	}
}

// GetConnectedDeviceCount 获取当前连接的设备数量
func (p *ParseDevice) GetConnectedDeviceCount() int {
	if p.tcpServer == nil {
		return 0
	}
	return p.tcpServer.GetClientCount()
}

// Stop 停止TCP服务
func (p *ParseDevice) Stop() {
	if p.tcpServer != nil {
		p.tcpServer.Stop()
		global.Logger.Info("解析设备TCP服务已停止")
	}
}
