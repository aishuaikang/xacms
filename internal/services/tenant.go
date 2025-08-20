package services

import (
	"new-spbatc-drone-platform/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TenantService 租户服务接口
type TenantService interface {
	GetTenants() ([]models.TenantModel, error)
	GetTenantByID(id uuid.UUID) (*models.TenantModel, error)
	CreateTenant(tenant *models.TenantModel) error
	UpdateTenant(tenant *models.TenantModel) error
	DeleteTenant(id uuid.UUID) error
}

// tenantService 租户服务实现
type tenantService struct {
	db *gorm.DB
}

// NewTenantService 创建租户服务实例
func NewTenantService(db *gorm.DB) TenantService {
	return &tenantService{
		db: db,
	}
}

// GetTenants 获取租户列表
func (s *tenantService) GetTenants() ([]models.TenantModel, error) {
	var tenants []models.TenantModel
	if err := s.db.Find(&tenants).Error; err != nil {
		return nil, err
	}
	return tenants, nil
}

// GetTenantByID 根据ID获取租户
func (s *tenantService) GetTenantByID(id uuid.UUID) (*models.TenantModel, error) {
	var tenant models.TenantModel
	if err := s.db.First(&tenant, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &tenant, nil
}

// CreateTenant 创建租户
func (s *tenantService) CreateTenant(tenant *models.TenantModel) error {
	if err := s.db.Create(tenant).Error; err != nil {
		return err
	}
	return nil
}

// UpdateTenant 更新租户
func (s *tenantService) UpdateTenant(tenant *models.TenantModel) error {
	if err := s.db.Save(tenant).Error; err != nil {
		return err
	}
	return nil
}

// DeleteTenant 删除租户
func (s *tenantService) DeleteTenant(id uuid.UUID) error {
	if err := s.db.Delete(&models.TenantModel{}, "id = ?", id).Error; err != nil {
		return err
	}
	return nil
}
