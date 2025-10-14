package routes

import (
	"errors"
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

// MenuRouter 菜单处理器
type MenuRouter struct {
	CommonService services.CommonService
	MenuService   services.MenuService
}

// RegisterRoutes 注册菜单相关路由
func (h *MenuRouter) RegisterRoutes(router *gin.RouterGroup) {

	menuGroup := router.Group("/menus")

	menuGroup.GET("", h.GetMenus)
	menuGroup.POST("", h.CreateMenu)
	menuGroup.GET("/:id", h.GetMenu)
	menuGroup.PUT("/:id", h.UpdateMenu)
	menuGroup.DELETE("/:id", h.DeleteMenu)
	menuGroup.GET("/tree", h.GetMenuTree)
	menuGroup.GET("/apis", h.GetAPIs)

}

// GetMenus 获取菜单列表
func (h *MenuRouter) GetMenus(c *gin.Context) {
	var menus []models.Menu
	if err := h.CommonService.GetItems(&menus); err != nil {
		global.Logger.Error("获取菜单列表失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(http.StatusInternalServerError, "获取菜单列表失败"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(menus))
}

// CreateMenu 创建菜单
func (h *MenuRouter) CreateMenu(c *gin.Context) {
	// 解析请求体
	var req dto.CreateMenuRequest
	if err := h.CommonService.ValidateBody(c, &req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}

	// 创建菜单
	menu, err := h.MenuService.CreateMenu(&req)
	if err != nil {
		// 使用通用的唯一约束错误检查
		if isUniqueConstraintError(err) {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, "菜单已存在"))
			return
		}

		global.Logger.Error("创建菜单失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(http.StatusInternalServerError, "创建菜单失败"))
		return
	}

	c.JSON(http.StatusCreated, dto.SuccessResponse(menu))
}

// GetMenu 获取单个菜单
func (h *MenuRouter) GetMenu(c *gin.Context) {
	id := c.Param("id")

	menuID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, "菜单ID格式无效"))
		return
	}

	// 获取菜单
	var menu models.Menu
	if err := h.CommonService.GetItemByID(menuID, &menu); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse(http.StatusNotFound, "菜单不存在"))
			return
		}
		global.Logger.Error("获取菜单失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(http.StatusInternalServerError, "获取菜单失败"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(menu))
}

// UpdateMenu 更新菜单
func (h *MenuRouter) UpdateMenu(c *gin.Context) {
	id := c.Param("id")

	menuID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, "菜单ID格式无效"))
		return
	}

	// 解析请求体
	var req dto.UpdateMenuRequest
	if err := h.CommonService.ValidateBody(c, &req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}

	// 更新菜单
	menu, err := h.MenuService.UpdateMenu(menuID, &req)
	if err != nil {
		// 使用通用的唯一约束错误检查
		if isUniqueConstraintError(err) {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, "菜单已存在"))
			return
		}

		global.Logger.Error("更新菜单失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(http.StatusInternalServerError, "更新菜单失败"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(menu))
}

// DeleteMenu 删除菜单
func (h *MenuRouter) DeleteMenu(c *gin.Context) {
	id := c.Param("id")

	menuID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, "菜单ID格式无效"))
		return
	}

	// 删除菜单
	if err := h.CommonService.DeleteItemByID(&models.Menu{}, menuID); err != nil {
		global.Logger.Error("删除菜单失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(http.StatusInternalServerError, "删除菜单失败"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

// GetMenuTree 获取菜单树结构
func (h *MenuRouter) GetMenuTree(c *gin.Context) {
	// 组装为树形结构
	menuTree, err := h.MenuService.GetMenuTree()
	if err != nil {
		global.Logger.Error("获取菜单树失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(http.StatusInternalServerError, "获取菜单树失败"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(menuTree))
}

// GetAPIs 获取API列表
func (h *MenuRouter) GetAPIs(c *gin.Context) {
	c.JSON(http.StatusOK, dto.SuccessResponse(h.CommonService.GetAPIs()))
}
