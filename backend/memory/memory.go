package memory

import (
	"github.com/disorganizer/brig/catfs"
	netMemory "github.com/disorganizer/brig/net/backend/memory"
)

type MemRepoBackend struct{}

func NewMemRepoBackend() *MemRepoBackend {
	return &MemRepoBackend{}
}

func (db *MemRepoBackend) Init(path string) error {
	// Nothing persistent needed for memory only
	return nil
}

type MemoryBackend struct {
	catfs.MemFsBackend
	MemRepoBackend
	*netMemory.NetBackend
}

func NewMemoryBackend() *MemoryBackend {
	return &MemoryBackend{
		MemFsBackend:   *catfs.NewMemFsBackend(),
		MemRepoBackend: *NewMemRepoBackend(),
		NetBackend:     netMemory.NewNetBackend(),
	}
}
