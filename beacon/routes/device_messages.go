package routes

import "bytes"
import "github.com/golang/protobuf/proto"

import "github.com/dadleyy/beacon.api/beacon/net"
import "github.com/dadleyy/beacon.api/beacon/defs"
import "github.com/dadleyy/beacon.api/beacon/device"
import "github.com/dadleyy/beacon.api/beacon/logging"
import "github.com/dadleyy/beacon.api/beacon/interchange"

// NewDeviceMessagesAPI returns a new api for creating device messages.
func NewDeviceMessagesAPI(index device.Index, auth device.TokenStore) *DeviceMessages {
	logger := logging.New(defs.DeviceMessagesAPILogPrefix, logging.Green)
	return &DeviceMessages{logger, auth, index}
}

// DeviceMessages is the route group that handles creating device messages
type DeviceMessages struct {
	logging.LeveledLogger
	device.TokenStore
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

	details, e := messages.FindDevice(message.DeviceID)

	if e != nil {
		messages.Warnf("unable to locate device: %s", message.DeviceID)
		return runtime.LogicError("not-found")
	}

	token := runtime.HeaderValue(defs.APIUserTokenHeader)

	if token == "" || messages.AuthorizeToken(details.DeviceID, token, controllerPermission) != true {
		messages.Warnf("unauthorized attempt to control device (token: %s, device: %s)", token, details.DeviceID)
		return runtime.LogicError("invalid-token")
	}

	messages.Debugf("creating device message for[%s]: %v", message.DeviceID, message)

	commandData, e := proto.Marshal(&interchange.ControlMessage{
		Frames: []*interchange.ControlFrame{
			&interchange.ControlFrame{
				Red:   message.Red,
				Green: message.Green,
				Blue:  message.Blue,
			},
		},
	})

	if e != nil {
		return net.HandlerResult{Errors: []error{e}}
	}

	deviceMessage := interchange.DeviceMessage{
		Type: interchange.DeviceMessageType_CONTROL,
		Authentication: &interchange.DeviceMessageAuthentication{
			DeviceID: details.DeviceID,
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
