package backend

import (
	"github.com/disorganizer/brig/backend/mock"
	"github.com/disorganizer/brig/catfs"
	"github.com/disorganizer/brig/net"
	"github.com/disorganizer/brig/repo"
)

// Backend is a amalgamation of all backend interfaces required for brig to work.
type Backend interface {
	repo.Backend
	catfs.FsBackend
	net.Backend
}

// FromName returns a suitable backend for a human readable name.
// If an invalid name is passed, nil is returned.
func FromName(name string) Backend {
	switch name {
	case "ipfs":
		return nil
	case "mock":
		return mock.NewMockBackend()
	}

	return nil
}
