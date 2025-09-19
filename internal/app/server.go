package app

import (
	"time"
	"uav_defender/internal/app/devices"
	"uav_defender/internal/pkg/global"
	"uav_defender/internal/routes"
	"uav_defender/internal/tasks"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
)

// type FiberServer struct {
// 	*fiber.App
// }

type GinServer struct {
	*gin.Engine
	router  *routes.Router
	tasks   *tasks.Tasks
	devices *devices.Devices
}

func NewGinServer(router *routes.Router, tasks *tasks.Tasks, devices *devices.Devices) *GinServer {
	engine := gin.New()

	engine.Use(ginzap.Ginzap(global.Logger, time.DateTime, false))
	engine.Use(ginzap.RecoveryWithZap(global.Logger, true))

	engine.MaxMultipartMemory = 2 << 30

	server := &GinServer{
		Engine:  engine,
		router:  router,
		tasks:   tasks,
		devices: devices,
	}

	apiV1 := engine.Group("/api/v1")

	// 注冊路由
	server.router.RegisterRoutes(apiV1)

	// 启动设备处理
	server.devices.Start()

	// 执行任务
	server.tasks.Execute()

	return server
}
