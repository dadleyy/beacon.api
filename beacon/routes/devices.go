package routes

import "bytes"
import "math/rand"
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
		Authentication: &interchange.DeviceMessageAuthentication{deviceID},
		Payload:        commandData,
	}

	runtime.Printf("attempting to update device %s to %s", deviceID, color)

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
