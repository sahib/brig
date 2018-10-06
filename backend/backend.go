package backend

import (
	"errors"
	"io"
	"os"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/sahib/brig/backend/ipfs"
	"github.com/sahib/brig/backend/mock"
	"github.com/sahib/brig/catfs"
	netBackend "github.com/sahib/brig/net/backend"
	"github.com/sahib/brig/repo"
)

var (
	ErrNoSuchBackend = errors.New("No such backend")
)

type VersionInfo interface {
	SemVer() string
	Name() string
	Rev() string
}

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

func ForwardLogByName(name string, w io.Writer) error {
	switch name {
	case "ipfs":
		ipfs.ForwardLog(w)
		return nil
	case "mock":
		return nil
	}

	return ErrNoSuchBackend
}

// FromName returns a suitable backend for a human readable name.
// If an invalid name is passed, nil is returned.
func FromName(name, path string, bootstrapNodes []string) (Backend, error) {
	switch name {
	case "ipfs":
		return ipfs.New(path, bootstrapNodes)
	case "mock":
		// This is silly, but it's only for testing.
		// Read the name and the port from the backend path.
		// Side effect: user cannot contain slashes currently.
		port := 9995
		if envPort := os.Getenv("BRIG_MOCK_PORT"); envPort != "" {
			newPort, err := strconv.Atoi(envPort)
			if err != nil {
				log.Warningf("Failed to parse BRIG_MOCK_PORT=%s: %s", envPort, err)
			} else {
				port = newPort
			}
		}

		user := "alice"
		if envUser := os.Getenv("BRIG_MOCK_USER"); envUser != "" {
			user = envUser
		}

		if envNetDbPath := os.Getenv("BRIG_MOCK_NET_DB_PATH"); envNetDbPath != "" {
			path = envNetDbPath
		}

		return mock.NewMockBackend(path, user, port), nil
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

func Version(name string) VersionInfo {
	switch name {
	case "ipfs":
		return ipfs.Version()
	case "mock":
		return mock.Version()
	default:
		return nil
	}
}
