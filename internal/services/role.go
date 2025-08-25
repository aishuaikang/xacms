package services

import (
	"errors"
	"xacms/internal/models"
	"xacms/internal/routes/dto"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RoleService 角色服务接口
type RoleService interface {
	CreateRole(req dto.CreateRoleRequest) (*models.RoleModel, error)
	UpdateRole(roleId uuid.UUID, req dto.UpdateRoleRequest) (*models.RoleModel, error)
	GetRoleMenus(roleId uuid.UUID) ([]models.MenuModel, error)
	AssignMenus(roleId uuid.UUID, req dto.AssignMenusRequest) (*models.RoleModel, error)
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
func (s *roleService) CreateRole(req dto.CreateRoleRequest) (*models.RoleModel, error) {
	role := &models.RoleModel{
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
func (s *roleService) UpdateRole(roleId uuid.UUID, req dto.UpdateRoleRequest) (*models.RoleModel, error) {
	var role models.RoleModel
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
func (s *roleService) GetRoleMenus(roleId uuid.UUID) ([]models.MenuModel, error) {
	var menus []models.MenuModel
	if err := s.db.Model(&models.RoleModel{ID: roleId}).Association("Menus").Find(&menus); err != nil {
		return nil, err
	}
	return menus, nil
}

// AssignMenus 分配菜单给角色
func (s *roleService) AssignMenus(roleId uuid.UUID, req dto.AssignMenusRequest) (*models.RoleModel, error) {
	var role models.RoleModel
	if err := s.commonService.GetItemByID(roleId, &role); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("角色不存在")
		}
		return nil, err
	}

	// 获取菜单实例
	var menus []models.MenuModel
	if err := s.db.Where("id IN ?", req.MenuIDs).Find(&menus).Error; err != nil {
		return nil, err
	}

	// 更新角色菜单
	if err := s.db.Model(&role).Association("Menus").Replace(menus); err != nil {
		return nil, err
	}

	return &role, nil
}
