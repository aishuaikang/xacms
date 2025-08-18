package middlewares

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// AuthMiddleware JWT认证中间件
func AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 获取Authorization头
		authHeader := c.Get("Authorization")

		if authHeader == "" {
			return c.Status(401).JSON(fiber.Map{
				"code":    401,
				"message": "Missing authorization header",
			})
		}

		// 检查Bearer token格式
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(401).JSON(fiber.Map{
				"code":    401,
				"message": "Invalid authorization format",
			})
		}

		// 提取token
		token := strings.TrimPrefix(authHeader, "Bearer ")

		// TODO: 验证JWT token
		// 这里你需要实现JWT token验证逻辑
		if token == "" {
			return c.Status(401).JSON(fiber.Map{
				"code":    401,
				"message": "Invalid token",
			})
		}

		// 将用户信息存储到上下文中
		// c.Locals("user_id", userID)
		// c.Locals("tenant_id", tenantID)

		return c.Next()
	}
}

// TenantMiddleware 多租户中间件
func TenantMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 从头部或子域名获取租户信息
		tenantID := c.Get("X-Tenant-ID")
		if tenantID == "" {
			// 也可以从子域名中提取租户信息
			// host := c.Get("Host")
			// tenantID = extractTenantFromHost(host)
		}

		if tenantID == "" {
			return c.Status(400).JSON(fiber.Map{
				"code":    400,
				"message": "Missing tenant information",
			})
		}

		// 验证租户是否存在且有效
		// TODO: 实现租户验证逻辑

		// 将租户ID存储到上下文中
		c.Locals("tenant_id", tenantID)

		return c.Next()
	}
}

// LoggerMiddleware 自定义日志中间件
func LoggerMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// TODO: 实现自定义日志记录
		// 可以记录请求信息、响应时间等

		return c.Next()
	}
}

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// TODO: 实现限流逻辑
		// 可以基于IP、用户ID或租户ID进行限流

		return c.Next()
	}
}
