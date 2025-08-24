package routes

import (
	"new-spbatc-drone-platform/internal/models"
	"new-spbatc-drone-platform/internal/routes/dto"
	"new-spbatc-drone-platform/internal/services"

	"github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MenuHandler 菜单处理器
type MenuHandler struct {
	CommonService services.CommonService
	MenuService   services.MenuService
}

// RegisterRoutes 注册菜单相关路由
func (h *MenuHandler) RegisterRoutes(router fiber.Router) {

	menuGroup := router.Group("/menus").Name("菜单管理.")

	menuGroup.Get("", h.GetMenus).Name("获取菜单列表")
	menuGroup.Post("", h.CreateMenu).Name("创建菜单")
	menuGroup.Get("/:id<guid>", h.GetMenu).Name("获取菜单详情")
	menuGroup.Put("/:id<guid>", h.UpdateMenu).Name("更新菜单")
	menuGroup.Delete("/:id<guid>", h.DeleteMenu).Name("删除菜单")
	menuGroup.Get("/tree", h.GetMenuTree).Name("获取菜单树")
	menuGroup.Get("/apis", h.GetAPIs).Name("获取API列表")

}

// GetMenus 获取菜单列表
func (h *MenuHandler) GetMenus(c *fiber.Ctx) error {
	var menus []models.MenuModel
	if err := h.CommonService.GetItems(&menus); err != nil {
		log.Errorf("获取菜单列表失败: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse(fiber.StatusInternalServerError, "获取菜单列表失败"))
	}

	return c.JSON(dto.SuccessResponse(menus))
}

// CreateMenu 创建菜单
func (h *MenuHandler) CreateMenu(c *fiber.Ctx) error {
	// 解析请求体
	var req dto.CreateMenuRequest
	if err := h.CommonService.ValidateBody(c, &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse(fiber.StatusBadRequest, err.Error()))
	}

	// 创建菜单
	menu, err := h.MenuService.CreateMenu(&req)
	if err != nil {
		log.Errorf("创建菜单失败: %#v", err)

		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			if mysqlErr.Number == 1062 {
				return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse(fiber.StatusBadRequest, "菜单已存在"))
			}
		}

		log.Errorf("创建菜单失败: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse(fiber.StatusInternalServerError, "创建菜单失败"))
	}

	return c.Status(fiber.StatusCreated).JSON(dto.SuccessResponse(menu))
}

// GetMenu 获取单个菜单
func (h *MenuHandler) GetMenu(c *fiber.Ctx) error {
	id := c.Params("id")

	// 验证 UUID 格式
	menuUUID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse(fiber.StatusBadRequest, "菜单ID格式无效"))
	}

	// 获取菜单
	var menu models.MenuModel
	if err := h.CommonService.GetItemByID(menuUUID, &menu); err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse(fiber.StatusNotFound, "菜单不存在"))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse(fiber.StatusInternalServerError, "获取菜单失败"))
	}

	return c.JSON(dto.SuccessResponse(menu))
}

// UpdateMenu 更新菜单
func (h *MenuHandler) UpdateMenu(c *fiber.Ctx) error {
	id := c.Params("id")

	// 验证 UUID 格式
	menuUUID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse(fiber.StatusBadRequest, "菜单ID格式无效"))
	}

	// 解析请求体
	var req dto.UpdateMenuRequest
	if err := h.CommonService.ValidateBody(c, &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse(fiber.StatusBadRequest, err.Error()))
	}

	// 更新菜单
	menu, err := h.MenuService.UpdateMenu(menuUUID, &req)
	if err != nil {
		log.Errorf("更新菜单失败: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse(fiber.StatusInternalServerError, "更新菜单失败"))
	}

	return c.JSON(dto.SuccessResponse(menu))
}

// DeleteMenu 删除菜单
func (h *MenuHandler) DeleteMenu(c *fiber.Ctx) error {
	id := c.Params("id")

	// 验证 UUID 格式
	menuUUID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse(fiber.StatusBadRequest, "菜单ID格式无效"))
	}

	// 删除菜单
	if err := h.CommonService.DeleteItemByID(&models.MenuModel{}, menuUUID); err != nil {
		log.Errorf("删除菜单失败: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse(fiber.StatusInternalServerError, "删除菜单失败"))
	}

	return c.JSON(dto.SuccessResponse(nil))
}

// GetMenuTree 获取菜单树结构
func (h *MenuHandler) GetMenuTree(c *fiber.Ctx) error {
	// 组装为树形结构
	menuTree, err := h.MenuService.GetMenuTree()
	if err != nil {
		log.Errorf("获取菜单树失败: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse(fiber.StatusInternalServerError, "获取菜单树失败"))
	}

	return c.JSON(dto.SuccessResponse(menuTree))
}

// GetAPIs 获取API列表
func (h *MenuHandler) GetAPIs(c *fiber.Ctx) error {
	return c.JSON(dto.SuccessResponse(h.CommonService.GetAPIs()))
}
