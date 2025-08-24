package routes

import "github.com/google/wire"

var RoutesSet = wire.NewSet(
	wire.Struct(new(RoleHandler), "*"),
	NewMenuHandler,
	wire.Struct(new(UserHandler), "*"),
	NewRouter,
)
