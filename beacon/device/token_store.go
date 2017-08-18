package device

// TokenDetails holds permission information for a given device token.
type TokenDetails struct {
	TokenID    string `json:"token_id"`
	DeviceID   string `json:"device_id"`
	Token      string `json:"token"`
	Name       string `json:"name"`
	Permission uint   `json:"permission"`
}

// TokenStore defines the interface for creating tokens.
type TokenStore interface {
	CreateToken(string, string, uint) (TokenDetails, error)
	FindToken(string) (TokenDetails, error)
	AuthorizeToken(string, string, uint) bool
}
