package routes

import "time"
import "github.com/dadleyy/beacon.api/beacon/net"

func System(runtime *net.RequestRuntime) net.HandlerResult {
	meta := make(map[string]interface{})

	meta["time"] = time.Now()

	return net.HandlerResult{Metadata: meta}
}
