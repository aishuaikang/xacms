package store

import "github.com/google/wire"

var StoreSet = wire.NewSet(NewCommonStore, NewDeviceStore, NewFPVStore, NewParseStore)
