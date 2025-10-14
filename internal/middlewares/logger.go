package middlewares

import (
	"bytes"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"regexp"
	"runtime/debug"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config 日志中间件配置
type Config struct {
	// TimeFormat 时间格式，例如: time.RFC3339
	TimeFormat string
	// UTC 是否使用 UTC 时区
	UTC bool
	// SkipPaths 跳过日志记录的路径列表
	SkipPaths []string
	// SkipPathRegexps 跳过日志记录的路径正则表达式列表
	SkipPathRegexps []*regexp.Regexp
	// Context 自定义上下文字段函数
	Context func(*gin.Context) []zapcore.Field
	// DefaultLevel 默认日志级别
	DefaultLevel zapcore.Level
	// Skipper 自定义跳过函数
	Skipper func(*gin.Context) bool
	// EnableBody 是否记录请求体
	EnableBody bool
	// MaxBodySize 最大请求体大小（字节），超过此大小不记录请求体，默认 4KB
	MaxBodySize int64
}

// GinLogger 自定义 Gin 日志中间件，包含更详细的信息
// 使用默认配置
func GinLogger(logger *zap.Logger) gin.HandlerFunc {
	return GinLoggerWithConfig(logger, &Config{
		TimeFormat:   time.DateTime,
		UTC:          false,
		DefaultLevel: zapcore.InfoLevel,
	})
}

// GinLoggerWithConfig 使用自定义配置的 Gin 日志中间件
// 支持跳过特定路径、自定义时间格式、日志级别等
func GinLoggerWithConfig(logger *zap.Logger, conf *Config) gin.HandlerFunc {
	// 构建跳过路径的 map 以提高查找效率
	skipPaths := make(map[string]bool, len(conf.SkipPaths))
	for _, path := range conf.SkipPaths {
		skipPaths[path] = true
	}

	// 设置默认最大请求体大小
	maxBodySize := conf.MaxBodySize
	if maxBodySize == 0 {
		maxBodySize = 4096 // 默认 4KB
	}

	return func(c *gin.Context) {
		start := time.Now()
		// 保存原始路径和查询参数，防止某些中间件修改
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// 读取并保存请求体
		var bodyBytes []byte
		if conf.EnableBody && c.Request.Body != nil {
			// 读取请求体
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			// 恢复请求体，以便后续处理器可以继续读取
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// 处理请求
		c.Next()

		// 判断是否需要记录日志
		track := true

		// 检查是否在跳过路径列表中
		if _, ok := skipPaths[path]; ok || (conf.Skipper != nil && conf.Skipper(c)) {
			track = false
		}

		// 检查是否匹配跳过路径的正则表达式
		if track && len(conf.SkipPathRegexps) > 0 {
			for _, reg := range conf.SkipPathRegexps {
				if reg.MatchString(path) {
					track = false
					break
				}
			}
		}

		if track {
			end := time.Now()
			latency := end.Sub(start)
			if conf.UTC {
				end = end.UTC()
			}

			// 获取处理函数名称
			handlerName := c.HandlerName()

			// 获取 URL 路径参数
			params := c.Params
			var paramsMap map[string]string
			if len(params) > 0 {
				paramsMap = make(map[string]string, len(params))
				for _, param := range params {
					paramsMap[param.Key] = param.Value
				}
			}

			// 构建日志字段
			fields := []zapcore.Field{
				zap.Int("status", c.Writer.Status()),
				zap.String("method", c.Request.Method),
				zap.String("path", path),
				zap.String("query", query),
				zap.String("ip", c.ClientIP()),
				zap.String("user-agent", c.Request.UserAgent()),
				zap.Duration("latency", latency),
				zap.String("handler", handlerName),
			}

			// 添加 URL 路径参数
			if len(paramsMap) > 0 {
				fields = append(fields, zap.Any("params", paramsMap))
			}

			// 添加请求体（如果启用且未超过大小限制）
			if conf.EnableBody && len(bodyBytes) > 0 {
				if int64(len(bodyBytes)) <= maxBodySize {
					// 检查 Content-Type 是否为文本类型
					contentType := c.GetHeader("Content-Type")
					if strings.Contains(contentType, "application/json") {
						// JSON 格式化处理
						var jsonObj interface{}
						if err := sonic.Unmarshal(bodyBytes, &jsonObj); err == nil {
							// JSON 解析成功，使用结构化字段
							fields = append(fields, zap.Any("body", jsonObj))
						} else {
							// JSON 解析失败，使用原始字符串
							fields = append(fields, zap.String("body", string(bodyBytes)))
						}
					} else if strings.Contains(contentType, "application/x-www-form-urlencoded") ||
						strings.Contains(contentType, "text/") ||
						strings.Contains(contentType, "application/xml") {
						fields = append(fields, zap.String("body", string(bodyBytes)))
					} else {
						fields = append(fields, zap.String("body", "[binary data]"))
					}
				} else {
					fields = append(fields, zap.String("body", "[body too large]"))
				}
			}

			// 添加时间字段
			if conf.TimeFormat != "" {
				fields = append(fields, zap.String("time", end.Format(conf.TimeFormat)))
			}

			// 添加自定义上下文字段
			if conf.Context != nil {
				fields = append(fields, conf.Context(c)...)
			}

			// 如果有错误，记录所有错误
			if len(c.Errors) > 0 {
				for _, e := range c.Errors.Errors() {
					logger.Error(e, fields...)
				}
			} else {
				// 根据状态码和配置的默认级别决定日志级别
				if c.Writer.Status() >= 500 {
					logger.Error(path, fields...)
				} else if c.Writer.Status() >= 400 {
					logger.Warn(path, fields...)
				} else {
					logger.Log(conf.DefaultLevel, path, fields...)
				}
			}
		}
	}
}

// GinRecovery 自定义 Recovery 中间件
// 从任何 panic 中恢复并使用 zap 记录请求
// stack 参数表示是否输出堆栈信息
func GinRecovery(logger *zap.Logger, stack bool) gin.HandlerFunc {
	return CustomGinRecovery(logger, stack, defaultHandleRecovery)
}

// CustomGinRecovery 自定义 Recovery 中间件，支持自定义恢复处理函数
func CustomGinRecovery(logger *zap.Logger, stack bool, recovery gin.RecoveryFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 检查是否为断开的连接，这种情况不需要记录完整的 panic 堆栈
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						errStr := strings.ToLower(se.Error())
						if strings.Contains(errStr, "broken pipe") ||
							strings.Contains(errStr, "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				// 获取 HTTP 请求信息
				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				handlerName := c.HandlerName()

				if brokenPipe {
					// 如果是断开的连接，只记录基本错误信息
					logger.Error(c.Request.URL.Path,
						zap.Time("time", time.Now()),
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
						zap.String("handler", handlerName),
					)
					// 连接已断开，无法写入状态码
					c.Error(err.(error)) //nolint: errcheck
					c.Abort()
					return
				}

				// 记录 panic 信息
				if stack {
					logger.Error("[Recovery from panic]",
						zap.Time("time", time.Now()),
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
						zap.String("handler", handlerName),
						zap.String("stack", string(debug.Stack())),
					)
				} else {
					logger.Error("[Recovery from panic]",
						zap.Time("time", time.Now()),
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
						zap.String("handler", handlerName),
					)
				}

				// 调用恢复处理函数
				recovery(c, err)
			}
		}()
		c.Next()
	}
}

// defaultHandleRecovery 默认的 panic 恢复处理函数
func defaultHandleRecovery(c *gin.Context, err interface{}) {
	c.AbortWithStatus(http.StatusInternalServerError)
}
