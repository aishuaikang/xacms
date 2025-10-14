package routes

import "github.com/google/wire"

var RoutesSet = wire.NewSet(
	wire.Struct(new(RoleRouter), "*"),
	wire.Struct(new(MenuRouter), "*"),
	wire.Struct(new(UserRouter), "*"),
	publicRoutesSet,
	NewRouter,
)

var publicRoutesSet = wire.NewSet(
	wire.Struct(new(UserPublicRouter), "*"),
)
