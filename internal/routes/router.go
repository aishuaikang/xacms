package routes

import (
	"new-spbatc-drone-platform/internal/database"
	"new-spbatc-drone-platform/internal/routes/dto"
	"new-spbatc-drone-platform/internal/server"
	"new-spbatc-drone-platform/internal/utils"

	"github.com/gofiber/fiber/v2"
)

// RouteModule 定义路由模块接口
type RouteModule interface {
	RegisterRoutes(router fiber.Router)
}

// BaseHandler 基础处理器，包含通用依赖
type BaseHandler struct {
	DB        database.Service
	Validator *utils.ValidationMiddleware
}

// NewBaseHandler 创建基础处理器
func NewBaseHandler(db database.Service, validator *utils.ValidationMiddleware) *BaseHandler {
	return &BaseHandler{
		DB:        db,
		Validator: validator,
	}
}

// Router 路由注册器
type Router struct {
	baseHandler *BaseHandler
	modules     []RouteModule
	server      *server.FiberServer
}

// NewRouter 创建路由注册器
func NewRouter(server *server.FiberServer) *Router {

	baseHandler := NewBaseHandler(database.New(), utils.NewValidationMiddleware())

	return &Router{
		baseHandler: baseHandler,
		modules: []RouteModule{
			NewUserHandler(baseHandler),
			NewMenuHandler(baseHandler),
			NewDepartmentHandler(baseHandler),
			NewRoleHandler(baseHandler),
			NewTenantHandler(baseHandler),
		},
		server: server,
	}
}

// RegisterRoutes 注册所有模块路由
func (r *Router) RegisterRoutes() {
	// 创建 API 版本组
	apiV1 := r.server.App.Group("/api/v1")

	// 注册公开路由（不需要认证）
	publicRoutes := apiV1.Group("/public")
	publicRoutes.Get("/health", r.HealthCheck)

	// 注册需要认证的路由
	protectedRoutes := apiV1.Group("/")
	// protectedRoutes.Use(middlewares.AuthMiddleware())
	// protectedRoutes.Use(middlewares.TenantMiddleware())

	// 注册所有模块路由到受保护的路由组
	for _, module := range r.modules {
		module.RegisterRoutes(protectedRoutes)
	}
}

// HealthCheck 健康检查端点
func (r *Router) HealthCheck(c *fiber.Ctx) error {
	// TODO: 可以添加数据库连接检查等
	return c.JSON(dto.SuccessResponse(map[string]interface{}{
		"status":  "ok",
		"service": "new-spbatc-drone-platform",
		"version": "1.0.0",
	}))
}

// AddModule 添加新的路由模块
func (r *Router) AddModule(module RouteModule) {
	r.modules = append(r.modules, module)
}
