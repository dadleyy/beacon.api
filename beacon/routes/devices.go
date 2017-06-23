package routes

import "bytes"
import "math/rand"
import "github.com/golang/protobuf/proto"

import "github.com/dadleyy/beacon.api/beacon/net"
import "github.com/dadleyy/beacon.api/beacon/defs"
import "github.com/dadleyy/beacon.api/beacon/device"
import "github.com/dadleyy/beacon.api/beacon/interchange"

// NewDevicesAPI constructs the devices api
func NewDevicesAPI(registry device.Registry) *Devices {
	return &Devices{registry}
}

// Devices route engine is responsible for CRUD operations on the device objects themselves.
type Devices struct {
	registry device.Registry
}

// ListDevices will return a list of the UUIDs registered in the registry
func (devices *Devices) ListDevices(runtime *net.RequestRuntime) net.HandlerResult {
	ids, e := devices.registry.List()

	if e != nil {
		runtime.Printf("unable to lookup device id list: %s", e.Error())
		return runtime.ServerError()
	}

	return net.HandlerResult{Results: ids}
}

// UpdateShorthand accepts a device id and a color (via url params from the req) and updates the device to that color.
func (devices *Devices) UpdateShorthand(runtime *net.RequestRuntime) net.HandlerResult {
	query, color := runtime.Get("uuid"), runtime.Get("color")
	details, e := devices.registry.Find(query)

	if e != nil {
		runtime.Printf("shorthand update w/ invalid device id: %s (%s)", query, e.Error())
		return runtime.LogicError("not-found")
	}

	frame := interchange.ControlFrame{}

	switch color {
	case "green":
		frame.Green = 255
	case "red":
		frame.Red = 255
	case "blue":
		frame.Blue = 255
	case "rand":
		frame = interchange.ControlFrame{devices.randColorValue(), devices.randColorValue(), devices.randColorValue()}
	default:
	}

	commandData, e := proto.Marshal(&interchange.ControlMessage{
		Frames: []*interchange.ControlFrame{&frame},
	})

	if e != nil {
		return net.HandlerResult{Errors: []error{e}}
	}

	message := interchange.DeviceMessage{
		Type:           interchange.DeviceMessageType_CONTROL,
		Authentication: &interchange.DeviceMessageAuthentication{details.DeviceID},
		Payload:        commandData,
	}

	runtime.Printf("attempting to update device %s to %s", details.DeviceID, color)

	data, e := proto.Marshal(&message)

	if e != nil {
		return net.HandlerResult{Errors: []error{e}}
	}

	runtime.Publish(defs.DeviceControlChannelName, bytes.NewBuffer(data))

	return net.HandlerResult{}
}

func (devices *Devices) randColorValue() uint32 {
	return uint32(rand.Intn(255))
}
