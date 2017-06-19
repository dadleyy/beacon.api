package device

import "github.com/satori/go.uuid"
import "github.com/gorilla/websocket"

type Connection struct {
	*websocket.Conn
	uuid.UUID
}

func (connection *Connection) GetID() string {
	return connection.UUID.String()
}
