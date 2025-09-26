package routes

import (
	"net/http"
	"strconv"
	"uav_defender/internal/dto"
	"uav_defender/internal/models"
	"uav_defender/internal/pkg/global"
	"uav_defender/internal/services"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"gorm.io/gorm"
)

// RoleRouter 角色处理器
type RoleRouter struct {
	RoleService   services.RoleService
	CommonService services.CommonService
}

// RegisterRoutes 注册角色相关路由
func (h *RoleRouter) RegisterRoutes(router *gin.RouterGroup) {
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
func (h *RoleRouter) GetRoles(c *gin.Context) {
	var roles []models.Role
	if err := h.CommonService.GetItems(&roles); err != nil {
		global.Logger.Error("获取角色列表失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(http.StatusInternalServerError, "获取角色列表失败"))
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(roles))
}

// CreateRole 创建角色
func (h *RoleRouter) CreateRole(c *gin.Context) {
	// 解析请求体
	var req dto.CreateRoleRequest
	if err := h.CommonService.ValidateBody(c, &req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}

	// 创建角色
	role, err := h.RoleService.CreateRole(req)
	if err != nil {
		global.Logger.Error("创建角色失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(http.StatusInternalServerError, "创建角色失败"))
		return
	}

	c.JSON(http.StatusCreated, dto.SuccessResponse(role))
}

// GetRole 获取角色详情
func (h *RoleRouter) GetRole(c *gin.Context) {
	id := c.Param("id")

	// 验证 ID 格式
	roleID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, "角色ID格式无效"))
		return
	}

	// 获取角色
	var role models.Role
	if err := h.CommonService.GetItemByID(uint(roleID), &role); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, dto.ErrorResponse(http.StatusNotFound, "角色不存在"))
			return
		}
		global.Logger.Error("获取角色失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(http.StatusInternalServerError, "获取角色失败"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(role))
}

// UpdateRole 更新角色
func (h *RoleRouter) UpdateRole(c *gin.Context) {
	id := c.Param("id")

	// 验证 ID 格式
	roleID, err := strconv.ParseUint(id, 10, 64)
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
	role, err := h.RoleService.UpdateRole(uint(roleID), req)
	if err != nil {
		global.Logger.Error("更新角色失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(http.StatusInternalServerError, "更新角色失败"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(role))
}

// DeleteRole 删除角色
func (h *RoleRouter) DeleteRole(c *gin.Context) {
	id := c.Param("id")

	// 验证 ID 格式
	roleID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, "角色ID格式无效"))
		return
	}

	// 删除角色
	if err := h.CommonService.DeleteItemByID(&models.Role{}, uint(roleID)); err != nil {
		global.Logger.Error("删除角色失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(http.StatusInternalServerError, "删除角色失败"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

// GetRoleMenus 获取角色菜单
func (h *RoleRouter) GetRoleMenus(c *gin.Context) {
	id := c.Param("id")

	// 验证 ID 格式
	roleID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, "角色ID格式无效"))
		return
	}

	// 获取角色菜单
	menus, err := h.RoleService.GetRoleMenus(uint(roleID))
	if err != nil {
		global.Logger.Error("获取角色菜单失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(http.StatusInternalServerError, "获取角色菜单失败"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(menus))
}

// AssignMenus 分配菜单给角色
func (h *RoleRouter) AssignMenus(c *gin.Context) {
	id := c.Param("id")

	// 验证 ID 格式
	roleID, err := strconv.ParseUint(id, 10, 64)
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
	role, err := h.RoleService.AssignMenus(uint(roleID), req)
	if err != nil {
		global.Logger.Error("分配菜单失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(http.StatusInternalServerError, "分配菜单失败"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(role))
}
