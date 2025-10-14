package services

import (
	"errors"
	"uav_defender/internal/dto"
	"uav_defender/internal/models"
	"uav_defender/internal/pkg/global"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// MenuService 菜单服务接口
type MenuService interface {
	CreateMenu(req *dto.CreateMenuRequest) (*models.Menu, error)
	UpdateMenu(menuUUID uint64, req *dto.UpdateMenuRequest) (*models.Menu, error)
	GetMenuTree() ([]dto.MenuWithChildren, error)
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

// CreateMenu 创建菜单
func (s *menuService) CreateMenu(req *dto.CreateMenuRequest) (*models.Menu, error) {
	// 判断 ParentID 是否存在
	if req.ParentID != nil {
		var parentMenu models.Menu
		if exists, err := s.commonService.IsExistByID(*req.ParentID, &parentMenu); err != nil {
			return nil, err
		} else if !exists {
			return nil, errors.New("父菜单不存在")
		}
	}

	menu := &models.Menu{
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
func (s *menuService) UpdateMenu(menuUUID uint64, req *dto.UpdateMenuRequest) (*models.Menu, error) {
	var menu models.Menu
	if err := s.commonService.GetItemByID(menuUUID, &menu); err != nil {
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
func (s *menuService) GetMenuTree() ([]dto.MenuWithChildren, error) {
	var menus []models.Menu
	if err := s.commonService.GetItems(&menus); err != nil {
		global.Logger.Error("获取菜单列表失败", zap.Error(err))
		return nil, errors.New("获取菜单列表失败")
	}

	// 递归组装菜单树
	var buildMenuTree func(parentID *uint64) []dto.MenuWithChildren
	buildMenuTree = func(parentID *uint64) []dto.MenuWithChildren {
		var children []dto.MenuWithChildren
		for _, menu := range menus {
			if (menu.ParentID == nil && parentID == nil) || (menu.ParentID != nil && parentID != nil && *menu.ParentID == *parentID) {
				children = append(children, dto.MenuWithChildren{
					Menu:     menu,
					Children: buildMenuTree(&menu.ID),
				})
			}
		}
		return children
	}

	return buildMenuTree(nil), nil
}
