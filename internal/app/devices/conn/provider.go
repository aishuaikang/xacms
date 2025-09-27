package conn

import "github.com/google/wire"

var ConnSet = wire.NewSet(NewFPVConnection, NewParseConnection, NewDetectorConnection)
