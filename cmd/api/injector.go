//go:build wireinject

package main

import (
	"xacms/internal/pkg/database"
	"xacms/internal/routes"
	"xacms/internal/server"
	"xacms/internal/services"
	"xacms/internal/utils"

	"github.com/google/wire"
)

func wireRouter(server *server.FiberServer, validator *utils.ValidationMiddleware) *routes.Router {
	wire.Build(
		database.NewDB,
		services.ServicesSet,
		routes.RoutesSet,
	)
	return nil
}
