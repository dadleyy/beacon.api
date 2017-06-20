package device

type Registry interface {
	Remove(string) error
	Exists(string) bool
	Insert(string) error
}
