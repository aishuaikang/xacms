package services

import "github.com/google/wire"

var ServicesSet = wire.NewSet(
	NewUserService,
	NewTenantService,
	NewRoleService,
	NewDepartmentService,
	NewMenuService,
)
