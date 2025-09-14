package devices

import "github.com/google/wire"

var DevicesSet = wire.NewSet(NewFPVDevice, NewParseDevice, NewDevices)
