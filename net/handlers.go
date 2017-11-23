package net

import (
	"bytes"
	"fmt"

	"github.com/disorganizer/brig/net/capnp"
)

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
	fmt.Println("PONG")
	return call.Results.SetReply("ALIVE")
}

func (hdl *handler) Version(call capnp.API_version) error {
	call.Results.SetVersion(1)
	return nil
}
