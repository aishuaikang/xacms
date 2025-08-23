package services

import (
	"new-spbatc-drone-platform/internal/models"
	"new-spbatc-drone-platform/internal/routes/dto"
	"sort"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// MenuService 菜单服务接口
type MenuService interface {
	GetMenus() ([]models.MenuModel, error)
	GetMenuByID(id string) (*models.MenuModel, error)
	CreateMenu(req *dto.CreateMenuRequest) error
	UpdateMenu(menu *models.MenuModel) error
	DeleteMenu(id string) error
	GetAPIs(allroutes []fiber.Route) []dto.ApiItem
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

// GetAPIs 获取API列表
func (s *menuService) GetAPIs(allroutes []fiber.Route) []dto.ApiItem {
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
			// 只有一个路由，无论方法如何都保留它
			ID := routes[0].Method + "_" + routes[0].Path
			result = append(result, dto.ApiItem{
				ID:     ID,
				Method: routes[0].Method,
				Path:   routes[0].Path,
				Name:   routes[0].Name,
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
					result = append(result, dto.ApiItem{
						ID:     routes[i].Method + "_" + routes[i].Path,
						Method: routes[i].Method,
						Path:   routes[i].Path,
						Name:   routes[i].Name,
					})
				}
			}

			// 如果没有找到非HEAD路由，保留HEAD路由
			if !hasNonHead && headRoute != nil {
				result = append(result, dto.ApiItem{
					ID:     headRoute.Method + "_" + headRoute.Path,
					Method: headRoute.Method,
					Path:   headRoute.Path,
					Name:   headRoute.Name,
				})
			}
		}
	}

	// 排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})
	return result
}
