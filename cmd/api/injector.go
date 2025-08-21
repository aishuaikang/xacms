//go:build wireinject

package main

import (
	"new-spbatc-drone-platform/internal/database"
	"new-spbatc-drone-platform/internal/routes"
	"new-spbatc-drone-platform/internal/server"
	"new-spbatc-drone-platform/internal/services"
	"new-spbatc-drone-platform/internal/utils"

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
