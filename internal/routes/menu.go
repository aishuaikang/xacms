package routes

import (
	"net/http"
	"uav_defender/internal/dto"
	"uav_defender/internal/models"
	"uav_defender/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"github.com/mattn/go-sqlite3"
	"gorm.io/gorm"
)

// MenuHandler 菜单处理器
type MenuHandler struct {
	CommonService services.CommonService
	MenuService   services.MenuService
}

// RegisterRoutes 注册菜单相关路由
func (h *MenuHandler) RegisterRoutes(router *gin.RouterGroup) {

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
func (h *MenuHandler) GetMenus(c *gin.Context) {
	var menus []models.MenuModel
	if err := h.CommonService.GetItems(&menus); err != nil {
		log.Errorf("获取菜单列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(http.StatusInternalServerError, "获取菜单列表失败"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(menus))
}

// CreateMenu 创建菜单
func (h *MenuHandler) CreateMenu(c *gin.Context) {
	// 解析请求体
	var req dto.CreateMenuRequest
	if err := h.CommonService.ValidateBody(c, &req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}

	// 创建菜单
	menu, err := h.MenuService.CreateMenu(&req)
	if err != nil {
		log.Errorf("创建菜单失败: %v", err)
		if sqliteErr, ok := err.(sqlite3.Error); ok {
			if sqliteErr.Code == sqlite3.ErrConstraint {
				c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, "菜单已存在"))
				return
			}
		}

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(http.StatusInternalServerError, "创建菜单失败"))
		return
	}

	c.JSON(http.StatusCreated, dto.SuccessResponse(menu))
}

// GetMenu 获取单个菜单
func (h *MenuHandler) GetMenu(c *gin.Context) {
	id := c.Param("id")

	// 验证 UUID 格式
	menuUUID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, "菜单ID格式无效"))
		return
	}

	// 获取菜单
	var menu models.MenuModel
	if err := h.CommonService.GetItemByID(menuUUID, &menu); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, dto.ErrorResponse(http.StatusNotFound, "菜单不存在"))
			return
		}
		log.Errorf("获取菜单失败: %v", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(http.StatusInternalServerError, "获取菜单失败"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(menu))
}

// UpdateMenu 更新菜单
func (h *MenuHandler) UpdateMenu(c *gin.Context) {
	id := c.Param("id")

	// 验证 UUID 格式
	menuUUID, err := uuid.Parse(id)
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
	menu, err := h.MenuService.UpdateMenu(menuUUID, &req)
	if err != nil {
		log.Errorf("更新菜单失败: %v", err)
		if sqliteErr, ok := err.(sqlite3.Error); ok {
			if sqliteErr.Code == sqlite3.ErrConstraint {
				c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, "菜单已存在"))
				return
			}
		}
		log.Errorf("更新菜单失败: %v", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(http.StatusInternalServerError, "更新菜单失败"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(menu))
}

// DeleteMenu 删除菜单
func (h *MenuHandler) DeleteMenu(c *gin.Context) {
	id := c.Param("id")

	// 验证 UUID 格式
	menuUUID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, "菜单ID格式无效"))
		return
	}

	// 删除菜单
	if err := h.CommonService.DeleteItemByID(&models.MenuModel{}, menuUUID); err != nil {
		log.Errorf("删除菜单失败: %v", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(http.StatusInternalServerError, "删除菜单失败"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

// GetMenuTree 获取菜单树结构
func (h *MenuHandler) GetMenuTree(c *gin.Context) {
	// 组装为树形结构
	menuTree, err := h.MenuService.GetMenuTree()
	if err != nil {
		log.Errorf("获取菜单树失败: %v", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(http.StatusInternalServerError, "获取菜单树失败"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(menuTree))
}

// GetAPIs 获取API列表
func (h *MenuHandler) GetAPIs(c *gin.Context) {
	c.JSON(http.StatusOK, dto.SuccessResponse(h.CommonService.GetAPIs()))
}
