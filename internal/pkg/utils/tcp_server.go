package utils

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
	"uav_defender/internal/pkg/global"
)

// MessageHandler 消息处理器接口
type MessageHandler interface {
	// OnMessage 当收到完整消息时调用
	// data: 消息数据
	// conn: 客户端连接，可用于回复消息
	// remoteAddr: 客户端地址
	// 返回值: 需要回复给客户端的数据，如果为nil则不回复
	OnMessage(data []byte, conn net.Conn, remoteAddr string) []byte

	// OnConnect 当客户端连接时调用（可选）
	OnConnect(conn net.Conn, remoteAddr string)

	// OnDisconnect 当客户端断开连接时调用（可选）
	OnDisconnect(conn net.Conn, remoteAddr string)
}

// ProtocolParser 协议解析器接口
type ProtocolParser interface {
	// Parse 从缓冲区解析出完整的消息包
	// 返回值: 消息数据, 已消费的字节数, 错误
	// 如果数据不完整返回 (nil, 0, nil)
	// 如果解析失败返回 (nil, consumedBytes, error)
	// 如果解析成功返回 (message, consumedBytes, nil)
	Parse(buffer []byte) ([]byte, int, error)

	// Build 构造符合协议的消息包
	Build(data []byte) []byte
}

// DelimiterParser 基于分隔符的协议解析器（默认实现）
type DelimiterParser struct {
	Delimiter byte
}

func (p *DelimiterParser) Parse(buffer []byte) ([]byte, int, error) {
	// 查找分隔符
	for i, b := range buffer {
		if b == p.Delimiter {
			// 找到完整消息
			message := buffer[:i]
			consumed := i + 1
			return message, consumed, nil
		}
	}
	// 数据不完整
	return nil, 0, nil
}

func (p *DelimiterParser) Build(data []byte) []byte {
	return append(data, p.Delimiter)
}

// TCPServerConfig TCP服务器配置
type TCPServerConfig struct {
	Address                string         // 监听地址，如 "0.0.0.0:8080"
	HeartbeatTimeout       time.Duration  // 心跳超时时间，默认30秒
	HeartbeatCheckInterval time.Duration  // 心跳检查间隔，默认10秒
	ReadBufferSize         int            // 读取缓冲区大小，默认4096
	MessageDelimiter       byte           // 消息分隔符，默认'\n'（仅在使用默认解析器时有效）
	EnableHeartbeat        bool           // 是否启用心跳检测，默认true
	Parser                 ProtocolParser // 自定义协议解析器，如果为nil则使用基于分隔符的解析器
}

// TCPServer TCP服务器
type TCPServer struct {
	config   TCPServerConfig
	handler  MessageHandler
	parser   ProtocolParser
	listener net.Listener
	connMap  sync.Map // 存储客户端连接信息，key为远程地址，value为*clientConnection
	stopCh   chan struct{}
}

// clientConnection 客户端连接信息
type clientConnection struct {
	conn          net.Conn
	lastActive    time.Time
	mu            sync.RWMutex
	stopHeartbeat chan struct{}
}

// NewTCPServer 创建TCP服务器
func NewTCPServer(config TCPServerConfig, handler MessageHandler) *TCPServer {
	// 设置默认值
	if config.HeartbeatTimeout == 0 {
		config.HeartbeatTimeout = 30 * time.Second
	}
	if config.HeartbeatCheckInterval == 0 {
		config.HeartbeatCheckInterval = 10 * time.Second
	}
	if config.ReadBufferSize == 0 {
		config.ReadBufferSize = 4096
	}
	if config.MessageDelimiter == 0 {
		config.MessageDelimiter = '\n'
	}
	if !config.EnableHeartbeat {
		config.EnableHeartbeat = true
	}

	// 如果没有指定自定义解析器，使用默认的分隔符解析器
	parser := config.Parser
	if parser == nil {
		parser = &DelimiterParser{Delimiter: config.MessageDelimiter}
	}

	return &TCPServer{
		config:  config,
		handler: handler,
		parser:  parser,
		stopCh:  make(chan struct{}),
	}
}

