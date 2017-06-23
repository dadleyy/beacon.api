package device

// The Index interface defines a store that is used to add, remove and lookup string based elements
type Index interface {
	Remove(string) error
	Exists(string) bool
}
