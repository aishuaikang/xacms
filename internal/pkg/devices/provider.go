package devices

import "github.com/google/wire"

var DevicesSet = wire.NewSet(
	NewParseDevice,
	NewFpvDevice,
	NewStrikeDevice,
	NewDeviceManager,
)
