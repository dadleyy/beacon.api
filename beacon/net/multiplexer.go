package net

import "net/http"
import "net/url"

// Multiplexer defines an interface that returns a handler, set of values and a boolean value indicating if a route
// has been found that matches the provided request.
type Multiplexer interface {
	MatchRequest(*http.Request) (bool, url.Values, Handler)
}
