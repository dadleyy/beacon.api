package defs

const (
	// ErrInvalidDeviceTokenName is returned when a user submits an invalid token name.
	ErrInvalidDeviceTokenName = "invalid-name"

	// ErrInvalidTokenRequest is returned from the registry when allocation is requested with bad contents.
	ErrInvalidTokenRequest = "invalid-request"

	// ErrInvalidRegistrationRequest is returned from the registry when allocation is requested with bad contents.
	ErrInvalidRegistrationRequest = "invalid-registration"

	// ErrNotFound returned when the application is unable to find the record it is looking for.
	ErrNotFound = "not-found"

	// ErrBadRedisResponse returned when unable to parse data from redis response.
	ErrBadRedisResponse = "storage-error"

	// ErrBadRequestFormat returned when api receives invalid body.
	ErrBadRequestFormat = "invalid-request-format"

	// ErrBadInterchangeData returned when unable to unmarshal interchange data.
	ErrBadInterchangeData = "interchange-error"

	// ErrBadInterchangeAuthentication returned when an interchange message has bad auth.
	ErrBadInterchangeAuthentication = "interchange-auth"

	// ErrInvalidContentType returned when clients make requests to the api with invalid data.
	ErrInvalidContentType = "invalid-content-type"

	// ErrServerError returned when an interchange message has bad auth.
	ErrServerError = "server-error"

	// ErrInvalidBackgroundChannel returned when attempting to publish to an invalid background channel
	ErrInvalidBackgroundChannel = "invalid-background-channel"

	// ErrInvalidDeviceSharedSecret returned when attempting to use an invalid shared secret during registration.
	ErrInvalidDeviceSharedSecret = "invalid-shared-secret"

	// ErrDuplicateRegistrationName returned when registering a name that already exists.
	ErrDuplicateRegistrationName = "duplicate-name"

	// ErrInvalidColorShorthand returned when the color shorthand request by the client is invalid.
	ErrInvalidColorShorthand = "invalid-color-shorthand"
)
