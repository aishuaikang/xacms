package global

import (
	"os"
	"time"
	"uav_defender/internal/pkg/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	Logger *zap.Logger
)

// NewZapLogger 创建一个新的 Zap Logger
// level: 日志级别 (DebugLevel, InfoLevel, WarnLevel, ErrorLevel)
// environment: 环境模式 (development=开发模式, production=生产模式)
// disableConsole: 是否禁用控制台输出
//
// 生产模式特点:
// 1. 同时输出到 stdout (供 systemd/journald 捕获) 和文件
// 2. 使用 JSON 格式便于日志收集和分析
// 3. 按日志级别分文件存储，便于排查问题
func NewZapLogger(level zapcore.Level, environment config.Environment, disableConsole bool) *zap.Logger {

	// 编码器配置
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.DateTime)
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	cores := []zapcore.Core{}

	// 控制台输出（开发环境：彩色输出；生产环境：JSON格式输出到stdout供journald捕获）
	// 如果 disableConsole=true，则完全跳过控制台输出
	if !disableConsole {
		if environment.IsDevelopment() {
			// 开发环境：彩色控制台输出
			consoleEncoderConfig := encoderConfig
			consoleEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
			consoleCore := zapcore.NewCore(
				zapcore.NewConsoleEncoder(consoleEncoderConfig),
				zapcore.AddSync(os.Stdout),
				level,
			)
			cores = append(cores, consoleCore)
		} else {
			// 生产环境：JSON格式输出到stdout，供systemd/journald捕获
			stdoutCore := zapcore.NewCore(
				zapcore.NewJSONEncoder(encoderConfig),
				zapcore.AddSync(os.Stdout),
				level,
			)
			cores = append(cores, stdoutCore)
		}
	}

	// 文件输出（生产环境用，同时输出到文件和控制台）
	if !environment.IsDevelopment() {
		maxSize := 100
		// 日志文件切割
		maxBackups := 10
		// 日志文件保存时间
		maxAge := 7

		// 不同等级日志文件分割
		infoWriter := zapcore.AddSync(&lumberjack.Logger{
			Filename:   "logs/info.log",
			MaxSize:    maxSize,
			MaxBackups: maxBackups,
			MaxAge:     maxAge,
			Compress:   true,
		})
		warnWriter := zapcore.AddSync(&lumberjack.Logger{
			Filename:   "logs/warn.log",
			MaxSize:    maxSize,
			MaxBackups: maxBackups,
			MaxAge:     maxAge,
			Compress:   true,
		})
		errorWriter := zapcore.AddSync(&lumberjack.Logger{
			Filename:   "logs/error.log",
			MaxSize:    maxSize,
			MaxBackups: maxBackups,
			MaxAge:     maxAge,
			Compress:   true,
		})

		// info 日志文件
		infoLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl == zapcore.InfoLevel
		})
		infoCore := zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			infoWriter,
			infoLevel,
		)
		cores = append(cores, infoCore)

		// warn 日志文件
		warnLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl == zapcore.WarnLevel
		})
		warnCore := zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			warnWriter,
			warnLevel,
		)
		cores = append(cores, warnCore)

		// error 及以上日志文件
		errorLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= zapcore.ErrorLevel
		})
		errorCore := zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			errorWriter,
			errorLevel,
		)
		cores = append(cores, errorCore)
	}

	// 多路输出
	core := zapcore.NewTee(cores...)

	// 构建 logger
	Logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	return Logger
}
