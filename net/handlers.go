package net

import (
	"context"
	"net"

	"zombiezen.com/go/capnproto2/rpc"

	"github.com/disorganizer/brig/net/capnp"
)

type handler struct{}

func (hdl *handler) Handle(ctx context.Context, conn net.Conn) {
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

func (hdl *handler) Quit() error {
	return nil
}

func (hdl *handler) GetStore(call capnp.Sync_getStore) error {
	// TODO: Implement.
	return nil
}

func (hdl *handler) Ping(call capnp.Meta_ping) error {
	return call.Results.SetReply("ALIVE")
}

func (hdl *handler) PubKey(call capnp.Meta_pubKey) error {
	// TODO: Implement.
	return nil
}

func (hdl *handler) Version(call capnp.API_version) error {
	call.Results.SetVersion(1)
	return nil
}
