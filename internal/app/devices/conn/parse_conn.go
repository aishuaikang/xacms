package conn

import (
	"sync"
)

type ParseConnection struct {
	Connection
}

func NewParseConnection() *ParseConnection {
	return &ParseConnection{
		Connection: Connection{
			Connections:      make(map[uint]Conn),
			ConnectionsMutex: sync.RWMutex{},
		},
	}
}
