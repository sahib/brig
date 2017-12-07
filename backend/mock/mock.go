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

// NewMockBackend returns a backend.Backend that operates only in memory
// and does not use any resources outliving the own process.
func NewMockBackend() *mockBackend {
	return &mockBackend{
		MemFsBackend:    catfs.NewMemFsBackend(),
		MockRepoBackend: repoMock.NewMockRepoBackend(),
		NetBackend:      netMock.NewNetBackend(),
	}
}
