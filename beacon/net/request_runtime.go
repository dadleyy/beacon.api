package net

import "io"
import "fmt"
import "net/url"
import "net/http"
import "encoding/json"
import "github.com/gorilla/websocket"

import "github.com/dadleyy/beacon.api/beacon/defs"
import "github.com/dadleyy/beacon.api/beacon/logging"

// RequestRuntime is used by the ServerRuntime to expose per-request packages of shared system interfaces
type RequestRuntime struct {
	url.Values
	websocket.Upgrader
	*logging.Logger
	*http.Request

	responseWriter     http.ResponseWriter
	backgroundChannels defs.BackgroundChannels
}

// ReadBody will attempt to fill the provided interface with values from the http request
func (runtime *RequestRuntime) ReadBody(target interface{}) error {
	decoder := json.NewDecoder(runtime.Request.Body)

	if e := decoder.Decode(target); e != nil {
		return e
	}

	return nil
}

// ServerError returns a HandlerResult w/ the standardized server error response text
func (runtime *RequestRuntime) ServerError() HandlerResult {
	return HandlerResult{Errors: []error{fmt.Errorf("server-error")}}
}

// LogicError will wrap the provided strin the appropriate error prefix and return a HandlerResult
func (runtime *RequestRuntime) LogicError(message string) HandlerResult {
	return HandlerResult{Errors: []error{fmt.Errorf(message)}}
}

// Publish sends the provided Reader item into the given channel, returning a boolean indicating if the channel exists
func (runtime *RequestRuntime) Publish(channelName string, message io.Reader) bool {
	s, ok := runtime.backgroundChannels[channelName]

	if ok != true {
		return false
	}

	s <- message
	return true
}

// Websocket attempts to updrade the request to a websocket connection
func (runtime *RequestRuntime) Websocket() (*websocket.Conn, error) {
	upgrader, responseWriter, request := runtime.Upgrader, runtime.responseWriter, runtime.Request
	return upgrader.Upgrade(responseWriter, request, nil)
}
