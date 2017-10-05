package server

import (
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/backend"
	"github.com/disorganizer/brig/repo"
)

type base struct {
	mu       sync.Mutex
	basePath string

	repo    *repo.Repository
	backend backend.Backend

	QuitCh chan struct{}
}

// Repo lazily-loads the repository on disk.
// On the next call it will be returned directly.
func (b *base) Repo() (*repo.Repository, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.repo != nil {
		return b.repo, nil
	}

	rp, err := repo.Open(b.basePath)
	if err != nil {
		log.Warningf("Failed to load repository at `%s`: %v", b.basePath, err)
		return nil, err
	}

	return rp, nil
}

func (b *base) Backend() (backend.Backend, error) {
	rp, err := b.Repo()
	if err != nil {
		return nil, err
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	if b.backend != nil {
		return b.backend, nil
	}

	bk, err := rp.LoadBackend()
	if err != nil {
		return nil, err
	}

	return bk, nil
}

func newBase(basePath string) (*base, error) {
	return &base{
		basePath: basePath,
		QuitCh:   make(chan struct{}, 1),
	}, nil
}
