package services

import (
	"new-spbatc-drone-platform/internal/database/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DepartmentService 部门服务接口
type DepartmentService interface {
	GetDepartments() ([]models.DepartmentModel, error)
	GetDepartmentByID(id uuid.UUID) (*models.DepartmentModel, error)
	CreateDepartment(department *models.DepartmentModel) error
	UpdateDepartment(department *models.DepartmentModel) error
	DeleteDepartment(id uuid.UUID) error
}

// departmentService 部门服务实现
type departmentService struct {
	db *gorm.DB
}

// NewDepartmentService 创建部门服务实例
func NewDepartmentService(db *gorm.DB) DepartmentService {
	return &departmentService{
		db: db,
	}
}

// GetDepartments 获取部门列表
func (s *departmentService) GetDepartments() ([]models.DepartmentModel, error) {
	var departments []models.DepartmentModel
	if err := s.db.Find(&departments).Error; err != nil {
		return nil, err
	}
	return departments, nil
}

// GetDepartmentByID 根据ID获取部门
func (s *departmentService) GetDepartmentByID(id uuid.UUID) (*models.DepartmentModel, error) {
	var department models.DepartmentModel
	if err := s.db.First(&department, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &department, nil
}

// CreateDepartment 创建部门
func (s *departmentService) CreateDepartment(department *models.DepartmentModel) error {
	if err := s.db.Create(department).Error; err != nil {
		return err
	}
	return nil
}

// UpdateDepartment 更新部门
func (s *departmentService) UpdateDepartment(department *models.DepartmentModel) error {
	if err := s.db.Save(department).Error; err != nil {
		return err
	}
	return nil
}

// DeleteDepartment 删除部门
func (s *departmentService) DeleteDepartment(id uuid.UUID) error {
	if err := s.db.Delete(&models.DepartmentModel{}, "id = ?", id).Error; err != nil {
		return err
	}
	return nil
}
