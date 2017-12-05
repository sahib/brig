package repo

import "io"

// Backend defines the method needed from the underlying
// storage backend to create & manage a repository.
type Backend interface {
	ForwardLog(w io.Writer)
}
