package database

import (
	"new-spbatc-drone-platform/internal/database/services"
)

// service 数据库服务实现（适配器模式，包装新的模块化服务）
type Service struct {
	ServiceManager *services.ServiceManager
}

var dbInstance *Service

// New 创建数据库服务实例
func New() *Service {
	// 重用连接
	if dbInstance != nil {
		return dbInstance
	}

	// 获取数据库连接
	db := GetDB()

	// 创建服务管理器
	serviceManager := services.NewServiceManager(db)

	dbInstance = &Service{
		ServiceManager: serviceManager,
	}

	return dbInstance
}
