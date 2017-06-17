package net

import "net/http"

type Renderer interface {
	Render(http.ResponseWriter, HandlerResult) error
}
