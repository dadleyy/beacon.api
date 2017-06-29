package routes

import "bytes"
import "github.com/golang/protobuf/proto"

import "github.com/dadleyy/beacon.api/beacon/net"
import "github.com/dadleyy/beacon.api/beacon/defs"
import "github.com/dadleyy/beacon.api/beacon/device"
import "github.com/dadleyy/beacon.api/beacon/interchange"

// DeviceMessages is the route group that handles creating device messages
type DeviceMessages struct {
	device.Index
}

// CreateMessage publishes a new DeviceMessage to the control stream
func (messages *DeviceMessages) CreateMessage(runtime *net.RequestRuntime) net.HandlerResult {
	message := struct {
		DeviceID string `json:"device_id"`
		Red      uint32 `json:"red"`
		Green    uint32 `json:"green"`
		Blue     uint32 `json:"blue"`
	}{}

	if e := runtime.ReadBody(&message); e != nil {
		return net.HandlerResult{Errors: []error{e}}
	}

	runtime.Printf("creating device message for[%s]: %v", message.DeviceID, message)

	if _, e := messages.Find(message.DeviceID); e != nil {
		runtime.Printf("unable to locate device: %s", message.DeviceID)
		return runtime.LogicError("not-found")
	}

	commandData, e := proto.Marshal(&interchange.ControlMessage{
		Frames: []*interchange.ControlFrame{
			&interchange.ControlFrame{message.Red, message.Green, message.Blue},
		},
	})

	if e != nil {
		return net.HandlerResult{Errors: []error{e}}
	}

	deviceMessage := interchange.DeviceMessage{
		Type: interchange.DeviceMessageType_CONTROL,
		Authentication: &interchange.DeviceMessageAuthentication{
			DeviceID: message.DeviceID,
		},
		Payload: commandData,
	}

	data, e := proto.Marshal(&deviceMessage)

	if e != nil {
		return net.HandlerResult{Errors: []error{e}}
	}

	runtime.Publish(defs.DeviceControlChannelName, bytes.NewBuffer(data))
	return net.HandlerResult{}
}
