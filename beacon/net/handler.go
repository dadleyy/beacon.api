package net

// Handler much like http.HandlerFunc, these are used as "route" handlers by the ServerRuntime
type Handler func(*RequestRuntime) HandlerResult
