package services

import (
	"errors"
	"fmt"
	"uav_defender/internal/dto"
	"uav_defender/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// RoleService 角色服务接口
type RoleService interface {
	CreateRole(req dto.CreateRoleRequest) (*models.Role, error)
	UpdateRole(roleId uint64, req dto.UpdateRoleRequest) (*models.Role, error)
	DeleteRole(roleId uint64) error
	GetRoleMenus(roleId uint64) ([]models.Menu, error)
	AssignMenus(roleId uint64, req dto.AssignMenusRequest) (*models.Role, error)
	IsRoleExist(roleId uint64) (bool, error)
}

// roleService 角色服务实现
type roleService struct {
	db            *gorm.DB
	commonService CommonService
}

// NewRoleService 创建角色服务实例
func NewRoleService(db *gorm.DB, commonService CommonService) RoleService {
	return &roleService{
		db:            db,
		commonService: commonService,
	}
}

// CreateRole 创建角色
func (s *roleService) CreateRole(req dto.CreateRoleRequest) (*models.Role, error) {
	role := &models.Role{
		Name:        req.Name,
		Description: req.Description,
		Order:       req.Order,
	}
	if err := s.db.Create(role).Error; err != nil {
		return nil, err
	}
	return role, nil
}

// UpdateRole 更新角色
func (s *roleService) UpdateRole(roleId uint64, req dto.UpdateRoleRequest) (*models.Role, error) {
	var role models.Role
	if err := s.commonService.GetItemByID(roleId, &role); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("角色不存在")
		}
		return nil, err
	}

	if req.Name != nil {
		role.Name = *req.Name
	}

	if req.Description != nil {
		role.Description = *req.Description
	}

	if req.Order != nil {
		role.Order = *req.Order
	}

	if err := s.db.Save(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

// GetRoleMenus 获取角色菜单列表
func (s *roleService) GetRoleMenus(roleId uint64) ([]models.Menu, error) {
	var menus []models.Menu
	if err := s.db.Model(&models.Role{ID: roleId}).Association("Menus").Find(&menus); err != nil {
		return nil, err
	}
	return menus, nil
}

// AssignMenus 分配菜单给角色
func (s *roleService) AssignMenus(roleId uint64, req dto.AssignMenusRequest) (*models.Role, error) {
	var role models.Role
	if err := s.commonService.GetItemByID(roleId, &role); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("角色不存在")
		}
		return nil, err
	}

	// 获取菜单实例
	var menus []models.Menu
	if err := s.db.Where("id IN ?", req.MenuIDs).Find(&menus).Error; err != nil {
		return nil, err
	}

	// 更新角色菜单
	if err := s.db.Model(&role).Association("Menus").Replace(menus); err != nil {
		return nil, err
	}

	return &role, nil
}

// IsRoleExist 检查角色是否存在
func (s *roleService) IsRoleExist(roleId uint64) (bool, error) {
	var count int64
	if err := s.db.Model(&models.Role{}).Where("id = ?", roleId).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// DeleteRole 删除角色
func (s *roleService) DeleteRole(roleId uint64) error {
	var role models.Role
	if err := s.db.Preload(clause.Associations).First(&role, roleId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("角色不存在")
		}
		return err
	}

	// 检查是否有关联的用户
	if len(role.Users) > 0 {
		return fmt.Errorf("无法删除有关联用户的角色,有 %d 个用户关联了该角色,请先解除用户关联", len(role.Users))
	}

	// 检查是否有关联的菜单
	if len(role.Menus) > 0 {
		return fmt.Errorf("无法删除有关联菜单的角色,有 %d 个菜单关联了该角色,请先解除菜单关联", len(role.Menus))
	}

	return s.db.Delete(&role).Error
}
