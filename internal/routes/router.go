package routes

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// Route 定义路由接口
type Route interface {
	RegisterRoutes(router *gin.RouterGroup)
}

// Router 路由器
type Router struct {
	routes       []Route
	publicRoutes []Route
}

// NewRouter 创建路由器
func NewRouter(
	userRouter *UserRouter,
	menuRouter *MenuRouter,
	roleRouter *RoleRouter,
	userPublicRouter *UserPublicRouter,
) *Router {

	return &Router{
		routes: []Route{
			userRouter,
			menuRouter,
			roleRouter,
		},
		publicRoutes: []Route{
			userPublicRouter,
		},
	}

}

// RegisterRoutes 注册所有模块路由
func (r *Router) RegisterRoutes(router *gin.RouterGroup) {

	// 注册公开路由（不需要认证）
	publicRoutes := router.Group("/public")

	// 注册需要认证的路由
	protectedRoutes := router.Group("/")

	// protectedRoutes.Use(func(c *gin.Context) {
	// 	// 如何匹配路由是否有权限
	// 	c.Next()
	// 	log.Info(c.Route().Name, c.Route().Path, c.Route().Method)

	// 	return nil
	// })
	// protectedRoutes.Use(middlewares.AuthMiddleware())
	// protectedRoutes.Use(middlewares.TenantMiddleware())

	// 注册所有模块路由到公开路由组
	for _, module := range r.publicRoutes {
		module.RegisterRoutes(publicRoutes)
	}

	// 注册所有模块路由到受保护的路由组
	for _, module := range r.routes {
		module.RegisterRoutes(protectedRoutes)
	}

}

// isUniqueConstraintError 检查是否为唯一约束错误
func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}

	// 检查常见的唯一约束错误消息
	errorMsg := strings.ToLower(err.Error())
	uniqueConstraintIndicators := []string{
		"unique constraint",
		"duplicate entry",
		"duplicate key",
		"1062",  // MySQL 重复键错误代码
		"23505", // PostgreSQL 唯一约束错误代码
		"2067",  // SQLite 唯一约束错误代码
	}

	for _, indicator := range uniqueConstraintIndicators {
		if strings.Contains(errorMsg, strings.ToLower(indicator)) {
			return true
		}
	}

	return false
}
