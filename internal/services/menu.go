package services

import (
	"errors"
	"xacms/internal/models"
	"xacms/internal/routes/dto"
	"xacms/internal/server"
	"xacms/internal/utils"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MenuService 菜单服务接口
type MenuService interface {
	CreateMenu(req *dto.CreateMenuRequest) (*models.MenuModel, error)
	UpdateMenu(menuUUID uuid.UUID, req *dto.UpdateMenuRequest) (*models.MenuModel, error)
	GetMenuTree() ([]dto.MenuTreeItem, error)
}

// menuService 菜单服务实现
type menuService struct {
	db            *gorm.DB
	commonService CommonService
}

// NewMenuService 创建菜单服务实例
func NewMenuService(db *gorm.DB, commonService CommonService, fiberServer *server.FiberServer) MenuService {
	return &menuService{
		db:            db,
		commonService: commonService,
	}
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
func (s *menuService) GetMenuTree() ([]dto.MenuTreeItem, error) {
	var menus []models.MenuModel
	if err := s.commonService.GetItems(&menus); err != nil {
		log.Errorf("获取菜单列表失败: %v", err)
		return nil, errors.New("获取菜单列表失败")
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
