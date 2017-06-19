package net

import "io"
import "log"
import "net/url"
import "net/http"
import "encoding/json"
import "github.com/gorilla/websocket"

import "github.com/dadleyy/beacon.api/beacon/defs"

type RequestRuntime struct {
	url.Values
	websocket.Upgrader
	*log.Logger

	responseWriter     http.ResponseWriter
	request            *http.Request
	backgroundChannels defs.BackgroundChannels
}

func (runtime *RequestRuntime) ReadBody(target interface{}) error {
	decoder := json.NewDecoder(runtime.request.Body)

	if e := decoder.Decode(target); e != nil {
		return e
	}

	return nil
}

func (runtime *RequestRuntime) Publish(channelName string, message io.Reader) bool {
	s, ok := runtime.backgroundChannels[channelName]

	if ok != true {
		return false
	}

	s <- message
	return true
}

func (runtime *RequestRuntime) Websocket() (*websocket.Conn, error) {
	upgrader, responseWriter, request := runtime.Upgrader, runtime.responseWriter, runtime.request
	return upgrader.Upgrade(responseWriter, request, nil)
}
