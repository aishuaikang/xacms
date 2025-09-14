package utils

import (
	"context"
	"net"

	"github.com/gofiber/fiber/v2/log"
)

// 构建TCP服务器
func BuildTcpServer(ctx context.Context, module string, addr string, handler func(module string, conn net.Conn)) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Errorf("%s 服务器监听地址 %s 失败: %v", module, addr, err)
		return
	}
	defer listener.Close()

	log.Infof("%s 服务器启动成功，监听地址: %s", module, addr)

	// 用一个 goroutine 监听 ctx.Done()，在取消时关闭 listener
	go func() {
		<-ctx.Done()
		log.Infof("%s 服务器正在停止...", module)
		listener.Close() // 会导致 Accept 返回错误，从而退出主循环
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				log.Infof("%s 服务器已正常退出", module)
				return
			default:
				log.Errorf("%s 服务器接受连接失败: %v", module, err)
				continue
			}
		}

		// 每个连接交给单独 goroutine 处理
		go handler(module, conn)
	}
}
