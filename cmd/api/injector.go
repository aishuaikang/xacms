//go:build wireinject

package main

import (
	"context"
	"xacms/internal/pkg/config"
	"xacms/internal/pkg/database"
	"xacms/internal/routes"
	"xacms/internal/server"
	"xacms/internal/services"
	"xacms/internal/store"
	"xacms/internal/utils"

	"github.com/google/wire"
)

func wireRouter(ctx context.Context, cfg *config.Config, server *server.FiberServer, validator *utils.ValidationMiddleware) *routes.Router {
	wire.Build(
		database.NewDB,
		services.ServicesSet,
		routes.RoutesSet,
		store.StoreSet,
	)
	return nil
}
