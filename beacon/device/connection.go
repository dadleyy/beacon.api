package device

import "io"
import "github.com/dadleyy/beacon.api/beacon/interchange"

// Connection defines an interface that describes the capabilities of a device connected to the api - send + receive
type Connection interface {
	Send(interchange.DeviceMessage) error
	Receive() (io.Reader, error)
	GetID() string
	Close() error
}
