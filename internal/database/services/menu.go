package services

import (
	"new-spbatc-drone-platform/internal/database/models"

	"gorm.io/gorm"
)

// MenuService 菜单服务接口
type MenuService interface {
	GetMenus() ([]models.MenuModel, error)
	GetMenuByID(id string) (*models.MenuModel, error)
	CreateMenu(menu *models.MenuModel) error
	UpdateMenu(menu *models.MenuModel) error
	DeleteMenu(id string) error
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
func (s *menuService) CreateMenu(menu *models.MenuModel) error {
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
