package defs

const (
	// ErrInvalidRegistrationRequest is returned from the registry when allocation is requested with bad contents.
	ErrInvalidRegistrationRequest = "invalid-registration"

	// ErrNotFound returned when the application is unable to find the record it is looking for.
	ErrNotFound = "not-found"

	// ErrBadRedisResponse returned when unable to parse data from redis response.
	ErrBadRedisResponse = "storage-error"

	// ErrBadInterchangeData returned when unable to unmarshal interchange data.
	ErrBadInterchangeData = "interchange-error"
)
