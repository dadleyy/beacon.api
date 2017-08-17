package routes

import "fmt"
import "github.com/dadleyy/beacon.api/beacon/net"
import "github.com/dadleyy/beacon.api/beacon/defs"
import "github.com/dadleyy/beacon.api/beacon/device"
import "github.com/dadleyy/beacon.api/beacon/logging"

// NewTokensAPI inititalizes a new token api.
func NewTokensAPI(store device.TokenStore, index device.Index) *Tokens {
	logger := logging.New(defs.TokensAPILogPrefix, logging.Green)
	return &Tokens{logger, store, index}
}

type tokenRequest struct {
	DeviceID string `json:"device_id"`
	Name     string `json:"name"`
}

// Tokens defines the api for creating/deleting device auth tokens.
type Tokens struct {
	logging.LeveledLogger
	device.TokenStore
	device.Index
}

// Create authenticates the incoming request and attempts to allocate a new auth token.
func (tokens *Tokens) Create(runtime *net.RequestRuntime) net.HandlerResult {
	request := tokenRequest{}

	if e := runtime.ReadBody(&request); e != nil {
		tokens.Warnf("received invalid request: %s", e.Error())
		return runtime.LogicError("invalid-request")
	}

	if valid := len(request.Name) >= 5; valid != true {
		return runtime.LogicError("invalid-name")
	}

	registration, e := tokens.FindDevice(request.DeviceID)

	if e != nil {
		tokens.Warnf("unable to find device (device id: %s): %s", request.DeviceID, e.Error())
		return runtime.LogicError("not-found")
	}

	token := runtime.HeaderValue(defs.APIUserTokenHeader)

	if token == "" {
		tokens.Warnf("attempt to create token w/o auth for device %s", registration.DeviceID)
		return runtime.LogicError("invalid-token")
	}

	if token == registration.SharedSecret {
		tokens.Debugf("creating device token via shared secret for device %s", registration.DeviceID)
		return tokens.create(registration.DeviceID, request.Name)
	}

	details, e := tokens.FindToken(token)

	if e != nil {
		tokens.Warnf("unable to find token: %s", e.Error())
		return runtime.LogicError("not-found")
	}

	if details.DeviceID != registration.DeviceID {
		tokens.Warnf("token mismatch: %s", e.Error())
		return runtime.LogicError("not-found")
	}

	tokens.Infof("creating token for device: %s", registration.DeviceID)

	return net.HandlerResult{}
}

func (tokens *Tokens) create(deviceID, name string) net.HandlerResult {
	token, e := tokens.CreateToken(deviceID, name)

	if e != nil {
		tokens.Warnf("unable to create token: %s (got %v)", e.Error(), token)
		return net.HandlerResult{Errors: []error{fmt.Errorf("server-error")}}
	}

	tokens.Debugf("created token: %s", token)

	return net.HandlerResult{}
}
