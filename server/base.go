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
	fserrs "github.com/sahib/brig/catfs/errors"
	"github.com/sahib/brig/events"
	"github.com/sahib/brig/fuse"
	"github.com/sahib/brig/gateway"
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

	// gateway is the control object for the gateway server
	gateway *gateway.Gateway

	// evListener is a listener that will h
	evListener *events.Listener

	// evListenerCtx is the context for the event subsystem
	evListenerCtx context.Context

	// evListenerCancel can be called on quitting the daemon
	evListenerCancel context.CancelFunc
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
	srv := capnp.API_ServerToClient(newAPIHandler(b))
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

func (b *base) updateBackendAddr(reg *registry.Registry, bk backend.Backend) error {
	rp, err := b.repoUnlocked()
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

	err = b.withCurrFs(func(fs *catfs.FS) error {
		b.gateway = gateway.NewGateway(fs, rp.Config.Section("gateway"))
		b.gateway.Start()
		return nil
	})

	if err != nil {
		return nil, err
	}

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
	log.Debugf("loading peer server")
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

	self, err := bk.Identity()
	if err != nil {
		return nil, err
	}

	b.evListenerCtx, b.evListenerCancel = context.WithCancel(context.Background())
	b.evListener = events.NewListener(rp.Config.Section("events"), bk, self.Addr)
	b.evListener.RegisterEventHandler(events.FsEvent, b.handleFsEvent)
	if err := b.evListener.SetupListeners(b.evListenerCtx, addrs); err != nil {
		log.Warningf("failed to setup event listeners: %v", err)
	}

	// Give peer server a small bit of time to start up, so it can Accept()
	// connections immediately after loadPeerServer. Also nice for tests.
	time.Sleep(50 * time.Millisecond)

	if err := b.initialSyncWithAutoUpdatePeers(); err != nil {
		log.Warningf("initial sync failed with one or more peers: %v", err)
	}

	// Now that we boooted up, we should tell other users that our fs changed.
	// It may or may not have, but other remotes judge that.
	b.notifyFsChangeEvent(rp)

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

func (b *base) withFsFromPath(path string, fn func(url *URL, fs *catfs.FS) error) error {
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

	ctl, err := p2pnet.Dial(subCtx, who, rp, bk)
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

		b.evListenerCancel()
		log.Infof("Shutting down event listener...")
		if b.evListener != nil {
			if err := b.evListener.Close(); err != nil {
				log.Warningf("shutting down event handler failed: %v", err)
			}
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
	ctx context.Context,
	port int64,
	basePath string,
	password string,
	bindHost string,
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

func (b *base) doFetch(who string) error {
	rp, err := b.Repo()
	if err != nil {
		return err
	}

	if who == rp.Owner {
		log.Infof("skipping fetch for own metadata")
		return nil
	}

	return b.withNetClient(who, func(ctl *p2pnet.Client) error {
		return b.withRemoteFs(who, func(remoteFs *catfs.FS) error {
			// Not all remotes might allow doing a full fetch.
			// This is only possible when having full access to all folders.
			if isAllowed, err := ctl.IsCompleteFetchAllowed(); isAllowed && err != nil {
				log.Debugf("fetch: doing complete fetch for %s", who)
				storeBuf, err := ctl.FetchStore()
				if err != nil {
					return e.Wrapf(err, "fetch-store")
				}

				return e.Wrapf(remoteFs.Import(storeBuf), "import")
			}

			// Ask our local copy of the remote what the last patch index was.
			fromIndex, err := remoteFs.LastPatchIndex()
			if err != nil {
				return err
			}

			// Get the missing changes since then:
			log.Debugf("fetch: doing partial fetch for %s starting at %d", who, fromIndex)
			patch, err := ctl.FetchPatch(fromIndex)
			if err != nil {
				return err
			}

			return remoteFs.ApplyPatch(patch)
		})
	})
}

func (b *base) doSync(withWhom string, needFetch bool, msg string) (*catfs.Diff, error) {
	if needFetch {
		if err := b.doFetch(withWhom); err != nil {
			return nil, e.Wrapf(err, "fetch")
		}
	}

	var diff *catfs.Diff

	return diff, b.withCurrFs(func(ownFs *catfs.FS) error {
		return b.withRemoteFs(withWhom, func(remoteFs *catfs.FS) error {
			// Automatically make a commit before merging with their state:
			timeStamp := time.Now().UTC().Format(time.RFC3339)
			commitMsg := fmt.Sprintf("sync with %s on %s", withWhom, timeStamp)
			if err := ownFs.MakeCommit(commitMsg); err != nil && err != fserrs.ErrNoChange {
				return e.Wrapf(err, "merge-commit")
			}

			cmtBefore, err := ownFs.Head()
			if err != nil {
				return err
			}

			log.Debugf("Starting sync with %s", withWhom)

			if err := ownFs.Sync(remoteFs, msg); err != nil {
				return err
			}

			log.Debugf("Sync with %s done", withWhom)

			cmtAfter, err := ownFs.Head()
			if err != nil {
				return err
			}

			diff, err = ownFs.MakeDiff(ownFs, cmtBefore, cmtAfter)
			return err
		})
	})
}

func (b *base) handleFsEvent(ev *events.Event) {
	log.Debugf("received fs event: %v", ev)
	rmt, err := b.repo.Remotes.RemoteByAddr(ev.Source)
	if err != nil {
		log.Warningf("failed to resolve '%s' to a remote name: %v", ev.Source, err)
		return
	}

	log.Debugf("resolved to remote: %v", rmt)
	if !rmt.AcceptAutoUpdates {
		log.Debugf("currently not accepting events from %s", rmt.Name)
		return
	}

	log.Infof("doing sync with '%s' since we received an update notification.", rmt.Name)

	msg := fmt.Sprintf("sync due to notification from »%s«", rmt.Name)
	if _, err := b.doSync(rmt.Name, true, msg); err != nil {
		log.Warningf("sync failed: %v", err)
	}
}

func (b *base) notifyFsChangeEventLocked() {
	rp, err := b.Repo()
	if err != nil {
		log.Warningf("failed to load repo: %v", err)
		return
	}

	b.notifyFsChangeEvent(rp)
}

func (b *base) notifyFsChangeEvent(rp *repo.Repository) {
	if b.evListener == nil {
		return
	}

	// Do not trigger events when we're looking at the store of somebody else.
	if rp.Owner != rp.CurrentUser() {
		return
	}

	log.Debugf("publishing fs event")
	ev := events.Event{
		Type: events.FsEvent,
	}

	if err := b.evListener.PublishEvent(ev); err != nil {
		log.Warningf("failed to publish filesystem change event: %v", err)
	}
}

func (b *base) initialSyncWithAutoUpdatePeers() error {
	rp, err := b.repoUnlocked()
	if err != nil {
		return err
	}

	rmts, err := rp.Remotes.ListRemotes()
	if err != nil {
		return err
	}

	for _, rmt := range rmts {
		if !rmt.AcceptAutoUpdates {
			continue
		}

		msg := fmt.Sprintf("sync with »%s« due to intial auto-update", rmt.Name)
		if _, err := b.doSync(rmt.Name, true, msg); err != nil {
			log.Warningf("failed to sync initially with %s: %v", rmt.Name, err)
		}
	}

	return nil
}
