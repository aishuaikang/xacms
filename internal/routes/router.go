package routes

import (
	"xacms/internal/server"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

// RouteModule 定义路由模块接口
type RouteModule interface {
	RegisterRoutes(router fiber.Router)
}

// Router 路由注册器
type Router struct {
	server  *server.FiberServer
	modules []RouteModule
}

// NewRouter 创建路由注册器
func NewRouter(server *server.FiberServer,
	userHandler *UserHandler,
	menuHandler *MenuHandler,
	roleHandler *RoleHandler,
) *Router {
	return &Router{
		server: server,
		modules: []RouteModule{
			userHandler,
			menuHandler,
			roleHandler,
		},
	}
}

// RegisterRoutes 注册所有模块路由
func (r *Router) RegisterRoutes() {
	// 创建 API 版本组
	apiV1 := r.server.App.Group("/api/v1")

	// 注册公开路由（不需要认证）
	// publicRoutes := apiV1.Group("/public")
	// publicRoutes.Get("/health", r.HealthCheck)

	// 注册需要认证的路由
	protectedRoutes := apiV1.Group("/")

	protectedRoutes.Use(func(c *fiber.Ctx) error {
		// 如何匹配路由是否有权限
		c.Next()
		log.Info(c.Route().Name, c.Route().Path, c.Route().Method)

		return nil
	})
	// protectedRoutes.Use(middlewares.AuthMiddleware())
	// protectedRoutes.Use(middlewares.TenantMiddleware())

	// 注册所有模块路由到受保护的路由组
	for _, module := range r.modules {
		module.RegisterRoutes(protectedRoutes)
	}
}
