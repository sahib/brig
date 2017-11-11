package net

import (
	"context"

	"github.com/disorganizer/brig/util/server"
)

type Server struct {
	baseServer *server.Server
	hdl        *handler
}

func (sv *Server) Serve() error {
	return sv.baseServer.Serve()
}

func (sv *Server) Close() error {
	return sv.baseServer.Close()
}

func NewServer(port int) (*Server, error) {
	hdl := &handler{}
	ctx := context.Background()

	baseServer, err := server.NewServer(port, hdl, ctx)
	if err != nil {
		return nil, err
	}

	return &Server{
		baseServer: baseServer,
		hdl:        hdl,
	}, nil
}
