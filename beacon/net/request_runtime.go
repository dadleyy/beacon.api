package net

import "fmt"
import "net/url"
import "net/http"
import "encoding/json"

import "github.com/dadleyy/beacon.api/beacon/bg"
import "github.com/dadleyy/beacon.api/beacon/defs"
import "github.com/dadleyy/beacon.api/beacon/logging"

// RequestRuntime is used by the ServerRuntime to expose per-request packages of shared system interfaces
type RequestRuntime struct {
	url.Values
	WebsocketUpgrader
	bg.ChannelPublisher
	*logging.Logger
	*http.Request

	responseWriter http.ResponseWriter
}

// GetQueryParam returns a parsed url.Values struct from the request query params.
func (runtime *RequestRuntime) GetQueryParam(queryParam string) string {
	values, e := url.ParseQuery(runtime.Request.URL.RawQuery)

	if e != nil {
		return ""
	}

	return values.Get(queryParam)
}

// HeaderValue returns value for the given header key.
func (runtime *RequestRuntime) HeaderValue(key string) string {
	return runtime.Header.Get(key)
}

// ContentType returns the request content type from the inbound request.
func (runtime *RequestRuntime) ContentType() string {
	return runtime.Header.Get(defs.APIContentTypeHeader)
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
	return HandlerResult{Errors: []error{fmt.Errorf(defs.ErrServerError)}}
}

// LogicError will wrap the provided strin the appropriate error prefix and return a HandlerResult
func (runtime *RequestRuntime) LogicError(message string) HandlerResult {
	return HandlerResult{Errors: []error{fmt.Errorf(message)}}
}

// Websocket attempts to updrade the request to a websocket connection
func (runtime *RequestRuntime) Websocket() (defs.Streamer, error) {
	responseWriter, request := runtime.responseWriter, runtime.Request
	return runtime.UpgradeWebsocket(responseWriter, request, nil)
}
