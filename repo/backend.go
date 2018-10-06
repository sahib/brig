package repo

import (
	h "github.com/sahib/brig/util/hashlib"
)

// Backend defines the method needed from the underlying
// storage backend to create & manage a repository.
type Backend interface {
	GC() ([]h.Hash, error)
}
