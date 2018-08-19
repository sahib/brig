package server

import (
	"context"
	"fmt"
	"net"

	log "github.com/Sirupsen/logrus"
	"github.com/sahib/brig/repo"
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

func BootServer(basePath, password, bindHost string, port int) (*Server, error) {
	addr := fmt.Sprintf("%s:%d", bindHost, port)
	log.Infof("Starting daemon from %s on port %s", basePath, addr)

	if err := repo.CheckPassword(basePath, password); err != nil {
		return nil, err
	}

	log.Infof("Password seems to be valid...")

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

	// TODO: Go online automatically
	// TODO: Mount fstab entries here automatically.

	return &Server{
		baseServer: baseServer,
		base:       base,
	}, nil
}
