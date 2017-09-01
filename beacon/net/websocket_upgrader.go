package net

import "net/http"
import "github.com/dadleyy/beacon.api/beacon/defs"

// WebsocketUpgrader defines an interface that upgrades an http request to a streamer interface.
type WebsocketUpgrader interface {
	UpgradeWebsocket(http.ResponseWriter, *http.Request, http.Header) (defs.Streamer, error)
}
