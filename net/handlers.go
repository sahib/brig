package net

import (
	"bytes"
	"context"
	"net"

	log "github.com/Sirupsen/logrus"
	"zombiezen.com/go/capnproto2/rpc"

	"github.com/disorganizer/brig/backend"
	"github.com/disorganizer/brig/net/capnp"
	"github.com/disorganizer/brig/repo"
)

type handler struct {
	bk backend.Backend
	rp *repo.Repository
}

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

// Quit is being called by the base server implementation
func (hdl *handler) Quit() error {
	return nil
}

func (hdl *handler) GetStore(call capnp.Sync_getStore) error {
	fs, err := hdl.rp.OwnFS(hdl.bk)
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	if err := fs.Export(buf); err != nil {
		return err
	}

	return call.Results.SetData(buf.Bytes())
}

func (hdl *handler) Ping(call capnp.Meta_ping) error {
	return call.Results.SetReply("ALIVE")
}

func (hdl *handler) PubKey(call capnp.Meta_pubKey) error {
	data, err := hdl.rp.Keyring().OwnPubKey()
	if err != nil {
		return err
	}

	return call.Results.SetKey(data)
}

func (hdl *handler) Version(call capnp.API_version) error {
	call.Results.SetVersion(1)
	return nil
}
