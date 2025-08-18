package services

import (
	"new-spbatc-drone-platform/internal/database/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RoleService 角色服务接口
type RoleService interface {
	GetRoles() ([]models.RoleModel, error)
	GetRoleByID(id uuid.UUID) (*models.RoleModel, error)
	CreateRole(role *models.RoleModel) error
	UpdateRole(role *models.RoleModel) error
	DeleteRole(id uuid.UUID) error
}

// roleService 角色服务实现
type roleService struct {
	db *gorm.DB
}

// NewRoleService 创建角色服务实例
func NewRoleService(db *gorm.DB) RoleService {
	return &roleService{
		db: db,
	}
}

// GetRoles 获取角色列表
func (s *roleService) GetRoles() ([]models.RoleModel, error) {
	var roles []models.RoleModel
	if err := s.db.Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

// GetRoleByID 根据ID获取角色
func (s *roleService) GetRoleByID(id uuid.UUID) (*models.RoleModel, error) {
	var role models.RoleModel
	if err := s.db.First(&role, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

// CreateRole 创建角色
func (s *roleService) CreateRole(role *models.RoleModel) error {
	if err := s.db.Create(role).Error; err != nil {
		return err
	}
	return nil
}

// UpdateRole 更新角色
func (s *roleService) UpdateRole(role *models.RoleModel) error {
	if err := s.db.Save(role).Error; err != nil {
		return err
	}
	return nil
}

// DeleteRole 删除角色
func (s *roleService) DeleteRole(id uuid.UUID) error {
	if err := s.db.Delete(&models.RoleModel{}, "id = ?", id).Error; err != nil {
		return err
	}
	return nil
}
