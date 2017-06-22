package routes

import "github.com/satori/go.uuid"
import "github.com/dadleyy/beacon.api/beacon/net"
import "github.com/dadleyy/beacon.api/beacon/device"

// Registration route engine handles receiving http reqests, upgrading and sending along to the registation stream
type Registration struct {
	Registrations chan<- *device.Connection
}

// Register is the route handler responsible for upgrating + registering connections
func (registration *Registration) Register(runtime *net.RequestRuntime) net.HandlerResult {
	connection, e := runtime.Websocket()

	if e != nil {
		runtime.Printf("[warn] unable to upgrade websocket: %s", e.Error())
		return net.HandlerResult{Errors: []error{e}}
	}

	deviceConnection := device.Connection{
		CommandStreamer: connection,
		UUID:            uuid.NewV4(),
	}

	registration.Registrations <- &deviceConnection
	return net.HandlerResult{NoRender: true}
}
