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

	command := interchange.ControlMessage{}

	switch color {
	case "green":
		command.Green = 255
	case "red":
		command.Red = 255
	default:
		command.Blue = 255
	}

	commandData, e := proto.Marshal(&command)

	if e != nil {
		return net.HandlerResult{Errors: []error{e}}
	}

	message := interchange.DeviceMessage{
		RequestPath:    "/device-state",
		Authentication: &interchange.DeviceMessageAuth{deviceId},
		RequestBody:    commandData,
	}

	runtime.Printf("attempting to update device %s to %s", deviceId, color)

	data, e := proto.Marshal(&message)

	if e != nil {
		return net.HandlerResult{Errors: []error{e}}
	}

	runtime.Publish(defs.DeviceControlChannelName, bytes.NewBuffer(data))

	return net.HandlerResult{}
}
