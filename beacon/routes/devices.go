package routes

import "bytes"
import "regexp"
import "math/rand"
import "encoding/hex"
import "github.com/golang/protobuf/proto"

import "github.com/dadleyy/beacon.api/beacon/net"
import "github.com/dadleyy/beacon.api/beacon/defs"
import "github.com/dadleyy/beacon.api/beacon/device"
import "github.com/dadleyy/beacon.api/beacon/logging"
import "github.com/dadleyy/beacon.api/beacon/interchange"

var (
	hexColorRegex = regexp.MustCompilePOSIX("[0-9a-f]{6}")
)

const (
	controllerPermission = defs.SecurityDeviceTokenPermissionController
)

// NewDevicesAPI constructs the devices api
func NewDevicesAPI(registry device.Registry, auth device.TokenStore) *Devices {
	logger := logging.New(defs.DevicesAPILogPrefix, logging.Green)
	return &Devices{logger, registry, auth}
}

// Devices route engine is responsible for CRUD operations on the device objects themselves.
type Devices struct {
	logging.LeveledLogger
	device.Registry
	device.TokenStore
}

// ListDevices will return a list of the UUIDs registered in the registry
func (devices *Devices) ListDevices(runtime *net.RequestRuntime) net.HandlerResult {
	ids, e := devices.ListRegistrations()

	if e != nil {
		devices.Errorf("unable to lookup device id list: %s", e.Error())
		return runtime.ServerError()
	}

	return net.HandlerResult{Results: ids}
}

// UpdateShorthand accepts a device id and a color (via url params from the req) and updates the device to that color.
func (devices *Devices) UpdateShorthand(runtime *net.RequestRuntime) net.HandlerResult {
	query, color := runtime.Get("uuid"), runtime.Get("color")
	details, e := devices.FindDevice(query)

	if e != nil {
		devices.Warnf("shorthand update w/ invalid device id: %s (%s)", query, e.Error())
		return runtime.LogicError("not-found")
	}

	token := runtime.HeaderValue(defs.APIUserTokenHeader)

	if token == "" || devices.AuthorizeToken(details.DeviceID, token, controllerPermission) != true {
		devices.Warnf("unauthorized attempt to control device (token: %s, device: %s)", token, details.DeviceID)
		return runtime.LogicError("invalid-token")
	}

	frame := interchange.ControlFrame{}

	switch {
	case color == "green":
		frame.Green = 255
	case color == "red":
		frame.Red = 255
	case color == "blue":
		frame.Blue = 255
	case color == "rand":
		frame = interchange.ControlFrame{
			Red:   devices.randColorValue(),
			Green: devices.randColorValue(),
			Blue:  devices.randColorValue(),
		}
	case hexColorRegex.MatchString(color):
		r, g, b := color[0:2], color[2:4], color[4:6]
		buff := make([]byte, 1)

		if _, e := hex.Decode(buff, []byte(r)); e != nil {
			devices.Warnf("[warn] invalid hex received: %s", e.Error())
			return runtime.LogicError("invalid-hex")
		}

		frame.Red = uint32(buff[0])

		if _, e := hex.Decode(buff, []byte(g)); e != nil {
			devices.Warnf("[warn] invalid hex received: %s", e.Error())
			return runtime.LogicError("invalid-hex")
		}

		frame.Green = uint32(buff[0])

		if _, e := hex.Decode(buff, []byte(b)); e != nil {
			devices.Warnf("[warn] invalid hex received: %s", e.Error())
			return runtime.LogicError("invalid-hex")
		}

		frame.Blue = uint32(buff[0])

		devices.Debugf("received rgb color: rgb(%d,%d,%d)", frame.Red, frame.Green, frame.Blue)
	default:
		devices.Debugf("received tricky color: %s", color)
	}

	commandData, e := proto.Marshal(&interchange.ControlMessage{
		Frames: []*interchange.ControlFrame{&frame},
	})

	if e != nil {
		return net.HandlerResult{Errors: []error{e}}
	}

	message := interchange.DeviceMessage{
		Type: interchange.DeviceMessageType_CONTROL,
		Authentication: &interchange.DeviceMessageAuthentication{
			DeviceID: details.DeviceID,
		},
		Payload: commandData,
	}

	devices.Debugf("attempting to update device %s to %s", details.DeviceID, color)

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
