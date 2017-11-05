package net

import (
	"context"
	"net"

	"zombiezen.com/go/capnproto2/rpc"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/net/capnp"
	"github.com/disorganizer/brig/util/server"
)

type Server struct {
	baseServer *server.Server
	hdl        *handler
}

func (hdl *handler) handle(ctx context.Context, conn net.Conn) {
	transport := rpc.StreamTransport(conn)
	srv := capnp.API_ServerToClient(hdl)
	rpcConn := rpc.NewConn(transport, rpc.MainInterface(srv.Client))

	if err := rpcConn.Wait(); err != nil {
		log.Warnf("Serving rpc failed: %v", err)
	}

	if err := rpcConn.Close(); err != nil {
		log.Warnf("Failed to close rpc conn: %v", err)
	}
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
