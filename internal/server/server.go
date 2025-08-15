package server

import (
	"github.com/gofiber/fiber/v2"

	"my_project/internal/database"
)

type FiberServer struct {
	*fiber.App

	db database.Service
}

func New() *FiberServer {
	server := &FiberServer{
		App: fiber.New(fiber.Config{
			ServerHeader: "my_project",
			AppName:      "my_project",
		}),

		db: database.New(),
	}

	return server
}
