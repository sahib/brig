package transfer

import (
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
	client := &APIClient{
		cnv:  cnv,
		node: node,
	}

	return client, nil
}

func (acl *APIClient) Close() error {
	return acl.cnv.Close()
}
