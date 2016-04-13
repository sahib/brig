package transfer

import (
	"time"

	"github.com/disorganizer/brig/util/ipfsutil"
)

// APIClient is a high-level client that talks to
// other peers in brig's network. Calls on it will
// directly
type APIClient struct {
	cnv    Conversation
	node   *ipfsutil.Node
	pinger *ipfsutil.Pinger
}

func newAPIClient(cnv Conversation, node *ipfsutil.Node) (*APIClient, error) {
	peer := cnv.Peer()

	pinger, err := node.Ping(peer.Hash())
	if err != nil {
		return nil, err
	}

	client := &APIClient{
		cnv:    cnv,
		node:   node,
		pinger: pinger,
	}

	return client, nil
}

func (acl *APIClient) LastSeen() time.Time {
	return acl.pinger.LastSeen()
}

func (acl *APIClient) Close() error {
	acl.pinger.Close()
	return acl.cnv.Close()
}
