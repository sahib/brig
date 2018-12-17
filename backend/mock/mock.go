package mock

import (
	"fmt"

	"github.com/sahib/brig/catfs"
	eventsMock "github.com/sahib/brig/events/mock"
	netMock "github.com/sahib/brig/net/mock"
	repoMock "github.com/sahib/brig/repo/mock"
)

// Backend is used for local testing.
type Backend struct {
	*catfs.MemFsBackend
	*repoMock.RepoBackend
	*netMock.NetBackend
	*eventsMock.EventsBackend
}

// NewMockBackend returns a backend.Backend that operates only in memory
// and does not use any resources outliving the own process, except the net
// part which stores connection info on disk.
func NewMockBackend(path, owner string, port int) *Backend {
	return &Backend{
		MemFsBackend:  catfs.NewMemFsBackend(),
		RepoBackend:   repoMock.NewMockRepoBackend(),
		NetBackend:    netMock.NewNetBackend(path, owner, port),
		EventsBackend: eventsMock.NewEventsBackend(fmt.Sprintf("%s-%d", owner, port)),
	}
}

// VersionInfo holds version info (yeah, golint)
type VersionInfo struct {
	semVer, name, rev string
}

// SemVer returns a version string complying semantic versioning
func (v *VersionInfo) SemVer() string { return v.semVer }

// Name returns the name of the backend
func (v *VersionInfo) Name() string { return v.name }

// Rev returns the git revision of the backend
func (v *VersionInfo) Rev() string { return v.rev }

// Version returns detailed version info as struct
func Version() *VersionInfo {
	return &VersionInfo{
		semVer: "0.0.1",
		name:   "mock",
		rev:    "HEAD",
	}
}
