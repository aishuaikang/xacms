//go:build wireinject

package main

import (
	"context"
	"uav_defender/internal/app"
	"uav_defender/internal/app/devices"
	"uav_defender/internal/cache"
	"uav_defender/internal/pkg/utils"
	"uav_defender/internal/routes"

	"uav_defender/internal/services"
	"uav_defender/internal/tasks"

	"github.com/google/wire"
	"gorm.io/gorm"
)

func wireServer(ctx context.Context, db *gorm.DB, validator *utils.ValidationMiddleware) *app.GinServer {
	wire.Build(
		app.NewGinServer,
		services.ServicesSet,
		routes.RoutesSet,
		devices.DevicesSet,
		cache.CacheSet,
		tasks.TaskSet,
	)
	return nil
}
