package app

import (
	"github.com/gin-gonic/gin"
)

// type FiberServer struct {
// 	*fiber.App
// }

type GinServer struct {
	*gin.Engine
}

func NewFiberServer() *GinServer {
	engine := gin.New()
	engine.Use(gin.Logger())   // 日志中间件
	engine.Use(gin.Recovery()) // Recovery 中间件，捕获 panic
	// engine.Use(middleware.Cors)
	engine.MaxMultipartMemory = 2 << 30

	// app := fiber.New(fiber.Config{
	// 	ServerHeader:  "uav_defender",
	// 	AppName:       "uav_defender",
	// 	CaseSensitive: true,
	// 	JSONEncoder:   sonic.Marshal,
	// 	JSONDecoder:   sonic.Unmarshal,
	// })

	// // 设置压缩中间件
	// app.Use(compress.New(compress.Config{
	// 	Level: compress.LevelBestCompression, // 2
	// }))

	// // 设置日志中间件
	// app.Use(logger.New(logger.Config{
	// 	TimeFormat: time.DateTime,
	// }))

	// server := &FiberServer{
	// 	App: app,
	// }
	server := &GinServer{
		Engine: engine,
	}

	return server
}
