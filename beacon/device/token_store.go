package device

// TokenDetails holds permission information for a given device token.
type TokenDetails struct {
	DeviceID string `json:"device_id"`
	Token    string `json:"token"`
	Name     string `json:"name"`
}

// TokenStore defines the interface for creating tokens.
type TokenStore interface {
	CreateToken(string, string) (TokenDetails, error)
	FindToken(string) (TokenDetails, error)
}
