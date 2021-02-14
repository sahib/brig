package backend

import (
	"errors"
	"io"
	"os"

	"github.com/sahib/brig/backend/httpipfs"
	"github.com/sahib/brig/backend/mock"
	"github.com/sahib/brig/catfs"
	eventsBackend "github.com/sahib/brig/events/backend"
	netBackend "github.com/sahib/brig/net/backend"
	"github.com/sahib/brig/repo"
	log "github.com/sirupsen/logrus"
)

var (
	// ErrNoSuchBackend is returned when passing an invalid backend name
	ErrNoSuchBackend = errors.New("No such backend")
)

// VersionInfo is a small interface that will return version info about the
// backend.
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
	eventsBackend.Backend
}

// ForwardLogByName will forward the logs of the backend `name` to `w`.
func ForwardLogByName(name string, w io.Writer) error {
	switch name {
	case "httpipfs":
		return nil
	case "mock":
		return nil
	}

	return ErrNoSuchBackend
}

// FromName returns a suitable backend for a human readable name.
// If an invalid name is passed, nil is returned.
func FromName(name, path, fingerprint string) (Backend, error) {
	switch name {
	case "httpipfs":
		return httpipfs.NewNode(path, fingerprint)
	case "mock":
		user := "alice"
		if envUser := os.Getenv("BRIG_MOCK_USER"); envUser != "" {
			user = envUser
		}

		if envNetDbPath := os.Getenv("BRIG_MOCK_NET_DB_PATH"); envNetDbPath != "" {
			path = envNetDbPath
		}

		return mock.NewMockBackend(path, user), nil
	}

	return nil, ErrNoSuchBackend
}

// Version returns version info for the backend `name`.
func Version(name, path string) VersionInfo {
	switch name {
	case "mock":
		return mock.Version()
	case "httpipfs":
		nd, err := httpipfs.NewNode(path, "")
		if err != nil {
			log.Debugf("failed to get version")
			return nil
		}

		defer nd.Close()
		return nd.Version()
	default:
		return nil
	}
}