// Start 启动TCP服务器
func (s *TCPServer) Start() error {
	if s.config.Address == "" {
		return fmt.Errorf("TCP服务器地址未配置")
	}

	global.Logger.Info(fmt.Sprintf("正在启动TCP服务器，监听地址: %s", s.config.Address))

	listener, err := net.Listen("tcp", s.config.Address)
	if err != nil {
		return fmt.Errorf("启动TCP服务器失败: %v", err)
	}

	s.listener = listener
	global.Logger.Info(fmt.Sprintf("TCP服务器启动成功，监听地址: %s", s.config.Address))

	// 在新的goroutine中接受连接
	go s.acceptConnections()

	return nil
}

// Stop 停止TCP服务器
func (s *TCPServer) Stop() {
	close(s.stopCh)

	if s.listener != nil {
		s.listener.Close()
		global.Logger.Info("TCP服务器已停止")
	}

	// 关闭所有客户端连接
	s.connMap.Range(func(key, value interface{}) bool {
		if client, ok := value.(*clientConnection); ok {
			client.conn.Close()
		}
		return true
	})
}

// acceptConnections 接受客户端连接
func (s *TCPServer) acceptConnections() {
	for {
		select {
		case <-s.stopCh:
			return
		default:
		}

		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.stopCh:
				return
			default:
				global.Logger.Error(fmt.Sprintf("接受客户端连接失败: %v", err))
				continue
			}
		}

		remoteAddr := conn.RemoteAddr().String()
		global.Logger.Info(fmt.Sprintf("新的客户端连接: %s", remoteAddr))

		// 调用OnConnect回调
		if s.handler != nil {
			s.handler.OnConnect(conn, remoteAddr)
		}

		// 为每个连接启动一个goroutine处理
		go s.handleConnection(conn)
	}
}

// handleConnection 处理客户端连接
func (s *TCPServer) handleConnection(conn net.Conn) {
	defer func() {
		remoteAddr := conn.RemoteAddr().String()
		conn.Close()

		// 从连接映射中移除
		s.connMap.Delete(remoteAddr)

		// 调用OnDisconnect回调
		if s.handler != nil {
			s.handler.OnDisconnect(conn, remoteAddr)
		}

		global.Logger.Info(fmt.Sprintf("客户端 %s 连接处理结束", remoteAddr))
	}()

	remoteAddr := conn.RemoteAddr().String()
	global.Logger.Info(fmt.Sprintf("开始处理客户端连接: %s", remoteAddr))

	// 创建客户端连接对象
	client := &clientConnection{
		conn:          conn,
		lastActive:    time.Now(),
		stopHeartbeat: make(chan struct{}),
	}

	// 保存到连接映射
	s.connMap.Store(remoteAddr, client)

	// 启动心跳检测
	if s.config.EnableHeartbeat {
		go s.checkHeartbeat(client, remoteAddr)
	}

	// 确保关闭心跳检测
	defer close(client.stopHeartbeat)

	// 使用缓冲区来处理粘包和分包
	buffer := bytes.NewBuffer(nil)
	tempBuf := make([]byte, s.config.ReadBufferSize)

	for {
		// 设置读取超时
		if s.config.EnableHeartbeat {
			conn.SetReadDeadline(time.Now().Add(s.config.HeartbeatTimeout))
		}

		n, err := conn.Read(tempBuf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				global.Logger.Warn(fmt.Sprintf("客户端 %s 读取超时", remoteAddr))
				break
			}
			if err != io.EOF {
				global.Logger.Info(fmt.Sprintf("客户端 %s 断开连接: %v", remoteAddr, err))
			} else {
				global.Logger.Info(fmt.Sprintf("客户端 %s 正常断开连接", remoteAddr))
			}
			break
		}

		if n > 0 {
			// 更新最后活跃时间
			client.mu.Lock()
			client.lastActive = time.Now()
			client.mu.Unlock()

			// 将读取的数据追加到缓冲区
			buffer.Write(tempBuf[:n])
			global.Logger.Debug(fmt.Sprintf("收到来自 %s 的数据，长度: %d bytes，缓冲区总长度: %d bytes",
				remoteAddr, n, buffer.Len()))

			// 处理缓冲区中的所有完整消息
			s.processBuffer(buffer, client, remoteAddr)
		}
	}
}

