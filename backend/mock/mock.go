package mock

import (
	"github.com/sahib/brig/catfs"
	netMock "github.com/sahib/brig/net/mock"
	repoMock "github.com/sahib/brig/repo/mock"
)

type mockBackend struct {
	*catfs.MemFsBackend
	*repoMock.MockRepoBackend
	*netMock.NetBackend
}

// TODO: Cleanup for net backend etc.

// NewMockBackend returns a backend.Backend that operates only in memory
// and does not use any resources outliving the own process, except the net
// part which stores connection info on disk.
func NewMockBackend(owner string, port int) (*mockBackend, error) {
	nb, err := netMock.NewNetBackend(owner, port)
	if err != nil {
		return nil, err
	}

	return &mockBackend{
		MemFsBackend:    catfs.NewMemFsBackend(),
		MockRepoBackend: repoMock.NewMockRepoBackend(),
		NetBackend:      nb,
	}, nil
}

type version struct {
	semVer, name, rev string
}

func (v *version) SemVer() string { return v.semVer }
func (v *version) Name() string   { return v.name }
func (v *version) Rev() string    { return v.rev }

func Version() *version {
	return &version{
		semVer: "0.0.1",
		name:   "mock",
		rev:    "HEAD",
	}
}
