package transfer

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/transfer/wire"
	"github.com/disorganizer/brig/util"
	"github.com/disorganizer/brig/util/ipfsutil"
	"github.com/gogo/protobuf/proto"
)

// APIClient is a high-level client that talks to
// other peers in brig's network. Calls on it will
// directly talk to the other side and convert the
// response back to native go structures.
type APIClient struct {
	cnv   Conversation
	node  *ipfsutil.Node
	idcnt int64
}

// newAPIClient returns a new APIClient on top of a conversation
func newAPIClient(cnv Conversation, node *ipfsutil.Node) (*APIClient, error) {
	return &APIClient{
		cnv:  cnv,
		node: node,
	}, nil
}

// send is the synchronous variant of SendAsync
func (acl *APIClient) send(req *wire.Request) (resp *wire.Response, err error) {
	// `0` is reserved for broadcast counters,
	// increment first therefore.
	acl.idcnt++
	req.ID = proto.Int64(acl.idcnt)

	done := make(chan util.Empty)
	err = acl.cnv.SendAsync(req, func(respIn *wire.Response) {
		resp = respIn
		done <- util.Empty{}
	})

	// TODO: Make that configurable?
	timer := time.NewTimer(10 * time.Second)

	// Wait until we got a response from SendAsync or until
	// we time out.
	select {
	case <-done:
		break
	case stamp := <-timer.C:
		log.Warningf("APIClient operation timed out at %v", stamp)
		return nil, util.ErrTimeout
	}

	return
}

// QueryStoreVersion returns the storage version of the remote store.
func (acl *APIClient) QueryStoreVersion() (int32, error) {
	req := &wire.Request{
		ReqType: wire.RequestType_STORE_VERSION.Enum(),
	}

	resp, err := acl.send(req)
	if err != nil {
		return -1, err
	}

	return resp.GetStoreVersionResp().GetVersion(), nil
}
