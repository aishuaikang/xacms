package database

import "gorm.io/gorm"

var dbInstance *gorm.DB

// NewDB 创建数据库服务实例
func NewDB() *gorm.DB {
	// 重用连接
	if dbInstance != nil {
		return dbInstance
	}

	// 获取数据库连接
	dbInstance = GetDB()

	// 创建服务管理器
	// serviceManager := services.NewServiceManager(db)

	return dbInstance
}
