package server

import (
	"github.com/gofiber/fiber/v2"

	"new-spbatc-drone-platform/internal/database"
)

type FiberServer struct {
	*fiber.App

	db database.Service
}

func New() *FiberServer {
	server := &FiberServer{
		App: fiber.New(fiber.Config{
			ServerHeader: "new-spbatc-drone-platform",
			AppName:      "new-spbatc-drone-platform",
		}),

		db: database.New(),
	}

	return server
}
