package utils

import (
	"context"
	"net"
	"uav_defender/internal/pkg/global"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// EqualUUID 判断两个UUID是否相等
func EqualUUID(a, b *uuid.UUID) bool {
	if a == b {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

// 构建TCP服务器
func BuildTcpServer(ctx context.Context, module string, addr string, handler func(module string, conn net.Conn)) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		global.Logger.Error("启动TCP服务器失败", zap.String("module", module), zap.String("address", addr), zap.Error(err))
		return
	}
	defer listener.Close()

	global.Logger.Info("启动TCP服务器成功", zap.String("module", module), zap.String("address", addr))

	// 用一个 goroutine 监听 ctx.Done()，在取消时关闭 listener
	go func() {
		<-ctx.Done()
		global.Logger.Info("停止TCP服务器", zap.String("module", module))
		listener.Close() // 会导致 Accept 返回错误，从而退出主循环
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				global.Logger.Info("停止TCP服务器", zap.String("module", module))
				return
			default:
				global.Logger.Error("启动TCP服务器失败", zap.String("module", module), zap.String("address", addr), zap.Error(err))
				continue
			}
		}

		// 每个连接交给单独 goroutine 处理
		go handler(module, conn)
	}
}
