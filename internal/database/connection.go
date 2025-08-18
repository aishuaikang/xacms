package database

import (
	"fmt"
	"log"
	"new-spbatc-drone-platform/internal/database/models"
	"os"
	"sync"

	_ "github.com/joho/godotenv/autoload"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	db   *gorm.DB
	once sync.Once
)

// GetDB 获取数据库连接实例（单例模式）
func GetDB() *gorm.DB {
	once.Do(func() {
		var err error
		dbname := os.Getenv("BLUEPRINT_DB_DATABASE")
		password := os.Getenv("BLUEPRINT_DB_PASSWORD")
		username := os.Getenv("BLUEPRINT_DB_USERNAME")
		port := os.Getenv("BLUEPRINT_DB_PORT")
		host := os.Getenv("BLUEPRINT_DB_HOST")

		db, err = gorm.Open(mysql.Open(fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", username, password, host, port, dbname)), &gorm.Config{
			// 打印日志
			Logger: logger.Default.LogMode(logger.Info),
		})
		if err != nil {
			log.Fatal("Failed to connect to database:", err)
		}

		// 自动迁移数据库
		err = db.AutoMigrate(
			&models.RoleModel{},
			&models.MenuModel{},
			&models.UserModel{},
			&models.TenantModel{},
			&models.DepartmentModel{},
		)
		if err != nil {
			log.Fatal("Failed to migrate database:", err)
		}
	})
	return db
}

// CloseDB 关闭数据库连接
func CloseDB() error {
	if db != nil {
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
