package server

import (
	"context"
	"net"
	"runtime/debug"

	"github.com/sahib/brig/fuse"
	"github.com/sahib/brig/util"
	"github.com/sahib/brig/util/server"
	log "github.com/sirupsen/logrus"
)

// Server is the local api server used by the command client.
type Server struct {
	daemonURL  string
	baseServer *server.Server
	base       *base
}

// Serve blocks until a quit command was send.
func (sv *Server) Serve() error {
	log.Infof("Serving local requests from now on.")
	return sv.baseServer.Serve()
}

// Close will clean up the listener resources.
func (sv *Server) Close() error {
	sv.baseServer.Quit()
	return sv.baseServer.Close()
}

func listenerFromServerURL(s string) (net.Listener, error) {
	scheme, addr, err := util.URLToSchemeAndAddr(s)
	if err != nil {
		return nil, err
	}

	return net.Listen(scheme, addr)
}

func applyFstabInitially(base *base) error {
	return fuse.FsTabApply(base.repo.Config.Section("mounts"), base.mounts)
}

// RepoPath returns the repo path we're operating on
func (sv *Server) RepoPath() string {
	return sv.base.basePath
}

func (sv *Server) DaemonURL() string {
	return sv.daemonURL
}

// BootServer will boot up the local server.
func BootServer(basePath string, serverURL string) (*Server, error) {
	defer func() {
		// If anything in the daemon goes fatally wrong and it blows up, we
		// want to log the panic at least. Otherwise we'll have a hard time
		// debugging why the daemon suddenly quit.
		if err := recover(); err != nil {
			log.Errorf("brig panicked with message: %v", err)
			log.Errorf("stack trace: %s", debug.Stack())
			panic(err)
		}
	}()

	log.Infof("starting daemon for %s at %s", basePath, serverURL)
	if err := increaseMaxOpenFds(); err != nil {
		log.Warningf("failed to increase number of open fds")
	}

	ctx := context.Background()
	quitCh := make(chan struct{})
	base := newBase(ctx, basePath, quitCh)
	lst, err := listenerFromServerURL(serverURL)
	if err != nil {
		return nil, err
	}

	baseServer, err := server.NewServer(ctx, lst, base)
	if err != nil {
		return nil, err
	}

	go func() {
		// Wait for a quit signal.
		<-quitCh
		baseServer.Quit()
		if err := baseServer.Close(); err != nil {
			log.Warnf("failed to close local server listener: %v", err)
		}
	}()

	if err := base.loadAll(); err != nil {
		return nil, err
	}

	if err := applyFstabInitially(base); err != nil {
		log.Warnf("could not mount fstab mounts: %v", err)
	}

	return &Server{
		daemonURL:  serverURL,
		baseServer: baseServer,
		base:       base,
	}, nil
}
