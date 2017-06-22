package net

import "fmt"
import "log"
import "net/http"

import "github.com/gorilla/websocket"
import "github.com/garyburd/redigo/redis"

import "github.com/dadleyy/beacon.api/beacon/defs"

// ServerRuntime defines the object that implments the http.Handler interface used during application startup to open
// the http server. It is also responsible for matching inbound requests with it's embedded routelist and creating the
// request runtime to be sent into the matching route handler.
type ServerRuntime struct {
	websocket.Upgrader
	RouteList
	*log.Logger

	BackgroundChannels defs.BackgroundChannels
	RedisConnection    redis.Conn
}

// ServerHTTP implmentation of the http.Handler interface method
func (runtime *ServerRuntime) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	found, params, handler := runtime.match(request)
	result := HandlerResult{Errors: []error{fmt.Errorf("not-found")}}

	runtime.Printf("%s %s %s\n", request.Method, request.URL.Path, request.URL.Host)

	requestRuntime := RequestRuntime{
		Values:   params,
		Upgrader: runtime.Upgrader,
		Logger:   runtime.Logger,

		responseWriter:     responseWriter,
		request:            request,
		backgroundChannels: runtime.BackgroundChannels,
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
		runtime.Printf("not rendering")
		return
	}

	switch request.Header.Get("accepts") {
	default:
		renderer = &JSONRenderer{"0.0.1"}
	}

	if e := renderer.Render(responseWriter, result); e != nil {
		runtime.Fatalf("unable to render results: %s", e.Error())
		responseWriter.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(responseWriter, "server error")
	}
}
