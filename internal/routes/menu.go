package routes

import (
	"new-spbatc-drone-platform/internal/routes/dto"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

// MenuHandler 菜单处理器
type MenuHandler struct {
	*BaseHandler
}

// NewMenuHandler 创建菜单处理器
func NewMenuHandler(base *BaseHandler) *MenuHandler {
	return &MenuHandler{
		BaseHandler: base,
	}
}

// RegisterRoutes 注册菜单相关路由
func (h *MenuHandler) RegisterRoutes(router fiber.Router) {
	menuGroup := router.Group("/menus")

	menuGroup.Get("/", h.GetMenus)
	menuGroup.Post("/", h.CreateMenu)
	menuGroup.Get("/:id", h.GetMenu)
	menuGroup.Put("/:id", h.UpdateMenu)
	menuGroup.Delete("/:id", h.DeleteMenu)
	menuGroup.Get("/tree", h.GetMenuTree)
}

// GetMenus 获取菜单列表
func (h *MenuHandler) GetMenus(c *fiber.Ctx) error {
	menus, err := h.DB.ServiceManager.MenuService.GetMenus()
	if err != nil {
		log.Errorf("获取菜单列表失败: %v", err)
		return c.Status(500).JSON(dto.ErrorResponse(500, "获取菜单列表失败"))
	}

	return c.JSON(dto.SuccessResponse(menus))
}

// CreateMenu 创建菜单
func (h *MenuHandler) CreateMenu(c *fiber.Ctx) error {
	var menuData map[string]interface{}
	if err := c.BodyParser(&menuData); err != nil {
		return c.Status(400).JSON(dto.ErrorResponse(400, "Invalid request body"))
	}

	// TODO: 实现创建菜单逻辑
	menuData["id"] = 3

	return c.Status(201).JSON(dto.SuccessResponse(menuData))
}

// GetMenu 获取单个菜单
func (h *MenuHandler) GetMenu(c *fiber.Ctx) error {
	id := c.Params("id")
	menuID, err := strconv.Atoi(id)
	if err != nil {
		return c.Status(400).JSON(dto.ErrorResponse(400, "Invalid menu ID"))
	}

	// TODO: 从数据库获取菜单
	menu := map[string]interface{}{
		"id":        menuID,
		"name":      "Menu " + id,
		"path":      "/menu" + id,
		"parent_id": nil,
	}

	return c.JSON(dto.SuccessResponse(menu))
}

// UpdateMenu 更新菜单
func (h *MenuHandler) UpdateMenu(c *fiber.Ctx) error {
	id := c.Params("id")
	menuID, err := strconv.Atoi(id)
	if err != nil {
		return c.Status(400).JSON(dto.ErrorResponse(400, "Invalid menu ID"))
	}

	var menuData map[string]interface{}
	if err := c.BodyParser(&menuData); err != nil {
		return c.Status(400).JSON(dto.ErrorResponse(400, "Invalid request body"))
	}

	// TODO: 实现更新菜单逻辑
	menuData["id"] = menuID

	return c.JSON(dto.SuccessResponse(menuData))
}

// DeleteMenu 删除菜单
func (h *MenuHandler) DeleteMenu(c *fiber.Ctx) error {
	id := c.Params("id")
	menuID, err := strconv.Atoi(id)
	if err != nil {
		return c.Status(400).JSON(dto.ErrorResponse(400, "Invalid menu ID"))
	}

	// TODO: 实现删除菜单逻辑

	return c.JSON(dto.SuccessResponse(map[string]interface{}{
		"message": "Menu deleted successfully",
		"id":      menuID,
	}))
}

// GetMenuTree 获取菜单树结构
func (h *MenuHandler) GetMenuTree(c *fiber.Ctx) error {
	// TODO: 实现获取菜单树逻辑
	menuTree := []map[string]interface{}{
		{
			"id":   1,
			"name": "Dashboard",
			"path": "/dashboard",
			"children": []map[string]interface{}{
				{"id": 3, "name": "Analytics", "path": "/dashboard/analytics"},
			},
		},
		{
			"id":       2,
			"name":     "Users",
			"path":     "/users",
			"children": []map[string]interface{}{},
		},
	}

	return c.JSON(dto.SuccessResponse(menuTree))
}
