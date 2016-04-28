package transfer

import (
	"github.com/disorganizer/brig/store"
	"github.com/disorganizer/brig/transfer/wire"
)

// Broadcaster offers the API to the individual broadcast messages.
type Broadcaster struct {
	cnc *Connector
}

// FileUpdate notifies all connected tpeers that `file` changed.
// TODO: Actually pass checkpoint?
func (bc *Broadcaster) FileUpdate(file *store.File) error {
	// TODO: Use `file` somehow (also: Document)
	req := &wire.Request{
		ReqType: wire.RequestType_UPDATE_FILE.Enum(),
	}
	return bc.cnc.broadcast(req)
}
