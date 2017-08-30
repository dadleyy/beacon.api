package defs

const (
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

	// ErrServerError returned when an interchange message has bad auth.
	ErrServerError = "server-error"

	// ErrInvalidColorShorthand returned when the color shorthand request by the client is invalid.
	ErrInvalidColorShorthand = "invalid-color-shorthand"
)
