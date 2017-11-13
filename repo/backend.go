package repo

// Backend defines the method needed from the underlying
// storage backend to create & manage a repository.
type Backend interface {
	Init(path string) error

	// Name returns the human readable name of the backend
	// (for use in the config)
	Name() string
}
