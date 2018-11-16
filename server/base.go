package server

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log/syslog"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"zombiezen.com/go/capnproto2/rpc"

	log "github.com/Sirupsen/logrus"
	e "github.com/pkg/errors"
	"github.com/sahib/brig/backend"
	"github.com/sahib/brig/catfs"
	"github.com/sahib/brig/fuse"
	p2pnet "github.com/sahib/brig/net"
	"github.com/sahib/brig/repo"
	"github.com/sahib/brig/server/capnp"
	"github.com/sahib/brig/util/conductor"
	"github.com/sahib/brig/util/registry"
)

type base struct {
	mu sync.Mutex

	// port used by the local server
	port int64

	// base path to the repository (i.e. BRIG_PATH)
	basePath string

	// password used to lock/unlock the repo.
	// This is currently stored until end of the daemon,
	// which is not optimal. Measures needs to be taken
	// to secure access to Password here.
	password string

	// On what host the server is running on
	// (e.g. localhost or 0.0.0.0;
	//  running on 0.0.0.0 is discouraged, but can be
	//  useful for running it in docker)
	bindHost string

	ctx context.Context

	repo       *repo.Repository
	mounts     *fuse.MountTable
	peerServer *p2pnet.Server

	// This the general backend, not a specific submodule one:
	backend backend.Backend
	quitCh  chan struct{}

	conductor *conductor.Conductor

	// fsLoaded is set to true once the backend is
	// loaded/accessed the first time.
	fsLoaded bool

	// backendLoaded is set to true the first time Backend()
	// returned successfully.
	backendLoaded bool

	// logToStdout is true when logging to stdout was explicitly requested.
	logToStdout bool
}

func repoIsInitialized(path string) error {
	data, err := ioutil.ReadFile(filepath.Join(path, "OWNER"))
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return fmt.Errorf("OWNER is empty")
	}

	return nil
}

// Handle is being called by the base server implementation
// for every local request that is being served to the brig daemon.
func (b *base) Handle(ctx context.Context, conn net.Conn) {
	transport := rpc.StreamTransport(conn)
	srv := capnp.API_ServerToClient(newApiHandler(b))
	rpcConn := rpc.NewConn(
		transport,
		rpc.MainInterface(srv.Client),
		rpc.ConnLog(nil),
	)

	if err := rpcConn.Wait(); err != nil {
		log.Warnf("Serving rpc failed: %v", err)
	}

	if err := rpcConn.Close(); err != nil {
		// Close seems to be complaining that the conn was
		// already closed, but be safe and expect this.
		if err != rpc.ErrConnClosed {
			log.Warnf("Failed to close rpc conn: %v", err)
		}
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
			"Repo does not look it is initialized: %s (did you brig init?)",
			b.basePath,
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

	// Adjust the backend's logging output here, since this should be done
	// before actually loading the backend (which might produce logs already)
	backendName := rp.BackendName()
	logName := fmt.Sprintf("brig-%s", backendName)
	wSyslog, err := syslog.New(syslog.LOG_NOTICE, logName)
	if err != nil {
		log.Warningf("Failed to open connection to syslog for ipfs: %v", err)
		log.Warningf("Will output ipfs logs to stderr for now")
		backend.ForwardLogByName(backendName, os.Stderr)
	} else {
		backend.ForwardLogByName(backendName, wSyslog)
	}

	return rp, nil
}

/////////

func (b *base) Backend() (backend.Backend, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.backendUnlocked()
}

func (b *base) BackendWasLoaded() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.backendLoaded
}

func (b *base) backendUnlocked() (backend.Backend, error) {
	if b.backend != nil {
		return b.backend, nil
	}

	return b.loadBackend()
}

func (base *base) updateBackendAddr(reg *registry.Registry, bk backend.Backend) error {
	rp, err := base.repoUnlocked()
	if err != nil {
		return err
	}

	info, err := bk.Identity()
	if err != nil {
		return err
	}

	repoID, err := rp.RepoID()
	if err != nil {
		return err
	}

	entry, err := reg.Entry(repoID)
	if err != nil {
		return err
	}

	log.Debugf("Updating backend addr (%s) in registry...", info.Addr)
	entry.Addr = info.Addr
	return reg.Update(repoID, entry)
}

func (b *base) findBootstrapAddrs(reg *registry.Registry) ([]string, error) {
	entries, err := reg.List()
	if err != nil {
		return nil, err
	}

	seenAddrs := make(map[string]bool)
	bootstrapAddrs := []string{}
	for _, entry := range entries {
		if len(entry.Addr) > 0 {
			if seenAddrs[entry.Addr] {
				continue
			}

			seenAddrs[entry.Addr] = true
			bootstrapAddrs = append(bootstrapAddrs, entry.Addr)
		}
	}

	return bootstrapAddrs, nil
}

