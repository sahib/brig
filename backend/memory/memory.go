package memory

import (
	"github.com/disorganizer/brig/catfs"
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
}

func NewMemoryBackend() *MemoryBackend {
	return &MemoryBackend{
		MemFsBackend:   *catfs.NewMemFsBackend(),
		MemRepoBackend: *NewMemRepoBackend(),
	}
}
