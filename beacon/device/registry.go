package device

// The Registry interface defines a store that is used to add, remove and lookup string based elements
type Registry interface {
	Remove(string) error
	Exists(string) bool
	Insert(string) error
}
