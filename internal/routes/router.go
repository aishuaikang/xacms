package routes

import (
	"new-spbatc-drone-platform/internal/server"

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
	// departmentHandler *DepartmentHandler,
	roleHandler *RoleHandler,
	// tenantHandler *TenantHandler,
) *Router {

	// baseHandler := NewBaseHandler(utils.NewValidationMiddleware())

	return &Router{
		// baseHandler: baseHandler,
		server: server,
		modules: []RouteModule{
			userHandler,
			menuHandler,
			// departmentHandler,
			roleHandler,
			// tenantHandler,
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

// HealthCheck 健康检查端点
// func (r *Router) HealthCheck(c *fiber.Ctx) error {
// 	// TODO: 可以添加数据库连接检查等
// 	return c.JSON(dto.SuccessResponse(map[string]interface{}{
// 		"status":  "ok",
// 		"service": "new-spbatc-drone-platform",
// 		"version": "1.0.0",
// 	}))
// }

// // AddModule 添加新的路由模块
// func (r *Router) AddModule(module RouteModule) {
// 	r.modules = append(r.modules, module)
// }
