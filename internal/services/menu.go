package services

import (
	"new-spbatc-drone-platform/internal/models"
	"new-spbatc-drone-platform/internal/routes/dto"
	"new-spbatc-drone-platform/internal/utils"
	"sort"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MenuService 菜单服务接口
type MenuService interface {
	GetMenus() ([]models.MenuModel, error)
	GetMenuByID(id string) (*models.MenuModel, error)
	CreateMenu(req *dto.CreateMenuRequest) error
	UpdateMenu(menu *models.MenuModel) error
	DeleteMenu(id string) error
	GetMenuTree() ([]dto.MenuTreeItem, error)
	GetAPIs(allroutes []fiber.Route, groupName string, nameMap map[string]string) []dto.ApiItem
}

// menuService 菜单服务实现
type menuService struct {
	db *gorm.DB
}

// NewMenuService 创建菜单服务实例
func NewMenuService(db *gorm.DB) MenuService {
	return &menuService{
		db: db,
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

// GetMenuByID 根据ID获取菜单
func (s *menuService) GetMenuByID(id string) (*models.MenuModel, error) {
	var menu models.MenuModel
	if err := s.db.First(&menu, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &menu, nil
}

// CreateMenu 创建菜单
func (s *menuService) CreateMenu(req *dto.CreateMenuRequest) error {

	menu := &models.MenuModel{
		ParentID:     req.ParentID,
		Name:         req.Name,
		RouteName:    req.RouteName,
		RoutePath:    req.RoutePath,
		ApiIds:       req.ApiIds,
		IsHidden:     req.IsHidden,
		IsFullScreen: req.IsFullScreen,
		IsTabs:       req.IsTabs,
		Component:    req.Component,
		Icon:         req.Icon,
		Order:        req.Order,
	}

	if err := s.db.Create(menu).Error; err != nil {
		return err
	}
	return nil
}

// UpdateMenu 更新菜单
func (s *menuService) UpdateMenu(menu *models.MenuModel) error {
	if err := s.db.Save(menu).Error; err != nil {
		return err
	}
	return nil
}

// DeleteMenu 删除菜单
func (s *menuService) DeleteMenu(id string) error {
	if err := s.db.Delete(&models.MenuModel{}, "id = ?", id).Error; err != nil {
		return err
	}
	return nil
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
func (s *menuService) GetAPIs(allroutes []fiber.Route, groupName string, nameMap map[string]string) []dto.ApiItem {
	routeMap := make(map[string][]fiber.Route) // 键: 路径+名称, 值: 具有相同路径+名称的路由

	// 按路径+名称分组路由
	for _, route := range allroutes {
		key := route.Path + "|" + route.Name
		routeMap[key] = append(routeMap[key], route)
	}

	var result []dto.ApiItem
	// 处理每个分组
	for _, routes := range routeMap {
		if len(routes) == 1 {
			name := strings.TrimPrefix(routes[0].Name, groupName)
			log.Infof("保留HEAD路由: %s", name)

			// 只有一个路由，无论方法如何都保留它
			result = append(result, dto.ApiItem{
				ID:     routes[0].Name,
				Method: routes[0].Method,
				Path:   routes[0].Path,
				Name:   nameMap[name],
			})
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
					name := strings.TrimPrefix(routes[i].Name, groupName)
					result = append(result, dto.ApiItem{
						ID:     routes[i].Name,
						Method: routes[i].Method,
						Path:   routes[i].Path,
						Name:   nameMap[name],
					})
				}
			}

			// 如果没有找到非HEAD路由，保留HEAD路由
			if !hasNonHead && headRoute != nil {
				name := strings.TrimPrefix(headRoute.Name, groupName)
				result = append(result, dto.ApiItem{
					ID:     headRoute.Name,
					Method: headRoute.Method,
					Path:   headRoute.Path,
					Name:   nameMap[name],
				})
			}
		}
	}

	// 排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}
