package main

import (
	"context"
	"fmt"
	"log"
	"new-spbatc-drone-platform/internal/server"
	"new-spbatc-drone-platform/internal/utils"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	_ "github.com/joho/godotenv/autoload"
)

func gracefulShutdown(fiberServer *server.FiberServer, done chan bool) {
	// 创建监听来自操作系统的中断信号的上下文。
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// 监听中断信号。
	<-ctx.Done()

	log.Println("优雅地关闭，再次按Ctrl+C强制关闭")
	stop() // 允许Ctrl+C强制关闭

	// 上下文用于通知服务器它有5秒钟的时间来完成
	// 它当前正在处理的请求
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := fiberServer.ShutdownWithContext(ctx); err != nil {
		log.Printf("服务器强制关闭，错误: %v", err)
	}

	log.Println("服务器正在退出")

	// 通知主goroutine关闭已完成
	done <- true
}

func main() {
	server := server.NewFiberServer()

	router := wireRouter(server, utils.NewValidationMiddleware())
	router.RegisterRoutes()

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
	log.Println("优雅关闭完成。")
}
