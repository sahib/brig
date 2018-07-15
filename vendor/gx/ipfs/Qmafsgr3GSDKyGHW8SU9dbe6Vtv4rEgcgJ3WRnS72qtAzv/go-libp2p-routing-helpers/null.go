package routinghelpers

import (
	"context"

	routing "gx/ipfs/QmPpdpS9fknTBM3qHDcpayU6nYPZQeVjia2fbNrD8YWDe6/go-libp2p-routing"
	ropts "gx/ipfs/QmPpdpS9fknTBM3qHDcpayU6nYPZQeVjia2fbNrD8YWDe6/go-libp2p-routing/options"

	pstore "gx/ipfs/QmZR2XWVVBCtbgBWnQhWk2xcQfaR3W8faQPriAiaaj7rsr/go-libp2p-peerstore"
	cid "gx/ipfs/QmapdYm1b22Frv3k17fqrBYTFRxwiaVJkB299Mfn33edeB/go-cid"
	peer "gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
)

// Null is a router that doesn't do anything.
type Null struct{}

// PutValue always returns ErrNotSupported
func (nr Null) PutValue(context.Context, string, []byte, ...ropts.Option) error {
	return routing.ErrNotSupported
}

// GetValue always returns ErrNotFound
func (nr Null) GetValue(context.Context, string, ...ropts.Option) ([]byte, error) {
	return nil, routing.ErrNotFound
}

// Provide always returns ErrNotSupported
func (nr Null) Provide(context.Context, *cid.Cid, bool) error {
	return routing.ErrNotSupported
}

// FindProvidersAsync always returns a closed channel
func (nr Null) FindProvidersAsync(context.Context, *cid.Cid, int) <-chan pstore.PeerInfo {
	ch := make(chan pstore.PeerInfo)
	close(ch)
	return ch
}

// FindPeer always returns ErrNotFound
func (nr Null) FindPeer(context.Context, peer.ID) (pstore.PeerInfo, error) {
	return pstore.PeerInfo{}, routing.ErrNotFound
}

// Bootstrap always succeeds instantly
func (nr Null) Bootstrap(context.Context) error {
	return nil
}

var _ routing.IpfsRouting = Null{}
