package routes

import "github.com/dadleyy/beacon.api/beacon/net"
import "github.com/dadleyy/beacon.api/beacon/device"

// NewTokensAPI inititalizes a new token api.
func NewTokensAPI(store device.TokenStore) *Tokens {
	return &Tokens{store}
}

// Tokens defines the api for creating/deleting device auth tokens.
type Tokens struct {
	device.TokenStore
}

// CreateToken authenticates the incoming request and attempts to allocate a new auth token.
func (tokens *Tokens) CreateToken(runtime *net.RequestRuntime) net.HandlerResult {
	return net.HandlerResult{}
}
