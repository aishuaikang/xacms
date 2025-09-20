package main

import (
	"context"
	"net/http"
	"os/signal"
	"strconv"
	"syscall"
	"time"
	"uav_defender/internal/pkg/config"
	"uav_defender/internal/pkg/database"
	"uav_defender/internal/pkg/global"
	"uav_defender/internal/pkg/utils"

	"go.uber.org/zap"
)

func gracefulShutdown(httpServer *http.Server, done chan bool) {
	// 创建监听来自操作系统的中断信号的上下文。
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// 监听中断信号。
	<-ctx.Done()

	global.Logger.Info("收到中断信号，正在关闭服务器...")
	stop() // 允许Ctrl+C强制关闭

	// 上下文用于通知服务器它有5秒钟的时间来完成
	// 它当前正在处理的请求
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		// log.Fatalf("服务器关闭时出现错误: %v", err)
		global.Logger.Fatal("服务器关闭时出现错误", zap.Error(err))
	}

	global.Logger.Info("服务器关闭")

	// 通知主goroutine关闭已完成
	done <- true
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.InitConfig()
	config.InitMediaMtxConfig()

	logger := global.NewZapLogger(cfg.Log.Level, cfg.Log.Enabled)
	defer logger.Sync() // 确保日志被刷新

	db, err := database.NewDB(cfg)
	if err != nil {
		logger.Fatal("数据库连接失败", zap.Error(err))
		return
	}
	defer func() {
		sqlDB, err := db.DB()
		if err != nil {
			logger.Error("获取数据库实例时出错", zap.Error(err))
			return
		}
		if err := sqlDB.Close(); err != nil {
			logger.Error("关闭数据库连接时出错", zap.Error(err))
		}
	}()

	server := wireServer(ctx, db, utils.NewValidationMiddleware())

	httpServer := &http.Server{
		Addr:           ":" + strconv.Itoa(cfg.Server.Port),
		Handler:        server,
		ReadTimeout:    300 * time.Second,
		WriteTimeout:   300 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// 创建一个完成通道，在关机完成后发出信号
	done := make(chan bool, 1)

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("HTTP服务器错误", zap.Error(err))
		}
	}()

	// 在单独的goroutine中运行优雅关闭
	go gracefulShutdown(httpServer, done)

	// 等待优雅关闭完成
	<-done
	logger.Info("服务器已关闭")
}
