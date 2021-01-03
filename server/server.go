package server

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log/syslog"
	"net"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/sahib/brig/defaults"
	"github.com/sahib/brig/fuse"
	"github.com/sahib/brig/repo"
	formatter "github.com/sahib/brig/util/log"
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

func switchToSyslog() {
	wSyslog, err := syslog.New(syslog.LOG_NOTICE, "brig")
	if err != nil {
		log.Warningf("failed to open connection to syslog for brig: %v", err)
		logFd, err := ioutil.TempFile("", "brig-*.log")
		if err != nil {
			log.Warningf("")
		} else {
			log.Warningf("Will log to %s from now on.", logFd.Name())
			log.SetOutput(logFd)
		}

		return
	}

	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&formatter.FancyLogFormatter{
		UseColors: false,
	})

	log.SetOutput(
		io.MultiWriter(
			formatter.NewSyslogWrapper(wSyslog),
			os.Stdout,
		),
	)
}

func applyFstabInitially(base *base) error {
	return fuse.FsTabApply(base.repo.Config.Section("mounts"), base.mounts)
}

// BootServer will boot up the local server.
// `basePath` is the path to the repository.
// `passwordFn` is a function that will deliver a password when
// no password was configured.
// `bindHost` is the host to bind too.
// `port` is the port to listen for requests.
// `logToStdout` should be true when logging to stdout.
func BootServer(
	basePath string,
	passwordFn func() (string, error),
	bindHost string,
	port int,
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

	if logToStdout {
		// Be sure it's really set to stdout.
		log.SetOutput(os.Stdout)
	} else {
		switchToSyslog()
	}

	addr := fmt.Sprintf("%s:%d", bindHost, port)
	log.Infof("starting daemon for %s on port %s", basePath, addr)

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
		int64(port),
		basePath,
		password,
		bindHost,
		quitCh,
		logToStdout,
	)

	lst, err := net.Listen("tcp", addr)
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
