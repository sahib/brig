package server

import (
	"context"

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

	base, err := newBase(basePath, password)
	if err != nil {
		return nil, err
	}

	baseServer, err := server.NewServer(port, base, ctx)
	if err != nil {
		return nil, err
	}

	return &Server{
		baseServer: baseServer,
		base:       base,
	}, nil
}
