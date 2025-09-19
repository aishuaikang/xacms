package database

import (
	"log"
	"time"
	"uav_defender/internal/models"
	"uav_defender/internal/pkg/config"
	"uav_defender/internal/pkg/global"

	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// getDB 获取数据库连接实例（单例模式）
func NewDB(config *config.Config) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(config.Database.Name), &gorm.Config{
		// 打印日志
		Logger: NewGormLogger(global.Logger, config.Log.DatabaseLevel, config.Log.Enabled),
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

type zapWriter struct {
	sugar *zap.SugaredLogger
}

func (z *zapWriter) Printf(format string, args ...interface{}) {
	z.sugar.Infof(format, args...)
}

func NewGormLogger(zapLogger *zap.Logger, logLevel logger.LogLevel, enabled bool) logger.Interface {
	return logger.New(
		&zapWriter{sugar: zapLogger.Sugar()}, // 实现 logger.Writer 接口
		logger.Config{
			SlowThreshold: time.Second, // 慢 SQL 阈值
			LogLevel:      logLevel,    // 日志级别
			Colorful:      enabled,     // zap 文件建议关闭彩色
		},
	)
}
