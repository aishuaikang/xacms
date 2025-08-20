package services

import (
	"gorm.io/gorm"
)

// ServiceManager 服务管理器，用于管理所有业务服务
type ServiceManager struct {
	UserService       UserService
	MenuService       MenuService
	RoleService       RoleService
	DepartmentService DepartmentService
	TenantService     TenantService
}

// NewServiceManager 创建服务管理器实例
func NewServiceManager(db *gorm.DB) *ServiceManager {
	return &ServiceManager{
		UserService:       NewUserService(db),
		MenuService:       NewMenuService(db),
		RoleService:       NewRoleService(db),
		DepartmentService: NewDepartmentService(db),
		TenantService:     NewTenantService(db),
	}
}
