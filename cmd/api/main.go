package main

import (
	"context"
	"net/http"
	"os/signal"
	"strconv"
	"syscall"
	"time"
	"uav_defender/internal/app"
	"uav_defender/internal/pkg/config"
	"uav_defender/internal/pkg/utils"

	"github.com/gofiber/fiber/v2/log"
)

func gracefulShutdown(httpServer *http.Server, done chan bool) {
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
	// if err := fiberServer.ShutdownWithContext(ctx); err != nil {
	// 	log.Infof("服务器强制关闭，错误: %v", err)
	// }

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Infof("服务器强制关闭，错误: %v", err)
	}

	log.Info("服务器正在退出")

	// 通知主goroutine关闭已完成
	done <- true
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.NewConfig()

	log.SetLevel(cfg.Log.Level)

	server := app.NewFiberServer()

	httpServer := &http.Server{
		Addr:           ":" + strconv.Itoa(cfg.Server.Port),
		Handler:        server,
		ReadTimeout:    300 * time.Second,
		WriteTimeout:   300 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	wireRouter(ctx, cfg, server, utils.NewValidationMiddleware()).RegisterRoutes()

	// 创建一个完成通道，在关机完成后发出信号
	done := make(chan bool, 1)

	go func() {
		// err := server.Run(fmt.Sprintf(":%d", cfg.Server.Port))
		// if err != nil {
		// 	panic(fmt.Sprintf("HTTP服务器错误: %s", err))
		// }

		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP服务器错误: %s", err)
		}

	}()

	// 在单独的goroutine中运行优雅关闭
	go gracefulShutdown(httpServer, done)

	// 等待优雅关闭完成
	<-done
	log.Info("优雅关闭完成。")
}
