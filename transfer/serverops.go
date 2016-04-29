package transfer

import (
	"github.com/disorganizer/brig/transfer/wire"
	"github.com/gogo/protobuf/proto"
)

func (sv *Connector) handleFetch(req *wire.Request) (*wire.Response, error) {
	protoStore, err := sv.rp.OwnStore.Export()
	if err != nil {
		return nil, err
	}

	return &wire.Response{
		FetchResp: &wire.FetchResponse{
			Store: protoStore,
		},
	}, nil
}

func (sv *Connector) handleUpdateFile(req *wire.Request) (*wire.Response, error) {
	return nil, nil
}

func (sv *Connector) handleStoreVersion(req *wire.Request) (*wire.Response, error) {
	return &wire.Response{
		StoreVersionResp: &wire.StoreVersionResponse{
			// TODO: Use actual version when ready
			Version: proto.Int32(42),
		},
	}, nil
}
