package defs

const (
	// The regular expression used for the device shorthand route
	DeviceShorthandRoute = "^/devices/(?P<uuid>[\\d\\w\\-]+)/(?P<color>red|blue|green)$"
)
