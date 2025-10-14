package app

import (
	"time"
	"uav_defender/internal/middlewares"
	"uav_defender/internal/pkg/devices"
	"uav_defender/internal/pkg/global"
	"uav_defender/internal/routes"
	"uav_defender/internal/tasks"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap/zapcore"
)

// type FiberServer struct {
// 	*fiber.App
// }

type GinServer struct {
	*gin.Engine
	router      *routes.Router
	taskManager *tasks.TaskManager
}

func NewGinServer(router *routes.Router, taskManager *tasks.TaskManager, deviceManager *devices.DeviceManager) *GinServer {
	engine := gin.New()

	// 使用自定义日志中间件，包含更详细的信息（如处理函数位置）
	engine.Use(middlewares.GinLoggerWithConfig(global.Logger, &middlewares.Config{
		TimeFormat:   time.DateTime,
		UTC:          false,
		DefaultLevel: zapcore.InfoLevel,
		EnableBody:   true,
	}))
	engine.Use(middlewares.GinRecovery(global.Logger, true))

	engine.MaxMultipartMemory = 2 << 30

	server := &GinServer{
		Engine:      engine,
		router:      router,
		taskManager: taskManager,
		// mqttClient: mqttClient,
	}

	apiV1 := engine.Group("")

	// 注册路由
	server.router.RegisterRoutes(apiV1)

	// 执行任务
	server.taskManager.Execute()

	// 运行设备管理器
	deviceManager.Run()

	// // 启动MQTT服务
	// if err := server.mqttClient.Connect(); err != nil {
	// 	global.Logger.Error("MQTT服务启动失败", zap.Error(err))
	// }

	return server
}
