package conn

import (
	"sync"

	"github.com/google/uuid"
)

type FpvConnection struct {
	Connection
}

func NewFPVConnection() *FpvConnection {
	return &FpvConnection{
		Connection: Connection{
			Connections:      make(map[uuid.UUID]Conn),
			ConnectionsMutex: sync.RWMutex{},
		},
	}
}
