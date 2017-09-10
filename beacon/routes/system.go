package routes

import "time"
import "github.com/dadleyy/beacon.api/beacon/net"

// SystemInfo is a simple route that prints out a success result (no errors) w/ the current time in the metadata
func SystemInfo(runtime *net.RequestRuntime) net.HandlerResult {
	meta := make(map[string]interface{})

	meta["time"] = time.Now();

	return net.HandlerResult{Metadata: meta}
}
