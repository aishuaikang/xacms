package services

import "github.com/google/wire"

var ServicesSet = wire.NewSet(
	NewUserService,
	NewRoleService,
	NewMenuService,
	NewCommonService,
	NewDeviceService,
)
