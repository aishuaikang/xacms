package routes

import (
	"new-spbatc-drone-platform/internal/routes/dto"

	"github.com/gofiber/fiber/v2"
)

// UserHandler 用户处理器
type UserHandler struct {
	*BaseHandler
}

// NewUserHandler 创建用户处理器
func NewUserHandler(base *BaseHandler) *UserHandler {
	return &UserHandler{
		BaseHandler: base,
	}
}

// RegisterRoutes 注册用户相关路由
func (h *UserHandler) RegisterRoutes(router fiber.Router) {
	userGroup := router.Group("/users")

	userGroup.Get("/", h.GetUsers)
	userGroup.Post("/", h.CreateUser)
	userGroup.Get("/:id<guid>", h.GetUser)
	userGroup.Put("/:id<guid>", h.UpdateUser)
	userGroup.Delete("/:id<guid>", h.DeleteUser)
}

// GetUsers 获取用户列表
func (h *UserHandler) GetUsers(c *fiber.Ctx) error {
	// 解析查询参数
	var req dto.UserQueryRequest
	if err := c.QueryParser(&req); err != nil {
		return c.Status(400).JSON(dto.ErrorResponse(400, "查询参数格式错误"))
	}

	// 验证查询参数
	if errors := h.Validator.ValidateStruct(&req); len(errors) > 0 {
		return c.Status(400).JSON(dto.ErrorResponse(400, errors[0]))
	}

	// 获取用户列表
	// users, err := h.DB.ServiceManager.UserService.GetUsers(req)
	// if err != nil {
	// 	log.Errorf("获取用户列表失败: %v", err)
	// 	return c.Status(500).JSON(dto.ErrorResponse(500, "获取用户列表失败"))
	// }

	return c.JSON(dto.SuccessResponse(nil))
}

// CreateUser 创建用户
func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	// 解析请求体到 DTO
	var req dto.CreateUserRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(dto.ErrorResponse(400, "请求体格式错误"))
	}

	// 验证请求数据
	if errors := h.Validator.ValidateStruct(&req); len(errors) > 0 {
		return c.Status(400).JSON(dto.ErrorResponse(400, errors[0]))
	}

	// // 创建用户
	// if err := h.DB.ServiceManager.UserService.CreateUser(req); err != nil {
	// 	log.Errorf("创建用户失败: %v", err)
	// 	return c.Status(500).JSON(dto.ErrorResponse(500, "创建用户失败"))
	// }

	return c.Status(201).JSON(dto.SuccessResponse(nil))
}

// GetUser 获取单个用户
func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	// id := c.Params("id")

	// 验证 UUID 格式
	// userUUID, err := uuid.Parse(id)
	// if err != nil {
	// 	return c.Status(400).JSON(dto.ErrorResponse(400, "用户ID格式无效"))
	// }

	// 获取用户
	// user, err := h.DB.ServiceManager.UserService.GetUser(userUUID)
	// log.Infof("err: %+#v", err)
	// if err != nil {
	// 	if err == gorm.ErrRecordNotFound {
	// 		return c.Status(404).JSON(dto.ErrorResponse(404, "用户不存在"))
	// 	}
	// 	log.Errorf("获取用户失败: %v", err)
	// 	return c.Status(500).JSON(dto.ErrorResponse(500, "获取用户失败"))
	// }

	return c.JSON(dto.SuccessResponse(nil))
}

// UpdateUser 更新用户
func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	// id := c.Params("id")

	// 验证 UUID 格式
	// userUUID, err := uuid.Parse(id)
	// if err != nil {
	// 	return c.Status(400).JSON(dto.ErrorResponse(400, "用户ID格式无效"))
	// }

	// 解析请求体
	var req dto.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(dto.ErrorResponse(400, "请求体格式错误"))
	}

	// 验证请求数据
	if errors := h.Validator.ValidateStruct(&req); len(errors) > 0 {
		return c.Status(400).JSON(dto.ErrorResponse(400, errors[0]))
	}

	// 更新用户
	// if err := h.DB.ServiceManager.UserService.UpdateUser(userUUID, req); err != nil {
	// 	log.Errorf("更新用户失败: %v", err)
	// 	return c.Status(500).JSON(dto.ErrorResponse(500, "更新用户失败"))
	// }

	return c.JSON(dto.SuccessResponse(nil))
}

// DeleteUser 删除用户
func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	// id := c.Params("id")

	// 验证 UUID 格式
	// userUUID, err := uuid.Parse(id)
	// if err != nil {
	// 	return c.Status(400).JSON(dto.ErrorResponse(400, "用户ID格式无效"))
	// }

	// 删除用户
	// if err := h.DB.ServiceManager.UserService.DeleteUser(userUUID); err != nil {
	// 	log.Errorf("删除用户失败: %v", err)
	// 	return c.Status(500).JSON(dto.ErrorResponse(500, "删除用户失败"))
	// }

	return c.JSON(dto.SuccessResponse(nil))
}
