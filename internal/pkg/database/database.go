package database

import (
	"log"
	"xacms/internal/models"
	"xacms/internal/pkg/config"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// getDB 获取数据库连接实例（单例模式）
func NewDB(config *config.Config) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(config.Database.Name), &gorm.Config{
		// 打印日志
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// 启用 WAL 模式
	// _ = db.Exec("PRAGMA journal_mode=WAL;")
	sqlDB, dbError := db.DB()
	if dbError != nil {
		log.Fatal("Failed to get database instance:", dbError)
	}
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetMaxOpenConns(10)

	// 自动迁移数据库
	if err = db.AutoMigrate(
		&models.RoleModel{},
		&models.MenuModel{},
		&models.UserModel{},
		&models.DeviceModel{},
		&models.DroneTargetModel{},
	); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
	log.Println("Database connected and migrated successfully")
	return db
}

// // CloseDB 关闭数据库连接
// func CloseDB() error {
// 	if db != nil {
// 		sqlDB, err := db.DB()
// 		if err != nil {
// 			return err
// 		}
// 		return sqlDB.Close()
// 	}
// 	return nil
// }
