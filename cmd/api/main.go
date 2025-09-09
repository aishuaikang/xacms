package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
	"xacms/internal/server"
	"xacms/internal/utils"

	"github.com/gofiber/fiber/v2/log"
	_ "github.com/joho/godotenv/autoload"
)

func gracefulShutdown(fiberServer *server.FiberServer, done chan bool) {
	// 创建监听来自操作系统的中断信号的上下文。
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// 监听中断信号。
	<-ctx.Done()

	log.Info("优雅地关闭，再次按Ctrl+C强制关闭")
	stop() // 允许Ctrl+C强制关闭

	// 上下文用于通知服务器它有5秒钟的时间来完成
	// 它当前正在处理的请求
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := fiberServer.ShutdownWithContext(ctx); err != nil {
		log.Infof("服务器强制关闭，错误: %v", err)
	}

	log.Info("服务器正在退出")

	// 通知主goroutine关闭已完成
	done <- true
}

func main() {
	// _, cancel := context.WithCancel(context.Background())
	// defer cancel()

	server := server.NewFiberServer()

	wireRouter(server, utils.NewValidationMiddleware()).RegisterRoutes()

	// 创建一个完成通道，在关机完成后发出信号
	done := make(chan bool, 1)

	go func() {
		port, _ := strconv.Atoi(os.Getenv("PORT"))
		err := server.Listen(fmt.Sprintf(":%d", port))
		if err != nil {
			panic(fmt.Sprintf("HTTP服务器错误: %s", err))
		}
	}()

	// 在单独的goroutine中运行优雅关闭
	go gracefulShutdown(server, done)

	// 等待优雅关闭完成
	<-done
	log.Info("优雅关闭完成。")
}
