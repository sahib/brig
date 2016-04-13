package transfer

import (
	"bytes"

	"github.com/disorganizer/brig/transfer/wire"
)

// NOTE: New handlers need to be registered in handlerMap in NewConnector.

func (sv *Connector) handleFetch(req *wire.Request) (*wire.Response, error) {
	buf := &bytes.Buffer{}
	if err := sv.rp.OwnStore.Export(buf); err != nil {
		return nil, err
	}

	return &wire.Response{Data: buf.Bytes()}, nil
}

func (sv *Connector) handleUpdateFile(req *wire.Request) (*wire.Response, error) {
	return nil, nil
}
