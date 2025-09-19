package global

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	Logger *zap.Logger
)

func NewZapLogger(level zapcore.Level, enabled bool) *zap.Logger {

	// 编码器配置
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.DateTime)
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	cores := []zapcore.Core{}

	// 控制台输出（可选，开发环境用）
	if enabled {
		consoleEncoderConfig := encoderConfig
		consoleEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		consoleCore := zapcore.NewCore(
			zapcore.NewConsoleEncoder(consoleEncoderConfig),
			zapcore.AddSync(os.Stdout),
			level,
		)
		cores = append(cores, consoleCore)
	} else {

		// 文件输出（生产环境用）
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
