package routes

import (
	"uav_defender/internal/app"
	"uav_defender/internal/app/devices"
	"uav_defender/internal/tasks"

	"github.com/gin-gonic/gin"
)

// RouteModule 定义路由模块接口
type RouteModule interface {
	RegisterRoutes(router *gin.RouterGroup)
}

// Router 路由注册器
type Router struct {
	server  *app.GinServer
	modules []RouteModule
}

// NewRouter 创建路由注册器
func NewRouter(server *app.GinServer,
	devices *devices.Devices,
	tasks *tasks.Tasks,
	userHandler *UserHandler,
	menuHandler *MenuHandler,
	roleHandler *RoleHandler,
	deviceHandler *DeviceHandler,
	droneTargetHandler *DroneTargetHandler,
	sseHandler *SSEHandler,
) *Router {

	// 启动设备相关服务
	devices.Start()

	// 执行任务
	tasks.Execute()

	return &Router{
		server: server,
		modules: []RouteModule{
			userHandler,
			menuHandler,
			roleHandler,
			deviceHandler,
			droneTargetHandler,
			sseHandler,
		},
	}
}

// RegisterRoutes 注册所有模块路由
func (r *Router) RegisterRoutes() {
	// 创建 API 版本组
	apiV1 := r.server.Engine.Group("/api/v1")

	// 注册公开路由（不需要认证）
	// publicRoutes := apiV1.Group("/public")
	// publicRoutes.Get("/health", r.HealthCheck)

	// 注册需要认证的路由
	protectedRoutes := apiV1.Group("/")

	// protectedRoutes.Use(func(c *gin.Context) {
	// 	// 如何匹配路由是否有权限
	// 	c.Next()
	// 	log.Info(c.Route().Name, c.Route().Path, c.Route().Method)

	// 	return nil
	// })
	// protectedRoutes.Use(middlewares.AuthMiddleware())
	// protectedRoutes.Use(middlewares.TenantMiddleware())

	// 注册所有模块路由到受保护的路由组
	for _, module := range r.modules {
		module.RegisterRoutes(protectedRoutes)
	}
}
