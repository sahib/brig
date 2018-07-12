package backend

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

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

// FromName returns a suitable backend for a human readable name.
// If an invalid name is passed, nil is returned.
func FromName(name, path string) (Backend, error) {
	switch name {
	case "ipfs":
		return ipfs.New(path)
	case "mock":
		// This is silly, but it's only for testing.
		// Read the name and the port from the backend path.
		// Side effect: user cannot contain slashes currently.
		patt := regexp.MustCompile(`/user=(.*)-port=(\d+)`)
		match := patt.FindStringSubmatch(path)
		if match == nil {
			return nil, fmt.Errorf(
				"test error: please encode the user name and port in the path",
			)
		}

		user := match[1]
		port, err := strconv.Atoi(match[2])
		if err != nil {
			return nil, fmt.Errorf(
				"invalid mock addr port: %s %v",
				path,
				err,
			)
		}

		path = strings.Replace(path, match[0], "/", 1)
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
