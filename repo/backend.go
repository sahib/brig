package repo

import "io"

// Backend defines the method needed from the underlying
// storage backend to create & manage a repository.
type Backend interface {
	// ForwardLog writes all logs of the backend to `w`.
	// The log level is chosen by the backend itself.
	ForwardLog(w io.Writer)
}
