package net

import (
	"context"
	"net"

	"github.com/disorganizer/brig/net/capnp"
)

type handler struct{}

func (hdl *handler) Handle(ctx context.Context, conn net.Conn) {
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
