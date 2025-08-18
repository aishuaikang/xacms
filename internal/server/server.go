package server

import (
	"github.com/gofiber/fiber/v2"
)

type FiberServer struct {
	*fiber.App

	// DB database.Service
}

func New() *FiberServer {
	server := &FiberServer{
		App: fiber.New(fiber.Config{
			ServerHeader: "new-spbatc-drone-platform",
			AppName:      "new-spbatc-drone-platform",
		}),
	}

	return server
}
