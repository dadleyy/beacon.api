package routes

import "bytes"
import "encoding/json"

import "github.com/dadleyy/beacon.api/beacon/net"
import "github.com/dadleyy/beacon.api/beacon/defs"

func CreateDeviceMessage(runtime *net.RequestRuntime) net.HandlerResult {
	message := struct{}{}
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
