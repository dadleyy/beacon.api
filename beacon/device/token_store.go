package device

// TokenDetails holds permission information for a given device token.
type TokenDetails struct {
	DeviceID string
}

// TokenStore defines the interface for creating tokens.
type TokenStore interface {
	CreateToken(string, string) (string, error)
	FindToken(string) (TokenDetails, error)
}
