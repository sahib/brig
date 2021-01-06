package server

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"path/filepath"
	"runtime/debug"

	"github.com/sahib/brig/defaults"
	"github.com/sahib/brig/fuse"
	"github.com/sahib/brig/repo"
	"github.com/sahib/brig/util/pwutil"
	"github.com/sahib/brig/util/server"
	log "github.com/sirupsen/logrus"
)

// Server is the local api server used by the command client.
type Server struct {
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

func readPasswordFromHelper(basePath string, passwordFn func() (string, error)) (string, error) {
	configPath := filepath.Join(basePath, "config.yml")
	cfg, err := defaults.OpenMigratedConfig(configPath)
	if err != nil {
		return "", err
	}

	passwordCmd := cfg.String("repo.password_command")
	if passwordCmd == "" {
		log.Infof("reading password via client logic")
		return passwordFn()
	}

	log.Infof("password was read from the password helper")
	return pwutil.ReadPasswordFromHelper(basePath, passwordCmd)
}

func listenerFromServerURL(s string) (net.Listener, error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}

	// NOTE: slightly confusing quirk: for unix sockets the
	switch u.Scheme {
	case "tcp":
		return net.Listen(u.Scheme, u.Host)
	case "unix":
		return net.Listen(u.Scheme, u.Path)
	default:
		return nil, fmt.Errorf("unsupported protocol: %v", u.Scheme)
	}
}

func applyFstabInitially(base *base) error {
	return fuse.FsTabApply(base.repo.Config.Section("mounts"), base.mounts)
}

// BootServer will boot up the local server.
func BootServer(
	basePath string,
	serverURL string,
	passwordFn func() (string, error),
	logToStdout bool,
) (*Server, error) {
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

	password, err := readPasswordFromHelper(basePath, passwordFn)
	if err != nil {
		return nil, err
	}

	if err := repo.CheckPassword(basePath, password); err != nil {
		return nil, err
	}

	log.Infof("password is valid")
	if err := increaseMaxOpenFds(); err != nil {
		log.Warningf("failed to increase number of open fds")
	}

	ctx := context.Background()
	quitCh := make(chan struct{})
	base := newBase(
		ctx,
		basePath,
		password,
		quitCh,
	)

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
		baseServer: baseServer,
		base:       base,
	}, nil
}
