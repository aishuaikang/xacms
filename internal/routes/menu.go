package routes

import (
	"new-spbatc-drone-platform/internal/routes/dto"
	"new-spbatc-drone-platform/internal/server"
	"new-spbatc-drone-platform/internal/services"
	"new-spbatc-drone-platform/internal/utils"
	"strconv"

	"github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

// MenuHandler 菜单处理器
type MenuHandler struct {
	Validator   *utils.ValidationMiddleware
	MenuService services.MenuService
	Server      *server.FiberServer
}

// RegisterRoutes 注册菜单相关路由
func (h *MenuHandler) RegisterRoutes(router fiber.Router) {
	menuGroup := router.Group("/menus")

	menuGroup.Get("/", h.GetMenus).Name("获取菜单列表")
	menuGroup.Post("/", h.CreateMenu).Name("创建菜单")
	menuGroup.Get("/:id<guid>", h.GetMenu).Name("获取菜单详情")
	menuGroup.Put("/:id<guid>", h.UpdateMenu).Name("更新菜单")
	menuGroup.Delete("/:id<guid>", h.DeleteMenu).Name("删除菜单")
	menuGroup.Get("/tree", h.GetMenuTree).Name("获取菜单树")
	menuGroup.Get("/apis", h.GetAPIs).Name("获取API列表")

}

// GetMenus 获取菜单列表
func (h *MenuHandler) GetMenus(c *fiber.Ctx) error {
	menus, err := h.MenuService.GetMenus()
	if err != nil {
		log.Errorf("获取菜单列表失败: %v", err)
		return c.Status(500).JSON(dto.ErrorResponse(500, "获取菜单列表失败"))
	}

	return c.JSON(dto.SuccessResponse(menus))
}

// CreateMenu 创建菜单
func (h *MenuHandler) CreateMenu(c *fiber.Ctx) error {
	// 解析请求体到 DTO
	var req dto.CreateMenuRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(dto.ErrorResponse(400, "请求体格式错误"))
	}

	// 验证请求数据
	if errors := h.Validator.ValidateStruct(&req); len(errors) > 0 {
		return c.Status(400).JSON(dto.ErrorResponse(400, errors[0]))
	}

	// 创建菜单
	if err := h.MenuService.CreateMenu(&req); err != nil {
		log.Errorf("创建菜单失败: %#v", err)

		// 我如何捕获这个错误并返回一个友好的错误消息？
		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			if mysqlErr.Number == 1062 {
				return c.Status(400).JSON(dto.ErrorResponse(400, "菜单已存在"))
			}
		}

		log.Errorf("创建菜单失败: %v", err)
		return c.Status(500).JSON(dto.ErrorResponse(500, "创建菜单失败"))
	}

	return c.Status(201).JSON(dto.SuccessResponse(req))
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

// GetAPIs 获取API列表
func (h *MenuHandler) GetAPIs(c *fiber.Ctx) error {
	allRoutes := h.Server.GetRoutes(true)
	routeMap := make(map[string][]fiber.Route) // 键: 路径+名称, 值: 具有相同路径+名称的路由

	// 按路径+名称分组路由
	for _, route := range allRoutes {
		key := route.Path + "|" + route.Name
		routeMap[key] = append(routeMap[key], route)
	}

	var result []fiber.Route

	// 处理每个分组
	for _, routes := range routeMap {
		if len(routes) == 1 {
			// 只有一个路由，无论方法如何都保留它
			result = append(result, routes[0])
		} else {
			// 具有相同路径+名称的多个路由
			hasNonHead := false
			var headRoute *fiber.Route

			for i := range routes {
				if routes[i].Method == fiber.MethodHead {
					if headRoute == nil {
						headRoute = &routes[i]
					}
				} else {
					hasNonHead = true
					result = append(result, routes[i])
				}
			}

			// 如果没有找到非HEAD路由，保留HEAD路由
			if !hasNonHead && headRoute != nil {
				result = append(result, *headRoute)
			}
		}
	}

	return c.JSON(dto.SuccessResponse(result))
}
