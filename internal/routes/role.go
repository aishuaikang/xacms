package routes

import (
	"net/http"
	"uav_defender/internal/dto"
	"uav_defender/internal/models"
	"uav_defender/internal/services"

	"github.com/gin-gonic/gin"
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
func (h *RoleHandler) RegisterRoutes(router *gin.RouterGroup) {
	roleGroup := router.Group("/roles")

	roleGroup.GET("", h.GetRoles)
	roleGroup.POST("", h.CreateRole)
	roleGroup.GET("/:id", h.GetRole)
	roleGroup.PUT("/:id", h.UpdateRole)
	roleGroup.DELETE("/:id", h.DeleteRole)
	roleGroup.GET("/:id/menus", h.GetRoleMenus)
	roleGroup.POST("/:id/menus", h.AssignMenus)
}

// GetRoles 获取角色列表
func (h *RoleHandler) GetRoles(c *gin.Context) {
	var roles []models.RoleModel
	if err := h.CommonService.GetItems(&roles); err != nil {
		log.Errorf("获取角色列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(http.StatusInternalServerError, "获取角色列表失败"))
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(roles))
}

// CreateRole 创建角色
func (h *RoleHandler) CreateRole(c *gin.Context) {
	// 解析请求体
	var req dto.CreateRoleRequest
	if err := h.CommonService.ValidateBody(c, &req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}

	// 创建角色
	role, err := h.RoleService.CreateRole(req)
	if err != nil {
		log.Errorf("创建角色失败: %v", err)
		// return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse(fiber.StatusInternalServerError, "创建角色失败"))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(http.StatusInternalServerError, "创建角色失败"))
		return
	}

	c.JSON(http.StatusCreated, dto.SuccessResponse(role))
}

// GetRole 获取角色详情
func (h *RoleHandler) GetRole(c *gin.Context) {
	id := c.Param("id")

	// 验证 UUID 格式
	roleUUID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, "角色ID格式无效"))
		return
	}

	// 获取角色
	var role models.RoleModel
	if err := h.CommonService.GetItemByID(roleUUID, &role); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, dto.ErrorResponse(http.StatusNotFound, "角色不存在"))
			return
		}
		log.Errorf("获取角色失败: %v", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(http.StatusInternalServerError, "获取角色失败"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(role))
}

// UpdateRole 更新角色
func (h *RoleHandler) UpdateRole(c *gin.Context) {
	id := c.Param("id")

	// 验证 UUID 格式
	roleUUID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, "角色ID格式无效"))
		return
	}

	// 解析请求体
	var req dto.UpdateRoleRequest
	if err := h.CommonService.ValidateBody(c, &req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}

	// 更新角色
	role, err := h.RoleService.UpdateRole(roleUUID, req)
	if err != nil {
		log.Errorf("更新角色失败: %v", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(http.StatusInternalServerError, "更新角色失败"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(role))
}

// DeleteRole 删除角色
func (h *RoleHandler) DeleteRole(c *gin.Context) {
	id := c.Param("id")

	// 验证 UUID 格式
	roleUUID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, "角色ID格式无效"))
		return
	}

	// 删除角色
	if err := h.CommonService.DeleteItemByID(&models.RoleModel{}, roleUUID); err != nil {
		log.Errorf("删除角色失败: %v", err)
		// return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse(fiber.StatusInternalServerError, "删除角色失败"))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(http.StatusInternalServerError, "删除角色失败"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

// GetRoleMenus 获取角色菜单
func (h *RoleHandler) GetRoleMenus(c *gin.Context) {
	id := c.Param("id")

	// 验证 UUID 格式
	roleUUID, err := uuid.Parse(id)
	if err != nil {
		// return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse(fiber.StatusBadRequest, "角色ID格式无效"))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, "角色ID格式无效"))
		return
	}

	// 获取角色菜单
	menus, err := h.RoleService.GetRoleMenus(roleUUID)
	if err != nil {
		log.Errorf("获取角色菜单失败: %v", err)
		// return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse(fiber.StatusInternalServerError, "获取角色菜单失败"))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(http.StatusInternalServerError, "获取角色菜单失败"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(menus))
}

// AssignMenus 分配菜单给角色
func (h *RoleHandler) AssignMenus(c *gin.Context) {
	id := c.Param("id")

	// 验证 UUID 格式
	roleUUID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, "角色ID格式无效"))
		return
	}

	var req dto.AssignMenusRequest
	if err := h.CommonService.ValidateBody(c, &req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}

	// 分配菜单
	role, err := h.RoleService.AssignMenus(roleUUID, req)
	if err != nil {
		log.Errorf("分配菜单失败: %v", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(http.StatusInternalServerError, "分配菜单失败"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(role))
}
