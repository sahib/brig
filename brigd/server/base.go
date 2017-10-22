package server

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/backend"
	"github.com/disorganizer/brig/catfs"
	"github.com/disorganizer/brig/repo"
)

type base struct {
	mu       sync.Mutex
	basePath string

	// password used to lock/unlock the repo.
	// This is currently stored until end of the daemon,
	// which is not optimal. Measures needs to be taken
	// to secure access to Password here.
	password string

	repo    *repo.Repository
	backend backend.Backend

	QuitCh chan struct{}
}

func repoIsInitialized(path string) error {
	data, err := ioutil.ReadFile(filepath.Join(path, "meta.yml"))
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return fmt.Errorf("meta.yml is empty")
	}

	return nil
}

// Repo lazily-loads the repository on disk.
// On the next call it will be returned directly.
func (b *base) Repo() (*repo.Repository, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.repo != nil {
		return b.repo, nil
	}

	// Sanity check, so that we do not call a repo command without
	// an initialized repo. Error early for a meaningful message here.
	if err := repoIsInitialized(b.basePath); err != nil {
		msg := fmt.Sprintf(
			"Repo does not look it is initialized: %v (did you brig init?)",
			err,
		)
		log.Warning(msg)
		return nil, errors.New(msg)
	}

	rp, err := repo.Open(b.basePath, b.password)
	if err != nil {
		log.Warningf("Failed to load repository at `%s`: %v", b.basePath, err)
		return nil, err
	}

	b.repo = rp
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

	b.backend = bk
	return bk, nil
}

func newBase(basePath string, password string) (*base, error) {
	return &base{
		basePath: basePath,
		password: password,
		QuitCh:   make(chan struct{}, 1),
	}, nil
}

func (b *base) withOwnFs(fn func(fs *catfs.FS) error) error {
	rp, err := b.Repo()
	if err != nil {
		return err
	}

	bk, err := b.Backend()
	if err != nil {
		return err
	}

	fs, err := rp.OwnFS(bk)
	if err != nil {
		return err
	}

	return fn(fs)
}

func (b *base) withRemoteFs(owner string, fn func(fs *catfs.FS) error) error {
	rp, err := b.Repo()
	if err != nil {
		return err
	}

	bk, err := b.Backend()
	if err != nil {
		return err
	}

	fs, err := rp.FS(owner, bk)
	if err != nil {
		return err
	}

	return fn(fs)
}
