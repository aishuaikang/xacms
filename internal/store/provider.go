package store

import "github.com/google/wire"

var StoreSet = wire.NewSet(NewDeviceStore, NewFPVStore)
