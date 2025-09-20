package conn

import (
	"fmt"
	"net"
	"sync"
	"time"
	"uav_defender/internal/pkg/global"

	"go.uber.org/zap"
)

var (
	// 响应超时
	ErrResponseTimeout = fmt.Errorf("响应超时")
)

type Conn interface {
	SendCommand(command string) error
	WaitResponse() (string, error)
	IsAlive() bool
	Close()
	CloseResponseChannel()
	GetConn() net.Conn
	GetResponseChannel() chan<- string
	SendMessage(message string) error
}

// Conn 对于 conn 的抽象封装
type conn struct {
	conn     net.Conn
	isAlive  bool
	response chan string
}

func NewConn(c net.Conn) Conn {
	return &conn{
		conn:     c,
		isAlive:  true,
		response: make(chan string, 1),
	}

}

// SendCommand 发送命令到 FPV 设备
func (f *conn) SendCommand(command string) error {
	global.Logger.Info("发送命令", zap.String("address", f.conn.RemoteAddr().String()), zap.String("command", command))
	_, err := f.conn.Write([]byte(command))
	return err
}

// WaitResponse 等待来自 FPV 的响应，超时返回错误
func (f *conn) WaitResponse() (string, error) {
	global.Logger.Info("等待响应", zap.String("address", f.conn.RemoteAddr().String()))
	select {
	case resp := <-f.response:
		return resp, nil
	case <-time.After(10 * time.Second):
		return "", fmt.Errorf("响应超时")
	}
}

// IsAlive 返回连接是否存活
func (f *conn) IsAlive() bool {
	return f.isAlive
}

// Close 关闭连接
func (f *conn) Close() {
	f.isAlive = false
	f.conn.Close()
}

// CloseResponseChannel 关闭响应通道
func (f *conn) CloseResponseChannel() {
	close(f.response)
}

// GetConn 返回底层 net.Conn
func (f *conn) GetConn() net.Conn {
	return f.conn
}

// GetResponseChannel 返回响应通道
func (f *conn) GetResponseChannel() chan<- string {
	return f.response
}

// SendMessage 发送消息到 FPV 设备
func (f *conn) SendMessage(message string) error {
	_, err := f.conn.Write([]byte(message))
	return err
}

type Connection struct {
	Connections      []Conn
	ConnectionsMutex sync.RWMutex
}

// GetAllConnections 获取所有连接
func (f *Connection) GetAllConnections() []Conn {
	f.ConnectionsMutex.RLock()
	defer f.ConnectionsMutex.RUnlock()
	return f.Connections
}

// SetConnections 设置所有连接
func (f *Connection) SetConnections(conns []Conn) {
	f.ConnectionsMutex.Lock()
	defer f.ConnectionsMutex.Unlock()
	f.Connections = conns
}

// AddConnection 添加新的连接
func (f *Connection) AddConnection(conn Conn) {
	f.ConnectionsMutex.Lock()
	defer f.ConnectionsMutex.Unlock()
	f.Connections = append(f.Connections, conn)
}

// RemoveConnection 移除连接
func (f *Connection) RemoveConnection(removeConn Conn) {
	filteredConnections := make([]Conn, 0)
	allConnections := f.GetAllConnections()
	for _, conn := range allConnections {
		if conn.GetConn().RemoteAddr().String() != removeConn.GetConn().RemoteAddr().String() {
			filteredConnections = append(filteredConnections, conn)
		} else {
			conn.Close()
			conn.CloseResponseChannel()
		}
	}
	f.SetConnections(filteredConnections)
}

// GetConnection 获取连接
func (f *Connection) GetConnection(addr string) (Conn, bool) {
	allConnections := f.GetAllConnections()
	for _, conn := range allConnections {
		if conn.GetConn().RemoteAddr().String() == addr {
			return conn, true
		}
	}
	return nil, false
}
