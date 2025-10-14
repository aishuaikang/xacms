package database

import (
	"time"
	"uav_defender/internal/models"
	"uav_defender/internal/pkg/config"
	"uav_defender/internal/pkg/global"

	"github.com/glebarez/sqlite"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// getDB 获取数据库连接实例（单例模式）
func NewDB(config *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(config.Database.Name), &gorm.Config{
		// 打印日志
		Logger: NewGormLogger(global.Logger, config.Log.DatabaseLevel, config.Environment),
	})
	if err != nil {
		return nil, err
	}

	// 启用 WAL 模式，提高并发性能
	_ = db.Exec("PRAGMA journal_mode=WAL;")

	// 设置忙碌超时（毫秒），避免 disk I/O error
	_ = db.Exec("PRAGMA busy_timeout=5000;")

	// 设置同步模式为 FULL，最大化数据安全性（生产环境推荐）
	// FULL: 最安全，确保断电时数据完整性
	// NORMAL: 较快，但极端情况下可能丢失最后的事务
	_ = db.Exec("PRAGMA synchronous=FULL;")

	// 增加缓存大小（页数，默认-2000，约2MB）
	_ = db.Exec("PRAGMA cache_size=-8000;") // 8MB

	// 设置临时文件存储在内存中
	_ = db.Exec("PRAGMA temp_store=MEMORY;")

	sqlDB, dbError := db.DB()
	if dbError != nil {
		return nil, dbError
	}

	// 优化连接池设置
	sqlDB.SetMaxIdleConns(2)            // 增加空闲连接数
	sqlDB.SetMaxOpenConns(1)            // SQLite 只支持单个写入，保持1
	sqlDB.SetConnMaxLifetime(time.Hour) // 连接最大生命周期

	// 自动迁移数据库
	if err = db.AutoMigrate(
		models.Role{},
		models.Menu{},
		models.User{},
		models.ModuleGroupDevice{},
		models.Whitelist{},
		models.DroneTarget{},
		models.FpvVideo{},
		models.StrikeReport{},
		// models.Name{},
	); err != nil {
		return nil, err
	}
	return db, nil
}

type zapWriter struct {
	sugar *zap.SugaredLogger
}

func (z *zapWriter) Printf(format string, args ...interface{}) {
	z.sugar.Infof(format, args...)
}

func NewGormLogger(zapLogger *zap.Logger, logLevel logger.LogLevel, environment config.Environment) logger.Interface {
	return logger.New(
		&zapWriter{sugar: zapLogger.Sugar()}, // 实现 logger.Writer 接口
		logger.Config{
			SlowThreshold: time.Second,                 // 慢 SQL 阈值
			LogLevel:      logLevel,                    // 日志级别
			Colorful:      environment.IsDevelopment(), // 当环境为开发环境时，启用彩色日志
		},
	)
}