func (b *base) loadBackend() (backend.Backend, error) {
	rp, err := b.repoUnlocked()
	if err != nil {
		return nil, err
	}

	backendName := rp.BackendName()
	log.Infof("Loading backend `%s`", backendName)

	backendPath := rp.BackendPath(backendName)

	reg, err := registry.Open()
	if err != nil {
		return nil, err
	}

	bootstrapAddrs, err := b.findBootstrapAddrs(reg)
	if err != nil {
		return nil, err
	}

	if len(bootstrapAddrs) > 0 {
		log.Debugf("Found local bootstrap addrs: %v", bootstrapAddrs)
	}

	realBackend, err := backend.FromName(backendName, backendPath, bootstrapAddrs)
	if err != nil {
		log.Errorf("Failed to load backend: %v", err)
		return nil, err
	}

	b.backend = realBackend
	if err := b.updateBackendAddr(reg, realBackend); err != nil {
		log.Warningf("Failed to update registry with backend addr: %v", err)
	}

	b.backendLoaded = true
	return realBackend, nil
}

/////////

func (b *base) PeerServer() (*p2pnet.Server, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.peerServerUnlocked()
}

func (b *base) peerServerUnlocked() (*p2pnet.Server, error) {
	if b.peerServer != nil {
		return b.peerServer, nil
	}

	return b.loadPeerServer()
}

func (b *base) loadPeerServer() (*p2pnet.Server, error) {
	bk, err := b.backendUnlocked()
	if err != nil {
		return nil, err
	}

	rp, err := b.repoUnlocked()
	if err != nil {
		return nil, err
	}

	srv, err := p2pnet.NewServer(rp, bk)
	if err != nil {
		return nil, err
	}

	go func() {
		if err := srv.Serve(); err != nil {
			log.Warningf("PeerServer.Serve() returned with error: %v", err)
		}
	}()

	b.peerServer = srv

	// Initially sync the ping map:
	addrs := []string{}
	remotes, err := rp.Remotes.ListRemotes()
	if err != nil {
		return nil, err
	}

	for _, remote := range remotes {
		addrs = append(addrs, remote.Fingerprint.Addr())
	}

	log.Infof("syncing pingers")
	if err := srv.PingMap().Sync(addrs); err != nil {
		return nil, err
	}

	// Give peer server a small bit of time to start up, so it can Accept()
	// connections immediately after loadPeerServer. Nice for tests.
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
	err := b.withCurrFs(func(fs *catfs.FS) error {
		b.mounts = fuse.NewMountTable(fs)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return b.mounts, nil
}

func (b *base) withCurrFs(fn func(fs *catfs.FS) error) error {
	rp, err := b.repoUnlocked()
	if err != nil {
		return err
	}

	bk, err := b.backendUnlocked()
	if err != nil {
		return err
	}

	user := rp.CurrentUser()
	fs, err := rp.FS(user, bk)
	if err != nil {
		return err
	}

	b.fsLoaded = true
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

func (b *base) withFsFromPath(path string, fn func(url *Url, fs *catfs.FS) error) error {
	url, err := parsePath(path)
	if err != nil {
		return err
	}

	if url.User == "" {
		return b.withCurrFs(func(fs *catfs.FS) error {
			return fn(url, fs)
		})
	}

	return b.withRemoteFs(url.User, func(fs *catfs.FS) error {
		return fn(url, fs)
	})
}

func (b *base) withNetClient(who string, fn func(ctl *p2pnet.Client) error) error {
	rp, err := b.Repo()
	if err != nil {
		return err
	}

	bk, err := b.Backend()
	if err != nil {
		return err
	}

	subCtx, cancel := context.WithCancel(b.ctx)
	defer cancel()

	ctl, err := p2pnet.Dial(who, rp, bk, subCtx)
	if err != nil {
		return e.Wrapf(err, "dial")
	}

	if err := fn(ctl); err != nil {
		ctl.Close()
		return err
	}

	return ctl.Close()
}

func (b *base) Quit() (err error) {
	log.Info("Shutting down brigd due to QUIT command")

	if b.peerServer != nil {
		log.Infof("Closing peer server...")
		if err = b.peerServer.Close(); err != nil {
			log.Warningf("Failed to close peer server: %v", err)
		}
	}

	log.Infof("Trying to lock repository...")

	rp, err := b.Repo()
	if err != nil {
		log.Warningf("Failed to access repository: %v", err)
	}

	if rp != nil {
		if err = rp.Close(b.password); err != nil {
			log.Warningf("Failed to lock repository: %v", err)
		}
	}

	log.Infof("Trying to unmount any mounts...")

	var mounts *fuse.MountTable

	// Only unmount things when we used the backend.
	// Otherwise we might load the backend implicitly
	// when doing unmounting which slows the shutdown process down.
	if b.fsLoaded {
		mounts, err = b.Mounts()
		if err != nil {
			return err
		}

		if err := mounts.Close(); err != nil {
			return err
		}
	}

	log.Infof("===== brigd can be considered dead now! ====")
	return nil
}

func newBase(
	port int64,
	basePath string,
	password string,
	bindHost string,
	ctx context.Context,
	quitCh chan struct{},
	logToStdout bool,
) (*base, error) {
	return &base{
		ctx:         ctx,
		port:        port,
		basePath:    basePath,
		password:    password,
		bindHost:    bindHost,
		quitCh:      quitCh,
		logToStdout: logToStdout,
		conductor:   conductor.New(5*time.Minute, 100),
	}, nil
}
