package device

import "github.com/satori/go.uuid"

import "github.com/dadleyy/beacon.api/beacon/defs"

type Connection struct {
	defs.CommandStreamer
	uuid.UUID
}

func (connection *Connection) GetID() string {
	return connection.UUID.String()
}
