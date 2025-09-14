package conn

import "sync"

type ParseConnection struct {
	Connection
}

func NewParseConnection() *ParseConnection {
	return &ParseConnection{
		Connection: Connection{
			Connections:      make([]Conn, 0),
			ConnectionsMutex: sync.RWMutex{},
		},
	}
}
