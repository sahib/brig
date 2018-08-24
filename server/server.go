package server

import (
	"context"
	"fmt"
	"io/ioutil"
	"log/syslog"
	"net"
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/sahib/brig/defaults"
	"github.com/sahib/brig/fuse"
	"github.com/sahib/brig/repo"
	formatter "github.com/sahib/brig/util/log"
	"github.com/sahib/brig/util/pwutil"
	"github.com/sahib/brig/util/server"
)

const (
	MaxConnections = 10
)

//////////////////////////////

type Server struct {
	baseServer *server.Server
	base       *base
}

func (sv *Server) Serve() error {
	log.Infof("Serving local requests from now on.")
	return sv.baseServer.Serve()
}

func (sv *Server) Close() error {
	return sv.baseServer.Close()
}

func readPasswordFromHelper(basePath string) (string, error) {
	configPath := filepath.Join(basePath, "config.yml")
	cfg, err := defaults.OpenMigratedConfig(configPath)
	if err != nil {
		return "", err
	}

	passwordCmd := cfg.String("repo.password_command")
	if passwordCmd == "" {
		return "", fmt.Errorf("no password helper set")
	}

	return pwutil.ReadPasswordFromHelper(basePath, passwordCmd)
}

func switchToSyslog() {
	wSyslog, err := syslog.New(syslog.LOG_NOTICE, "brig")
	if err != nil {
		log.Warningf("Failed to open connection to syslog for brig: %v", err)
	}

	log.SetFormatter(&formatter.ColorfulLogFormatter{})
	log.SetLevel(log.DebugLevel)
	log.SetOutput(formatter.NewSyslogWrapper(wSyslog))
}

func updateRegistry(basePath string, port int) error {
	data, err := ioutil.ReadFile(filepath.Join(basePath, "REPO_ID"))
	if err != nil {
		return err
	}

	uuid := string(data)

	// TODO: Move repo.OpenRegisry to util somewhere.
	// It is also used in the client.
	registry, err := repo.OpenRegistry()
	if err != nil {
		return err
	}

	entry, err := registry.Entry(uuid)
	if err != nil {
		return err
	}

	entry.Port = int64(port)
	entry.Path = basePath
	return registry.Update(uuid, entry)
}

func applyFstabInitially(base *base) error {
	rp, err := base.Repo()
	if err != nil {
		return err
	}

	mounts, err := base.Mounts()
	if err != nil {
		return err
	}

	return fuse.FsTabApply(rp.Config.Section("mounts"), mounts)
}

func startNetLayer(base *base) error {
	_, err := base.PeerServer()
	return err
}

func BootServer(basePath string, passwordFn func() (string, error), bindHost string, port int, logToStdout bool) (*Server, error) {
	if !logToStdout {
		switchToSyslog()
	} else {
		log.SetOutput(os.Stdout)
	}

	addr := fmt.Sprintf("%s:%d", bindHost, port)
	log.Infof("Starting daemon from %s on port %s", basePath, addr)

	password, err := readPasswordFromHelper(basePath)
	if err != nil {
		log.Infof("Failed to read password from helper: %s", err)
		log.Infof("Attempting to read it from stdin")

		password, err = passwordFn()
		if err != nil {
			return nil, err
		}
	} else {
		log.Infof("Password is coming from the configured password helper")
	}

	if err := repo.CheckPassword(basePath, password); err != nil {
		return nil, err
	}

	log.Infof("Password seems to be valid...")

	if err := updateRegistry(basePath, port); err != nil {
		log.Warningf("could not update global registry: %v", err)
	}

	if err := increaseMaxOpenFds(); err != nil {
		log.Warningf("Failed to incrase number of open fds")
	}

	ctx := context.Background()
	quitCh := make(chan struct{})
	base, err := newBase(basePath, password, bindHost, ctx, quitCh)
	if err != nil {
		return nil, err
	}

	lst, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	baseServer, err := server.NewServer(lst, base, ctx)
	if err != nil {
		return nil, err
	}

	go func() {
		<-quitCh
		baseServer.Quit()
		if err := baseServer.Close(); err != nil {
			log.Warnf("Failed to close local server listener: %v", err)
		}
	}()

	// Do the rest of the init in the background.
	// This will curently log warnings for a not yet initialized repo.
	go func() {
		if err := startNetLayer(base); err != nil {
			log.Warnf("could not start the net layer yet: %v", err)
		}

		if err := applyFstabInitially(base); err != nil {
			log.Warnf("could not mount fstab mounts: %v", err)
		}
	}()

	return &Server{
		baseServer: baseServer,
		base:       base,
	}, nil
}
