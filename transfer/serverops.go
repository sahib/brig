package transfer

import (
	"bytes"
	"fmt"

	"github.com/disorganizer/brig/transfer/wire"
	"github.com/gogo/protobuf/proto"
)

// NOTE: New handlers need to be registered in handlerMap in NewConnector.
//       (The map is not here, because we use method values)

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

func (sv *Connector) handleStoreVersion(req *wire.Request) (*wire.Response, error) {
	fmt.Println("handle store version")
	return &wire.Response{
		StoreVersionResp: &wire.StoreVersionResponse{
			// TODO: Use actual version when ready
			Version: proto.Int32(42),
		},
	}, nil
}
