package routes

import "github.com/google/wire"

var RoutesSet = wire.NewSet(
	wire.Struct(new(RoleHandler), "*"),
	NewMenuHandler,
	NewUserHandler,
	NewRouter,
)
