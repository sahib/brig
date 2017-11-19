package server

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"path/filepath"
	"sync"
	"time"

	"zombiezen.com/go/capnproto2/rpc"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/backend"
	"github.com/disorganizer/brig/brigd/capnp"
	"github.com/disorganizer/brig/catfs"
	"github.com/disorganizer/brig/fuse"
	peernet "github.com/disorganizer/brig/net"
	"github.com/disorganizer/brig/repo"
)

type base struct {
	mu sync.Mutex

	// base path to the repository (i.e. BRIG_PATH)
	basePath string

	// password used to lock/unlock the repo.
	// This is currently stored until end of the daemon,
	// which is not optimal. Measures needs to be taken
	// to secure access to Password here.
	password string

	ctx context.Context

	repo       *repo.Repository
	mounts     *fuse.MountTable
	peerServer *peernet.Server

	// This the general backend, not a specific submodule one:
	backend backend.Backend

	// This channel is triggered once the QUIT command was received
	// (or if a deadly/terminating signal was received)
	quitCh chan struct{}
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

// Handle is being called by the base server implementation
// for every local request that is being served to the brig daemon.
func (b *base) Handle(ctx context.Context, conn net.Conn) {
	transport := rpc.StreamTransport(conn)
	srv := capnp.API_ServerToClient(newApiHandler(b))
	rpcConn := rpc.NewConn(transport, rpc.MainInterface(srv.Client))

	if err := rpcConn.Wait(); err != nil {
		log.Warnf("Serving rpc failed: %v", err)
	}

	if err := rpcConn.Close(); err != nil {
		log.Warnf("Failed to close rpc conn: %v", err)
	}
}

/////////

// Repo lazily-loads the repository on disk.
// On the next call it will be returned directly.
func (b *base) Repo() (*repo.Repository, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.repoUnlocked()
}

func (b *base) repoUnlocked() (*repo.Repository, error) {
	if b.repo != nil {
		return b.repo, nil
	}

	return b.loadRepo()
}

func (b *base) loadRepo() (*repo.Repository, error) {
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

/////////

func (b *base) Backend() (backend.Backend, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.backendUnlocked()
}

func (b *base) backendUnlocked() (backend.Backend, error) {
	if b.backend != nil {
		return b.backend, nil
	}

	return b.loadBackend()
}

func (b *base) loadBackend() (backend.Backend, error) {
	rp, err := b.repoUnlocked()
	if err != nil {
		return nil, err
	}

	backendName := rp.BackendName()
	log.Infof("Loading backend `%s`", backendName)

	realBackend, err := backend.FromName(backendName, b.basePath)
	if err != nil {
		log.Errorf("Failed to load backend: %v", err)
		return nil, err
	}

	b.backend = realBackend
	return realBackend, nil
}

/////////

func (b *base) PeerServer() (*peernet.Server, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.peerServerUnlocked()
}

func (b *base) peerServerUnlocked() (*peernet.Server, error) {
	if b.peerServer != nil {
		return b.peerServer, nil
	}

	return b.loadPeerServer()
}

func (b *base) loadPeerServer() (*peernet.Server, error) {
	bk, err := b.backendUnlocked()
	if err != nil {
		return nil, err
	}

	log.Infof("Launching peer server...")
	srv, err := peernet.NewServer(bk)
	if err != nil {
		return nil, err
	}

	go func() {
		if err := srv.Serve(); err != nil {
			log.Warningf("PeerServer.Serve() returned with error: %v", err)
		}
	}()

	b.peerServer = srv

	time.Sleep(50 * time.Millisecond)
	return srv, nil
}

/////////

func (b *base) Mounts() (*fuse.MountTable, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.mountsUnlocked()
}

func (b *base) mountsUnlocked() (*fuse.MountTable, error) {
	if b.mounts != nil {
		return b.mounts, nil
	}

	return b.loadMounts()
}

func (b *base) loadMounts() (*fuse.MountTable, error) {
	err := b.withOwnFs(func(fs *catfs.FS) error {
		b.mounts = fuse.NewMountTable(fs)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return b.mounts, nil
}

func (b *base) withOwnFs(fn func(fs *catfs.FS) error) error {
	rp, err := b.repoUnlocked()
	if err != nil {
		return err
	}

	bk, err := b.backendUnlocked()
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
	rp, err := b.repoUnlocked()
	if err != nil {
		return err
	}

	bk, err := b.backendUnlocked()
	if err != nil {
		return err
	}

	fs, err := rp.FS(owner, bk)
	if err != nil {
		return err
	}

	return fn(fs)
}

func (b *base) Quit() (err error) {
	log.Info("Shutting down brigd due to QUIT command")
	b.quitCh <- struct{}{}

	if b.peerServer != nil {
		log.Infof("Closing peer server...")
		if err = b.peerServer.Close(); err != nil {
			log.Warningf("Failed to close peer server: %v", err)
		}
	}

	log.Infof("Trying to lock repository...")

	var rp *repo.Repository
	rp, err = b.Repo()
	if err != nil {
		log.Warningf("Failed to access repository: %v", err)
	}

	if err = rp.Close(b.password); err != nil {
		log.Warningf("Failed to lock repository: %v", err)
	}

	log.Infof("Trying to unmount any mounts...")

	var mounts *fuse.MountTable
	mounts, err = b.Mounts()
	if err != nil {
		return err
	}

	if err := mounts.Close(); err != nil {
		return err
	}

	log.Infof("brigd can be considered dead now!")
	return nil
}

func newBase(basePath string, password string, ctx context.Context) (*base, error) {
	return &base{
		ctx:      ctx,
		basePath: basePath,
		password: password,
		quitCh:   make(chan struct{}, 1),
	}, nil
}
