package server

import (
	"github.com/disorganizer/brig/catfs"
	"github.com/disorganizer/brig/repo"
)

// Backend is a mix-in of all backend interfaces used in the http server.
type Backend interface {
	repo.RepoBackend
	catfs.FsBackend
}

type DummyBackend struct {
	repo.DummyBackend
	*catfs.MemFsBackend
}

func NewDummyBackend() *DummyBackend {
	return &DummyBackend{
		MemFsBackend: catfs.NewMemFsBackend(),
		DummyBackend: repo.DummyBackend{},
	}
}

type base struct {
	Repo    *repo.Repository
	Backend Backend
}

func newBase(basePath string, backend Backend) error {
	repo, err := repo.Open(basePath, backend)
	if err != nil {
		return err
	}

	return &base{
		Repo:    repo,
		Backend: backend,
	}
}
