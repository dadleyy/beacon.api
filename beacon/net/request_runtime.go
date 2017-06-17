package net

import "log"
import "net/url"
import "net/http"
import "github.com/gorilla/websocket"

type RequestRuntime struct {
	url.Values
	websocket.Upgrader
	*log.Logger

	responseWriter http.ResponseWriter
	request        *http.Request
}

func (runtime *RequestRuntime) Websocket() (*websocket.Conn, error) {
	upgrader, responseWriter, request := runtime.Upgrader, runtime.responseWriter, runtime.request
	return upgrader.Upgrade(responseWriter, request, nil)
}
