package conn

import "sync"

type FpvConnection struct {
	Connection
}

func NewFPVConnection() *FpvConnection {
	return &FpvConnection{
		Connection: Connection{
			Connections:      make([]Conn, 0),
			ConnectionsMutex: sync.RWMutex{},
		},
	}
}
