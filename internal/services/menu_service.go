package services

import (
	"errors"
	"uav_defender/internal/dto"
	"uav_defender/internal/models"
	"uav_defender/internal/pkg/global"
	"uav_defender/internal/pkg/utils"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// MenuService 菜单服务接口
type MenuService interface {
	CreateMenu(req *dto.CreateMenuRequest) (*models.MenuModel, error)
	UpdateMenu(menuUUID uuid.UUID, req *dto.UpdateMenuRequest) (*models.MenuModel, error)
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
func (s *menuService) CreateMenu(req *dto.CreateMenuRequest) (*models.MenuModel, error) {
	// 判断 ParentID 是否存在
	if req.ParentID != nil {
		var parentMenu models.MenuModel
		if exists, err := s.commonService.IsExistByID(*req.ParentID, &parentMenu); err != nil {
			return nil, err
		} else if !exists {
			return nil, errors.New("父菜单不存在")
		}
	}

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
func (s *menuService) GetMenuTree() ([]dto.MenuWithChildren, error) {
	var menus []models.MenuModel
	if err := s.commonService.GetItems(&menus); err != nil {
		global.Logger.Error("获取菜单列表失败", zap.Error(err))
		return nil, errors.New("获取菜单列表失败")
	}

	// global.Logger.Debugf("所有菜单: %+v", menus)

	// 递归组装菜单树
	var buildMenuTree func(parentID *uuid.UUID) []dto.MenuWithChildren
	buildMenuTree = func(parentID *uuid.UUID) []dto.MenuWithChildren {
		var children []dto.MenuWithChildren
		for _, menu := range menus {
			global.Logger.Debug("检查菜单父ID", zap.Any("菜单ID", menu.ID), zap.Any("菜单父ID", menu.ParentID), zap.Any("当前父ID", parentID))
			if utils.EqualUUID(menu.ParentID, parentID) {
				children = append(children, dto.MenuWithChildren{
					MenuModel: menu,
					Children:  buildMenuTree(&menu.ID),
				})
			}
		}
		return children
	}

	return buildMenuTree(nil), nil
}
