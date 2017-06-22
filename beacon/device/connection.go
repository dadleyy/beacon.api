package device

import "github.com/satori/go.uuid"

import "github.com/dadleyy/beacon.api/beacon/defs"

// Connection defines the interface of a device connection - a wrapper around a CommandStreamer with a unique ID
type Connection struct {
	defs.CommandStreamer
	uuid.UUID
}

// GetID returns the string version of the unique identifier for the connection
func (connection *Connection) GetID() string {
	return connection.UUID.String()
}
