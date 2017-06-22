package routes

import "bytes"
import "github.com/golang/protobuf/proto"

import "github.com/dadleyy/beacon.api/beacon/net"
import "github.com/dadleyy/beacon.api/beacon/defs"
import "github.com/dadleyy/beacon.api/beacon/device"
import "github.com/dadleyy/beacon.api/beacon/interchange"

type Devices struct {
	device.Registry
}

func (devices *Devices) UpdateShorthand(runtime *net.RequestRuntime) net.HandlerResult {
	deviceId, color := runtime.Get("uuid"), runtime.Get("color")

	if exists := devices.Exists(deviceId); exists != true {
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
		Authentication: &interchange.DeviceMessageAuthentication{deviceId},
		RequestBody:    &interchange.DeviceMessage_Control{&command},
	}

	runtime.Printf("attempting to update device %s to %s", deviceId, color)

	data, e := proto.Marshal(&message)

	if e != nil {
		return net.HandlerResult{Errors: []error{e}}
	}

	runtime.Publish(defs.DeviceControlChannelName, bytes.NewBuffer(data))

	return net.HandlerResult{}
}
