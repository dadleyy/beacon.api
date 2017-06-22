package net

import "net/http"

// The interface used by the `ServerRuntime` to render out the `HandlerResult` returned by the matching route's handler
type Renderer interface {
	Render(http.ResponseWriter, HandlerResult) error
}
