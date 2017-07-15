package device

// TokenStore defines the interface for creating tokens.
type TokenStore interface {
	CreateToken(string) (string, error)
}
