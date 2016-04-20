package transfer

import (
	"github.com/disorganizer/brig/store"
	"github.com/disorganizer/brig/transfer/wire"
)

type Broadcaster struct {
	cnc *Connector
}

func (bc *Broadcaster) FileUpdate(file *store.File) error {
	// TODO: Use `file` somehow (also: Document)
	req := &wire.Request{
		ReqType: wire.RequestType_UPDATE_FILE.Enum(),
	}
	return bc.cnc.broadcast(req)
}
