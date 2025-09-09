package database

import (
	"log"
	"os"
	"sync"
	"xacms/internal/models"

	_ "github.com/joho/godotenv/autoload"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	db   *gorm.DB
	once sync.Once
)

// getDB 获取数据库连接实例（单例模式）
func NewDB() *gorm.DB {
	once.Do(func() {
		var err error
		// dbname := os.Getenv("BLUEPRINT_DB_DATABASE")
		// password := os.Getenv("BLUEPRINT_DB_PASSWORD")
		// username := os.Getenv("BLUEPRINT_DB_USERNAME")
		// port := os.Getenv("BLUEPRINT_DB_PORT")
		dbPath := os.Getenv("DB")

		db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
			// 打印日志
			Logger: logger.Default.LogMode(logger.Info),
		})
		if err != nil {
			log.Fatal("Failed to connect to database:", err)
		}

		// 启用 WAL 模式
		_ = db.Exec("PRAGMA journal_mode=WAL;")
		sqlDB, dbError := db.DB()
		if dbError != nil {
			log.Fatal("Failed to get database instance:", dbError)
		}
		sqlDB.SetMaxIdleConns(1)
		sqlDB.SetMaxOpenConns(10)

		// 自动迁移数据库
		err = db.AutoMigrate(
			&models.RoleModel{},
			&models.MenuModel{},
			&models.UserModel{},
			&models.DeviceModel{},
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
