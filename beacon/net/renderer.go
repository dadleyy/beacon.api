package net

import "net/http"

// Renderer defines the interface used by the `ServerRuntime` to render `HandlerResult`s returned by route handlers
type Renderer interface {
	Render(http.ResponseWriter, HandlerResult) error
}
