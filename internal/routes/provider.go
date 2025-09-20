package routes

import "github.com/google/wire"

var RoutesSet = wire.NewSet(
	wire.Struct(new(RoleRouter), "*"),
	wire.Struct(new(MenuRouter), "*"),
	wire.Struct(new(UserRouter), "*"),
	wire.Struct(new(DeviceRouter), "*"),
	wire.Struct(new(SSERouter), "*"),
	wire.Struct(new(DroneTargetRouter), "*"),
	wire.Struct(new(WhitelistRouter), "*"),
	wire.Struct(new(FPVRouter), "*"),
	publicRoutesSet,
	NewRouter,
)

var publicRoutesSet = wire.NewSet(
	wire.Struct(new(UserPublicRouter), "*"),
)
