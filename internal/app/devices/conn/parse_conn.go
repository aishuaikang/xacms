package conn

import (
	"sync"

	"github.com/google/uuid"
)

type ParseConnection struct {
	Connection
}

func NewParseConnection() *ParseConnection {
	return &ParseConnection{
		Connection: Connection{
			Connections:      make(map[uuid.UUID]Conn),
			ConnectionsMutex: sync.RWMutex{},
		},
	}
}
