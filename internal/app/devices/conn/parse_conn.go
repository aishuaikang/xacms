package conn

import "sync"

var ParseConnPool = Connection{
	Connections:      make(map[uint]Conn),
	ConnectionsMutex: sync.RWMutex{},
}
