package backend

import (
	"fmt"

	"github.com/disorganizer/brig/backend/ipfs"
	"github.com/disorganizer/brig/backend/mock"
	"github.com/disorganizer/brig/catfs"
	netBackend "github.com/disorganizer/brig/net/backend"
	"github.com/disorganizer/brig/repo"
)

// Backend is a amalgamation of all backend interfaces required for brig to work.
type Backend interface {
	repo.Backend
	catfs.FsBackend
	netBackend.Backend
}

// FromName returns a suitable backend for a human readable name.
// If an invalid name is passed, nil is returned.
func FromName(name, path string) (Backend, error) {
	switch name {
	case "ipfs":
		return ipfs.New(path)
	case "mock":
		return mock.NewMockBackend(), nil
	}

	return nil, fmt.Errorf("No such backend `%s`", name)
}

func IsValidName(name string) bool {
	switch name {
	case "ipfs", "mock":
		return true
	default:
		return false
	}
}
