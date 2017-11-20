package net

import (
	"context"

	"github.com/disorganizer/brig/backend"
	"github.com/disorganizer/brig/net/peer"
	"github.com/disorganizer/brig/repo"
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

func NewServer(rp *repo.Repository, bk backend.Backend) (*Server, error) {
	hdl := &handler{
		rp: rp,
		bk: bk,
	}

	lst, err := bk.Listen("brig/caprpc")
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	baseServer, err := server.NewServer(lst, hdl, ctx)
	if err != nil {
		return nil, err
	}

	return &Server{
		baseServer: baseServer,
		bk:         bk,
		hdl:        hdl,
	}, nil
}

func (sv *Server) Locate(who peer.Name) ([]peer.Info, error) {
	return sv.bk.ResolveName(who)
}

func (sv *Server) Identity() (peer.Info, error) {
	return sv.bk.Identity()
}
