package routes

import "github.com/google/wire"

var RoutesSet = wire.NewSet(
	wire.Struct(new(TenantHandler), "*"),
	wire.Struct(new(RoleHandler), "*"),
	wire.Struct(new(DepartmentHandler), "*"),
	wire.Struct(new(MenuHandler), "*"),
	wire.Struct(new(UserHandler), "*"),
	NewRouter,
)
