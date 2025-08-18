package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

type FiberServer struct {
	*fiber.App
}

func New() *FiberServer {

	app := fiber.New(fiber.Config{
		ServerHeader: "new-spbatc-drone-platform",
		AppName:      "new-spbatc-drone-platform",
	})

	// 设置压缩中间件
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestCompression, // 2
	}))

	// 设置日志中间件
	app.Use(logger.New(logger.Config{
		Format: "[${ip}]:${port} ${status} - ${method} ${path}\n",
	}))

	server := &FiberServer{
		App: app,
	}

	return server
}
