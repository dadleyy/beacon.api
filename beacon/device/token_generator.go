package device

// TokenGenerator defines an interface for generating random tokens.
type TokenGenerator interface {
	GenerateToken() (string, error)
}
