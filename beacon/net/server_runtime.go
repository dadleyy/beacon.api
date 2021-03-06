package net

import "fmt"
import "net/http"

import "github.com/dadleyy/beacon.api/beacon/bg"
import "github.com/dadleyy/beacon.api/beacon/defs"
import "github.com/dadleyy/beacon.api/beacon/logging"

// ServerRuntime defines the object that implments the http.Handler interface used during application startup to open
// the http server. It is also responsible for matching inbound requests with it's embedded routelist and creating the
// request runtime to be sent into the matching route handler.
type ServerRuntime struct {
	WebsocketUpgrader
	Multiplexer
	bg.ChannelPublisher
	*logging.Logger
	ApplicationVersion string
}

// ServerHTTP implmentation of the http.Handler interface method
func (runtime *ServerRuntime) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	found, params, handler := runtime.MatchRequest(request)

	result := HandlerResult{
		Errors: []error{fmt.Errorf(defs.ErrNotFound)},
		Status: 404,
	}

	runtime.Debugf("%s %s %s\n", request.Method, request.URL.Path, request.URL.Host)

	requestRuntime := RequestRuntime{
		Values:            params,
		WebsocketUpgrader: runtime.WebsocketUpgrader,
		Logger:            runtime.Logger,
		Request:           request,
		ChannelPublisher:  runtime.ChannelPublisher,

		responseWriter: responseWriter,
	}

	if found == true {
		result = handler(&requestRuntime)
	}

	if len(result.Redirect) >= 1 {
		responseWriter.Header().Set("Location", result.Redirect)
		responseWriter.WriteHeader(http.StatusTemporaryRedirect)
		return
	}

	var renderer Renderer

	if result.NoRender {
		runtime.Debugf("skipping server runtime render, response already sent")
		return
	}

	switch request.Header.Get("accepts") {
	default:
		renderer = &JSONRenderer{
			version: runtime.ApplicationVersion,
		}
	}

	if e := renderer.Render(responseWriter, result); e != nil {
		runtime.Errorf("unable to render results: %s", e.Error())
		responseWriter.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(responseWriter, "server error")
	}
}
