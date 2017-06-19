package routes

import "bytes"
import "encoding/json"

import "github.com/dadleyy/beacon.api/beacon/net"
import "github.com/dadleyy/beacon.api/beacon/defs"
import "github.com/dadleyy/beacon.api/beacon/device"

func UpdateDeviceShorthand(runtime *net.RequestRuntime) net.HandlerResult {
	deviceId, color := runtime.Get("uuid"), runtime.Get("color")

	message := device.ControlMessage{DeviceId: deviceId}

	switch color {
	case "green":
		message.Green = 255
	case "red":
		message.Red = 255
	default:
		message.Blue = 255
	}

	data, e := json.Marshal(&message)

	if e != nil {
		return net.HandlerResult{Errors: []error{e}}
	}

	runtime.Publish(defs.DeviceControlChannelName, bytes.NewBuffer(data))

	return net.HandlerResult{}
}

func CreateDeviceMessage(runtime *net.RequestRuntime) net.HandlerResult {
	message := device.ControlMessage{}
	if e := runtime.ReadBody(&message); e != nil {
		return net.HandlerResult{Errors: []error{e}}
	}

	data, e := json.Marshal(&message)

	if e != nil {
		return net.HandlerResult{Errors: []error{e}}
	}

	runtime.Publish(defs.DeviceControlChannelName, bytes.NewBuffer(data))
	return net.HandlerResult{}
}
