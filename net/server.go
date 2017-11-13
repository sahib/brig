package net

import (
	"context"

	"github.com/disorganizer/brig/net/backend"
	"github.com/disorganizer/brig/net/peer"
	"github.com/disorganizer/brig/util/server"
)

type Server struct {
	bk         backend.Backend
	baseServer *server.Server
	hdl        *handler
}

func (sv *Server) Serve() error {
	return sv.baseServer.Serve()
}

func (sv *Server) Close() error {
	return sv.baseServer.Close()
}

func NewServer(bk backend.Backend) (*Server, error) {
	hdl := &handler{}
	ctx := context.Background()

	lst, err := bk.Listen("brig-caprpc")
	if err != nil {
		return nil, err
	}

	baseServer, err := server.NewServer(lst, hdl, ctx)
	if err != nil {
		return nil, err
	}

	return &Server{
		baseServer: baseServer,
		hdl:        hdl,
	}, nil
}

func (sv *Server) Locate(who peer.Name) ([]peer.Info, error) {
	return sv.bk.ResolveName(who)
}

func (sv *Server) Identity() (peer.Info, error) {
	return sv.bk.Identity()
}
