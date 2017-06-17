package routes

import "github.com/dadleyy/beacon.api/beacon/net"

func Register(runtime *net.RequestRuntime) net.HandlerResult {
	if _, e := runtime.Websocket(); e != nil {
		runtime.Printf("[warn] unable to upgrade websocket: %s", e.Error())
		return net.HandlerResult{Errors: []error{e}}
	}

	runtime.Printf("[debug] websocket open")

	return net.HandlerResult{NoRender: true}
}
