package backend

import (
	"github.com/disorganizer/brig/backend/memory"
	"github.com/disorganizer/brig/catfs"
	netBackend "github.com/disorganizer/brig/net/backend"
)

type RepoBackend interface {
	Init(path string) error
}

// Backend is a amalgamation of all backend interfaces required for brig to work.
type Backend interface {
	RepoBackend
	catfs.FsBackend
	netBackend.Backend
}

// FromName returns a suitable backend for a human readable name.
// If an invalid name is passed, nil is returned.
func FromName(name string) Backend {
	switch name {
	case "ipfs":
		return nil
	case "memory":
		return memory.NewMemoryBackend()
	}

	return nil
}
