package defs

const (
	// APIContentTypeHeader is the content type header.
	APIContentTypeHeader = "Content-Type"

	// APIDeviceRegistrationHeader is the header key used by devices to send their shared secret when connecting.
	APIDeviceRegistrationHeader = "x-device-auth"

	// APIUserTokenHeader is the header key used by users to send a device token.
	APIUserTokenHeader = "x-user-auth"

	// APIFeedbackContentTypeHeader is the content type required for requests sent to the feedback api.
	APIFeedbackContentTypeHeader = "application/octet-stream"
)
