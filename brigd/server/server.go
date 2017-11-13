package server

import (
	"context"
	"fmt"
	"net"

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
	return sv.baseServer.Serve()
}

func (sv *Server) Close() error {
	return sv.baseServer.Close()
}

func BootServer(basePath, password string, port int) (*Server, error) {
	if err := repo.CheckPassword(basePath, password); err != nil {
		return nil, err
	}

	ctx := context.Background()
	base, err := newBase(basePath, password, ctx)
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

	return &Server{
		baseServer: baseServer,
		base:       base,
	}, nil
}
