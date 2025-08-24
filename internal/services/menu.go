package services

import (
	"errors"
	"new-spbatc-drone-platform/internal/models"
	"new-spbatc-drone-platform/internal/routes/dto"
	"new-spbatc-drone-platform/internal/utils"
	"sort"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MenuService 菜单服务接口
type MenuService interface {
	GetMenus() ([]models.MenuModel, error)
	CreateMenu(req *dto.CreateMenuRequest) (*models.MenuModel, error)
	UpdateMenu(menuUUID uuid.UUID, req *dto.UpdateMenuRequest) (*models.MenuModel, error)
	GetMenuTree() ([]dto.MenuTreeItem, error)
	GetAPIs(allroutes []fiber.Route) []fiber.Route
}

// menuService 菜单服务实现
type menuService struct {
	db            *gorm.DB
	commonService CommonService
}

// NewMenuService 创建菜单服务实例
func NewMenuService(db *gorm.DB, commonService CommonService) MenuService {
	return &menuService{
		db:            db,
		commonService: commonService,
	}
}

// GetMenus 获取菜单列表
func (s *menuService) GetMenus() ([]models.MenuModel, error) {
	var menus []models.MenuModel
	if err := s.db.Find(&menus).Error; err != nil {
		return nil, err
	}
	return menus, nil
}

// CreateMenu 创建菜单
func (s *menuService) CreateMenu(req *dto.CreateMenuRequest) (*models.MenuModel, error) {

	menu := &models.MenuModel{
		ParentID:     req.ParentID,
		Name:         req.Name,
		RouteName:    req.RouteName,
		RoutePath:    req.RoutePath,
		ApiNames:     req.ApiNames,
		IsHidden:     req.IsHidden,
		IsFullScreen: req.IsFullScreen,
		IsTabs:       req.IsTabs,
		Component:    req.Component,
		Icon:         req.Icon,
		Order:        req.Order,
	}

	if err := s.db.Create(menu).Error; err != nil {
		return nil, err
	}
	return menu, nil
}

// UpdateMenu 更新菜单
func (s *menuService) UpdateMenu(menuUUID uuid.UUID, req *dto.UpdateMenuRequest) (*models.MenuModel, error) {
	var menu models.MenuModel
	if err := s.commonService.GetItemByID(&menu, menuUUID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("菜单不存在")
		}
		return nil, err
	}

	if req.ParentID != nil {
		menu.ParentID = req.ParentID
	}

	if req.Name != nil {
		menu.Name = *req.Name
	}

	if req.RouteName != nil {
		menu.RouteName = *req.RouteName
	}

	if req.RoutePath != nil {
		menu.RoutePath = *req.RoutePath
	}

	if req.ApiNames != nil {
		menu.ApiNames = req.ApiNames
	}
	if req.IsHidden != nil {
		menu.IsHidden = *req.IsHidden
	}
	if req.IsFullScreen != nil {
		menu.IsFullScreen = *req.IsFullScreen
	}
	if req.IsTabs != nil {
		menu.IsTabs = *req.IsTabs
	}
	if req.Component != nil {
		menu.Component = *req.Component
	}
	if req.Icon != nil {
		menu.Icon = req.Icon
	}
	if req.Order != nil {
		menu.Order = *req.Order
	}

	if err := s.db.Save(menu).Error; err != nil {
		return nil, err
	}
	return &menu, nil
}

// GetMenuTree 获取菜单树
func (s *menuService) GetMenuTree() ([]dto.MenuTreeItem, error) {
	menus, err := s.GetMenus()
	if err != nil {
		return nil, err
	}

	// 递归组装菜单树
	var buildMenuTree func(parentID *uuid.UUID) []dto.MenuTreeItem
	buildMenuTree = func(parentID *uuid.UUID) []dto.MenuTreeItem {
		var children []dto.MenuTreeItem
		for _, menu := range menus {
			if utils.EqualUUID(menu.ParentID, parentID) {
				children = append(children, dto.MenuTreeItem{
					MenuModel: menu,
					Children:  buildMenuTree(&menu.ID),
				})
			}
		}
		return children
	}

	return buildMenuTree(nil), nil
}

// GetAPIs 获取API列表
func (s *menuService) GetAPIs(allroutes []fiber.Route) []fiber.Route {
	routeMap := make(map[string][]fiber.Route) // 键: 路径+名称, 值: 具有相同路径+名称的路由

	// 按路径+名称分组路由
	for _, route := range allroutes {
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

	// 排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}
