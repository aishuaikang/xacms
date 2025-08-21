package routes

import "github.com/google/wire"

var RoutesSet = wire.NewSet(
	NewTenantHandler,
	NewRoleHandler,
	NewDepartmentHandler,
	NewMenuHandler,
	NewUserHandler,
	NewRouter,
)