// processBuffer 处理缓冲区中的数据，解决粘包和分包问题
func (s *TCPServer) processBuffer(buffer *bytes.Buffer, client *clientConnection, remoteAddr string) {
	for {
		// 使用协议解析器解析消息
		message, consumed, err := s.parser.Parse(buffer.Bytes())

		if err != nil {
			// 解析错误，记录并清空缓冲区
			global.Logger.Error(fmt.Sprintf("解析消息失败: %v，来自: %s", err, remoteAddr))
			// 丢弃已消费的字节
			if consumed > 0 && consumed <= buffer.Len() {
				buffer.Next(consumed)
			} else {
				// 如果消费字节数异常，清空整个缓冲区
				buffer.Reset()
			}
			continue
		}

		if message == nil {
			// 数据不完整，等待更多数据（分包情况）
			// 只在缓冲区确实有数据时才打印日志，避免误导
			if buffer.Len() > 0 {
				global.Logger.Debug(fmt.Sprintf("消息未完整接收，等待更多数据，缓冲区当前有 %d bytes", buffer.Len()))
			}
			break
		}

		// 从缓冲区移除已消费的字节
		if consumed > 0 {
			buffer.Next(consumed)
		}

		// 更新最后活跃时间（收到有效数据）
		client.mu.Lock()
		client.lastActive = time.Now()
		client.mu.Unlock()

		global.Logger.Debug(fmt.Sprintf("解析到完整消息，长度: %d bytes，来自: %s", len(message), remoteAddr))

		// 调用业务处理器
		if s.handler != nil {
			response := s.handler.OnMessage(message, client.conn, remoteAddr)

			// 如果有响应数据，发送回客户端
			if len(response) > 0 {
				s.sendMessage(client.conn, response, remoteAddr)
			}
		}

		// 继续处理缓冲区中的下一条消息（处理粘包情况）
	}
} // sendMessage 发送消息到客户端
func (s *TCPServer) sendMessage(conn net.Conn, data []byte, remoteAddr string) error {
	// 使用协议解析器构造消息
	message := s.parser.Build(data)

	// 确保完整发送
	totalSent := 0
	for totalSent < len(message) {
		n, err := conn.Write(message[totalSent:])
		if err != nil {
			global.Logger.Error(fmt.Sprintf("发送消息到 %s 失败: %v", remoteAddr, err))
			return err
		}
		totalSent += n
	}

	global.Logger.Debug(fmt.Sprintf("成功发送消息到 %s，长度: %d bytes", remoteAddr, len(data)))
	return nil
}

// checkHeartbeat 检测客户端心跳
func (s *TCPServer) checkHeartbeat(client *clientConnection, remoteAddr string) {
	ticker := time.NewTicker(s.config.HeartbeatCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 检查最后活跃时间
			client.mu.RLock()
			lastActive := client.lastActive
			client.mu.RUnlock()

			if time.Since(lastActive) > s.config.HeartbeatTimeout {
				global.Logger.Warn(fmt.Sprintf("客户端 %s 心跳超时，最后活跃时间: %s，关闭连接",
					remoteAddr, lastActive.Format(time.DateTime)))
				client.conn.Close()
				return
			}

			global.Logger.Debug(fmt.Sprintf("客户端 %s 心跳正常，最后活跃: %s",
				remoteAddr, lastActive.Format(time.DateTime)))

		case <-client.stopHeartbeat:
			global.Logger.Debug(fmt.Sprintf("停止客户端 %s 的心跳检测", remoteAddr))
			return
		}
	}
}

// SendToClient 向指定客户端发送消息
func (s *TCPServer) SendToClient(remoteAddr string, data []byte) error {
	value, ok := s.connMap.Load(remoteAddr)
	if !ok {
		return fmt.Errorf("客户端 %s 不存在或已断开", remoteAddr)
	}

	client, ok := value.(*clientConnection)
	if !ok {
		return fmt.Errorf("无效的客户端连接对象")
	}

	return s.sendMessage(client.conn, data, remoteAddr)
}

// Broadcast 广播消息给所有客户端
func (s *TCPServer) Broadcast(data []byte) {
	s.connMap.Range(func(key, value interface{}) bool {
		if remoteAddr, ok := key.(string); ok {
			if client, ok := value.(*clientConnection); ok {
				s.sendMessage(client.conn, data, remoteAddr)
			}
		}
		return true
	})
}

// GetClientCount 获取当前连接的客户端数量
func (s *TCPServer) GetClientCount() int {
	count := 0
	s.connMap.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}
