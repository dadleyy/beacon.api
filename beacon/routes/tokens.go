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
	DeviceID   string `json:"device_id"`
	Name       string `json:"name"`
	Permission uint   `json:"permission"`
}

// Tokens defines the api for creating/deleting device auth tokens.
type Tokens struct {
	logging.LeveledLogger
	device.TokenStore
	device.Index
}

// CreateToken authenticates the incoming request and attempts to allocate a new auth token.
func (tokens *Tokens) CreateToken(requestRuntime *net.RequestRuntime) net.HandlerResult {
	request := tokenRequest{}

	if e := requestRuntime.ReadBody(&request); e != nil {
		tokens.Warnf("received invalid request: %s", e.Error())
		return requestRuntime.LogicError("invalid-request")
	}

	if request.Permission&defs.SecurityDeviceTokenPermissionAll == 0 {
		tokens.Infof("no permission found - defaulting to viewer")
		request.Permission = defs.SecurityDeviceTokenPermissionViewer
	}

	if valid := len(request.Name) >= 5; valid != true {
		return requestRuntime.LogicError("invalid-name")
	}

	registration, e := tokens.FindDevice(request.DeviceID)

	if e != nil {
		tokens.Warnf("unable to find device (device id: %s): %s", request.DeviceID, e.Error())
		return requestRuntime.LogicError("not-found")
	}

	token := requestRuntime.HeaderValue(defs.APIUserTokenHeader)

	if token == "" {
		tokens.Warnf("attempt to create token w/o auth for device %s", registration.DeviceID)
		return requestRuntime.LogicError("invalid-token")
	}

	// Attempt to authorize the provided token against the admin permission.
	if tokens.AuthorizeToken(registration.DeviceID, token, defs.SecurityDeviceTokenPermissionAdmin) != true {
		tokens.Warnf("unauthorized attempt to create token (token: %s, device: %s)", token, registration.DeviceID)
		return requestRuntime.LogicError("invalid-token")
	}

	tokens.Debugf("creating device token for device %s (permission: %b)", registration.DeviceID, request.Permission)
	return tokens.create(registration.DeviceID, request.Name, request.Permission)
}

// ListTokens returns a set tokens based on the device id provided.
func (tokens *Tokens) ListTokens(requestRuntime *net.RequestRuntime) net.HandlerResult {
	id := requestRuntime.GetQueryParam("device_id")

	if id == "" {
		return requestRuntime.LogicError("invalid-device-id")
	}

	registration, e := tokens.FindDevice(id)

	if e != nil {
		return requestRuntime.LogicError("not-found")
	}

	token := requestRuntime.HeaderValue(defs.APIUserTokenHeader)

	if token == "" {
		tokens.Warnf("attempt to create token w/o auth for device %s", registration.DeviceID)
		return requestRuntime.LogicError("invalid-token")
	}

	// Attempt to authorize the provided token against the admin permission.
	if tokens.AuthorizeToken(registration.DeviceID, token, defs.SecurityDeviceTokenPermissionAdmin) != true {
		tokens.Warnf("unauthorized attempt to create token (token: %s, device: %s)", token, registration.DeviceID)
		return requestRuntime.LogicError("invalid-token")
	}

	deviceTokens, e := tokens.TokenStore.ListTokens(registration.DeviceID)

	if e != nil {
		tokens.Errorf("invalid response from token lookup: %s", e.Error())
		return requestRuntime.ServerError()
	}

	return net.HandlerResult{Results: deviceTokens}
}

func (tokens *Tokens) create(deviceID, name string, permission uint) net.HandlerResult {
	token, e := tokens.TokenStore.CreateToken(deviceID, name, permission)

	if e != nil {
		tokens.Warnf("unable to create token: %s (got %v)", e.Error(), token)
		return net.HandlerResult{Errors: []error{fmt.Errorf("server-error")}}
	}

	tokens.Debugf("created token: %v", token)

	return net.HandlerResult{Results: []device.TokenDetails{token}}
}
