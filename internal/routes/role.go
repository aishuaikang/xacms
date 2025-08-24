package routes

import (
	"new-spbatc-drone-platform/internal/models"
	"new-spbatc-drone-platform/internal/routes/dto"
	"new-spbatc-drone-platform/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RoleHandler 角色处理器
type RoleHandler struct {
	RoleService   services.RoleService
	CommonService services.CommonService
}

// RegisterRoutes 注册角色相关路由
func (h *RoleHandler) RegisterRoutes(router fiber.Router) {
	roleGroup := router.Group("/roles").Name("角色管理.")

	roleGroup.Get("", h.GetRoles).Name("获取角色列表")
	roleGroup.Post("", h.CreateRole).Name("创建角色")
	roleGroup.Get("/:id<guid>", h.GetRole).Name("获取角色详情")
	roleGroup.Put("/:id<guid>", h.UpdateRole).Name("更新角色")
	roleGroup.Delete("/:id<guid>", h.DeleteRole).Name("删除角色")
	roleGroup.Get("/:id<guid>/menus", h.GetRoleMenus).Name("获取角色菜单")
	roleGroup.Post("/:id<guid>/menus", h.AssignMenus).Name("分配角色菜单")
}

// GetRoles 获取角色列表
func (h *RoleHandler) GetRoles(c *fiber.Ctx) error {
	var roles []models.RoleModel
	if err := h.CommonService.GetItems(&roles); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse(fiber.StatusInternalServerError, "获取角色列表失败"))
	}
	return c.JSON(dto.SuccessResponse(roles))
}

// CreateRole 创建角色
func (h *RoleHandler) CreateRole(c *fiber.Ctx) error {
	// 解析请求体
	var req dto.CreateRoleRequest
	if err := h.CommonService.ValidateBody(c, &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse(fiber.StatusBadRequest, err.Error()))
	}

	// 创建角色
	role, err := h.RoleService.CreateRole(req)
	if err != nil {
		log.Errorf("创建角色失败: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse(fiber.StatusInternalServerError, "创建角色失败"))
	}

	return c.Status(201).JSON(dto.SuccessResponse(role))
}

// GetRole 获取角色详情
func (h *RoleHandler) GetRole(c *fiber.Ctx) error {
	id := c.Params("id")

	// 验证 UUID 格式
	roleUUID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse(fiber.StatusBadRequest, "角色ID格式无效"))
	}

	// 获取角色
	var role models.RoleModel
	if err := h.CommonService.GetItemByID(roleUUID, &role); err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse(fiber.StatusNotFound, "角色不存在"))
		}
		log.Errorf("获取角色失败: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse(fiber.StatusInternalServerError, "获取角色失败"))
	}

	return c.JSON(dto.SuccessResponse(role))
}

// UpdateRole 更新角色
func (h *RoleHandler) UpdateRole(c *fiber.Ctx) error {
	id := c.Params("id")

	// 验证 UUID 格式
	roleUUID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse(fiber.StatusBadRequest, "角色ID格式无效"))
	}

	// 解析请求体
	var req dto.UpdateRoleRequest
	if err := h.CommonService.ValidateBody(c, &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse(fiber.StatusBadRequest, err.Error()))
	}

	// 更新角色
	role, err := h.RoleService.UpdateRole(roleUUID, req)
	if err != nil {
		log.Errorf("更新角色失败: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse(fiber.StatusInternalServerError, "更新角色失败"))
	}

	return c.JSON(dto.SuccessResponse(role))
}

// DeleteRole 删除角色
func (h *RoleHandler) DeleteRole(c *fiber.Ctx) error {
	id := c.Params("id")

	// 验证 UUID 格式
	roleUUID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse(fiber.StatusBadRequest, "角色ID格式无效"))
	}

	// 删除角色
	if err := h.CommonService.DeleteItemByID(&models.RoleModel{}, roleUUID); err != nil {
		log.Errorf("删除角色失败: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse(fiber.StatusInternalServerError, "删除角色失败"))
	}

	return c.JSON(dto.SuccessResponse(nil))
}

// GetRoleMenus 获取角色菜单
func (h *RoleHandler) GetRoleMenus(c *fiber.Ctx) error {
	id := c.Params("id")

	// 验证 UUID 格式
	roleUUID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse(fiber.StatusBadRequest, "角色ID格式无效"))
	}

	// 获取角色菜单
	menus, err := h.RoleService.GetRoleMenus(roleUUID)
	if err != nil {
		log.Errorf("获取角色菜单失败: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse(fiber.StatusInternalServerError, "获取角色菜单失败"))
	}

	return c.JSON(dto.SuccessResponse(menus))
}

// AssignMenus 分配菜单给角色
func (h *RoleHandler) AssignMenus(c *fiber.Ctx) error {
	id := c.Params("id")

	// 验证 UUID 格式
	roleUUID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse(fiber.StatusBadRequest, "角色ID格式无效"))
	}

	var req dto.AssignMenusRequest
	if err := h.CommonService.ValidateBody(c, &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse(fiber.StatusBadRequest, err.Error()))
	}

	// 分配菜单
	role, err := h.RoleService.AssignMenus(roleUUID, req)
	if err != nil {
		log.Errorf("分配菜单失败: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse(fiber.StatusInternalServerError, "分配菜单失败"))
	}

	return c.JSON(dto.SuccessResponse(role))
}
