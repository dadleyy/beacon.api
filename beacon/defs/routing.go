package defs

const (
	// DeviceListRoute is the regular expression used for the device list route
	DeviceListRoute = "^/devices$"

	// DeviceShorthandRoute is the regular expression used for the device shorthand route
	DeviceShorthandRoute = "^/devices/(?P<uuid>[\\d\\w\\-]+)/(?P<color>red|blue|green|off|rand|[0-9a-f]{6})$"

	// DeviceRegistrationRoute is used by devices to register with the server
	DeviceRegistrationRoute = "^/register$"

	// DeviceMessagesRoute is used to crate device messages
	DeviceMessagesRoute = "^/device-messages$"

	// SystemRoute prints out system information
	SystemRoute = "^/system$"
)
