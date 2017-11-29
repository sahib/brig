package backend

import (
	"errors"

	"github.com/disorganizer/brig/backend/ipfs"
	"github.com/disorganizer/brig/backend/mock"
	"github.com/disorganizer/brig/catfs"
	netBackend "github.com/disorganizer/brig/net/backend"
	"github.com/disorganizer/brig/repo"
)

var (
	ErrNoSuchBackend = errors.New("No such backend")
)

// Backend is a amalgamation of all backend interfaces required for brig to work.
type Backend interface {
	repo.Backend
	catfs.FsBackend
	netBackend.Backend
}

func InitByName(name, path string) error {
	switch name {
	case "ipfs":
		return ipfs.Init(path, 2048)
	case "mock":
		return nil
	}

	return ErrNoSuchBackend
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

	return nil, ErrNoSuchBackend
}

func IsValidName(name string) bool {
	switch name {
	case "ipfs", "mock":
		return true
	default:
		return false
	}
}
