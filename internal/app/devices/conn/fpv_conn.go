package conn

import (
	"sync"
)

type FpvConnection struct {
	Connection
}

func NewFPVConnection() *FpvConnection {
	return &FpvConnection{
		Connection: Connection{
			Connections:      make(map[uint]Conn),
			ConnectionsMutex: sync.RWMutex{},
		},
	}
}
