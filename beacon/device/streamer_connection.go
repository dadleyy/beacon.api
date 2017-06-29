package device

import "io"
import "github.com/satori/go.uuid"
import "github.com/golang/protobuf/proto"

import "github.com/dadleyy/beacon.api/beacon/defs"
import "github.com/dadleyy/beacon.api/beacon/interchange"

// NewStreamerConnection returns a device connection who's underlying IO is managed through a streamer interface
func NewStreamerConnection(stream defs.Streamer, id uuid.UUID, sign defs.Signer) *StreamerConnection {
	return &StreamerConnection{stream, sign, id}
}

// StreamerConnection is an implementation of the device.Connection interface using a websocket
type StreamerConnection struct {
	defs.Streamer
	defs.Signer
	id uuid.UUID
}

// Send writes the provided byte data to the next available writer from the underlying streamer interface
func (connection *StreamerConnection) Send(message interchange.DeviceMessage) error {
	d, e := proto.Marshal(&message)

	if e != nil {
		return e
	}

	w, e := connection.NextWriter(defs.TextWriter)

	if e != nil {
		return e
	}

	defer w.Close()

	return connection.Sign(w, d)
}

// Receive returns the next available reader from the underlying streamer interface
func (connection *StreamerConnection) Receive() (io.Reader, error) {
	_, r, e := connection.NextReader()
	return r, e
}

// GetID returns the unique identifier created for this connection as a string
func (connection *StreamerConnection) GetID() string {
	return connection.id.String()
}
