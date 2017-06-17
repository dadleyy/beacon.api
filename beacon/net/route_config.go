package net

import "regexp"

type RouteConfig struct {
	Method  string
	Pattern *regexp.Regexp
}
