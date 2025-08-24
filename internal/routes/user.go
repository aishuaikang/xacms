package routes

import (
	"new-spbatc-drone-platform/internal/routes/dto"
	"new-spbatc-drone-platform/internal/services"
	"new-spbatc-drone-platform/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserHandler 用户处理器
type UserHandler struct {
	Validator   *utils.ValidationMiddleware
	UserService services.UserService
}

// RegisterRoutes 注册用户相关路由
func (h *UserHandler) RegisterRoutes(router fiber.Router) {
	userGroup := router.Group("/users").Name("用户管理.")

	userGroup.Get("", h.GetUsers).Name("获取用户列表")
	userGroup.Post("", h.CreateUser).Name("创建用户")
	userGroup.Get("/:id<guid>", h.GetUser).Name("获取用户详情")
	userGroup.Put("/:id<guid>", h.UpdateUser).Name("更新用户")
	userGroup.Delete("/:id<guid>", h.DeleteUser).Name("删除用户")
}

// GetUsers 获取用户列表
func (h *UserHandler) GetUsers(c *fiber.Ctx) error {
	// 解析查询参数
	var req dto.UserQueryRequest
	if err := c.QueryParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse(fiber.StatusBadRequest, "查询参数格式错误"))
	}

	// 验证查询参数
	if errors := h.Validator.ValidateStruct(&req); len(errors) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse(fiber.StatusBadRequest, errors[0]))
	}

	// 获取用户列表
	users, err := h.UserService.GetUsers(req)
	if err != nil {
		log.Errorf("获取用户列表失败: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse(fiber.StatusInternalServerError, "获取用户列表失败"))
	}

	return c.JSON(dto.SuccessResponse(users))
}

// CreateUser 创建用户
func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	// 解析请求体到 DTO
	var req dto.CreateUserRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse(fiber.StatusBadRequest, "请求体格式错误"))
	}

	// 验证请求数据
	if errors := h.Validator.ValidateStruct(&req); len(errors) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse(fiber.StatusBadRequest, errors[0]))
	}

	// 创建用户
	user, err := h.UserService.CreateUser(req)
	if err != nil {
		log.Errorf("创建用户失败: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse(fiber.StatusInternalServerError, "创建用户失败"))
	}

	return c.Status(fiber.StatusCreated).JSON(dto.SuccessResponse(user))
}

// GetUser 获取单个用户
func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	id := c.Params("id")

	// 验证 UUID 格式
	userUUID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse(fiber.StatusBadRequest, "用户ID格式无效"))
	}

	// 获取用户
	user, err := h.UserService.GetUser(userUUID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse(fiber.StatusNotFound, "用户不存在"))
		}
		log.Errorf("获取用户失败: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse(fiber.StatusInternalServerError, "获取用户失败"))
	}

	return c.JSON(dto.SuccessResponse(user))
}

// UpdateUser 更新用户
func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	id := c.Params("id")

	// 验证 UUID 格式
	userUUID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse(fiber.StatusBadRequest, "用户ID格式无效"))
	}

	// 解析请求体
	var req dto.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse(fiber.StatusBadRequest, "请求体格式错误"))
	}

	// 验证请求数据
	if errors := h.Validator.ValidateStruct(&req); len(errors) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse(fiber.StatusBadRequest, errors[0]))
	}

	// 更新用户
	user, err := h.UserService.UpdateUser(userUUID, req)
	if err != nil {
		log.Errorf("更新用户失败: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse(fiber.StatusInternalServerError, "更新用户失败"))
	}

	return c.JSON(dto.SuccessResponse(user))
}

// DeleteUser 删除用户
func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")

	// 验证 UUID 格式
	userUUID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse(fiber.StatusBadRequest, "用户ID格式无效"))
	}

	// 删除用户
	if err := h.UserService.DeleteUser(userUUID); err != nil {
		log.Errorf("删除用户失败: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse(fiber.StatusInternalServerError, "删除用户失败"))
	}

	return c.JSON(dto.SuccessResponse(nil))
}
