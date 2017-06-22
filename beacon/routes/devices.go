package routes

import "bytes"
import "github.com/golang/protobuf/proto"

import "github.com/dadleyy/beacon.api/beacon/net"
import "github.com/dadleyy/beacon.api/beacon/defs"
import "github.com/dadleyy/beacon.api/beacon/device"
import "github.com/dadleyy/beacon.api/beacon/interchange"

// Devices route engine is responsible for CRUD operations on the device objects themselves.
type Devices struct {
	device.Registry
}

// UpdateShorthand accepts a device id and a color (via url params from the req) and updates the device to that color.
func (devices *Devices) UpdateShorthand(runtime *net.RequestRuntime) net.HandlerResult {
	deviceID, color := runtime.Get("uuid"), runtime.Get("color")

	if exists := devices.Exists(deviceID); exists != true {
		return runtime.LogicError("not-found")
	}

	frame := interchange.ControlFrame{}

	switch color {
	case "green":
		frame.Green = 255
	case "red":
		frame.Red = 255
	default:
		frame.Blue = 255
	}

	command := interchange.ControlMessage{
		Frames: []*interchange.ControlFrame{&frame},
	}

	message := interchange.DeviceMessage{
		Authentication: &interchange.DeviceMessageAuthentication{deviceID},
		RequestBody:    &interchange.DeviceMessage_Control{&command},
	}

	runtime.Printf("attempting to update device %s to %s", deviceID, color)

	data, e := proto.Marshal(&message)

	if e != nil {
		return net.HandlerResult{Errors: []error{e}}
	}

	runtime.Publish(defs.DeviceControlChannelName, bytes.NewBuffer(data))

	return net.HandlerResult{}
}
