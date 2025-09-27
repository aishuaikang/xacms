package conn

import (
	"sync"
)

var FPVConnPool = Connection{
	Connections:      make(map[uint]Conn),
	ConnectionsMutex: sync.RWMutex{},
}
