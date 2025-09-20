package cache

import "github.com/google/wire"

var CacheSet = wire.NewSet(
	NewDevicesCache,
	NewDecryptTokenCache,
	NewFPVWarningDataCache,
	NewParseCache,
	NewCommonCache,
)
