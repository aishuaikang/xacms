//go:build wireinject

package main

import (
	"context"
	"xacms/internal/app"
	"xacms/internal/app/devices"
	"xacms/internal/app/devices/conn"
	"xacms/internal/cache"
	"xacms/internal/pkg/config"
	"xacms/internal/pkg/database"
	"xacms/internal/pkg/utils"
	"xacms/internal/routes"

	"xacms/internal/services"
	"xacms/internal/tasks"

	"github.com/google/wire"
)

func wireRouter(ctx context.Context, cfg *config.Config, server *app.FiberServer, validator *utils.ValidationMiddleware) *routes.Router {
	wire.Build(
		database.NewDB,
		services.ServicesSet,
		routes.RoutesSet,
		devices.DevicesSet,
		cache.CacheSet,
		tasks.TaskSet,
		conn.ConnSet,
	)
	return nil
}
