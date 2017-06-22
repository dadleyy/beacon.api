package net

import "regexp"

// RouteConfig defines a simple structure composed of the http method and a regular expression path matcher
type RouteConfig struct {
	Method  string
	Pattern *regexp.Regexp
}
