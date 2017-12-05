package server

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/repo"
	"github.com/disorganizer/brig/util/server"
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

func setLogPath(path string) error {
	switch path {
	case "stdout":
		log.SetOutput(os.Stdout)
	case "stderr":
		log.SetOutput(os.Stderr)
	default:
		fd, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}

		log.SetOutput(fd)
	}

	return nil
}

func BootServer(basePath, password, logPath string, port int) (*Server, error) {
	if logPath == "" {
		logPath = filepath.Join(basePath, "logs", "main.log")
		if err := os.MkdirAll(filepath.Dir(logPath), 0700); err != nil {
			return nil, err
		}
	}

	if err := setLogPath(logPath); err != nil {
		return nil, err
	}

	log.Infof("Starting server from %s at port :%d", basePath, port)

	if err := repo.CheckPassword(basePath, password); err != nil {
		return nil, err
	}

	log.Infof("Password seems to be valid...")

	ctx := context.Background()
	quitCh := make(chan struct{})
	base, err := newBase(basePath, password, ctx, quitCh)
	if err != nil {
		return nil, err
	}

	addr := fmt.Sprintf("localhost:%d", port)
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

	return &Server{
		baseServer: baseServer,
		base:       base,
	}, nil
}
