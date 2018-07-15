// Package nilrouting implements a routing client that does nothing.
package nilrouting

import (
	"context"
	"errors"

	routing "gx/ipfs/QmPpdpS9fknTBM3qHDcpayU6nYPZQeVjia2fbNrD8YWDe6/go-libp2p-routing"
	ropts "gx/ipfs/QmPpdpS9fknTBM3qHDcpayU6nYPZQeVjia2fbNrD8YWDe6/go-libp2p-routing/options"
	record "gx/ipfs/QmVsp2KdPYE6M8ryzCk5KHLo3zprcY5hBDaYx6uPCFUdxA/go-libp2p-record"
	pstore "gx/ipfs/QmZR2XWVVBCtbgBWnQhWk2xcQfaR3W8faQPriAiaaj7rsr/go-libp2p-peerstore"
	cid "gx/ipfs/QmapdYm1b22Frv3k17fqrBYTFRxwiaVJkB299Mfn33edeB/go-cid"
	p2phost "gx/ipfs/Qmb8T6YBBsjYsVGfrihQLfCJveczZnneSBqBKkYEBWDjge/go-libp2p-host"
	peer "gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
	ds "gx/ipfs/QmeiCcJfDW1GJnWUArudsv5rQsihpi4oyddPhdqo3CfX6i/go-datastore"
)

type nilclient struct {
}

func (c *nilclient) PutValue(_ context.Context, _ string, _ []byte, _ ...ropts.Option) error {
	return nil
}

func (c *nilclient) GetValue(_ context.Context, _ string, _ ...ropts.Option) ([]byte, error) {
	return nil, errors.New("tried GetValue from nil routing")
}

func (c *nilclient) FindPeer(_ context.Context, _ peer.ID) (pstore.PeerInfo, error) {
	return pstore.PeerInfo{}, nil
}

func (c *nilclient) FindProvidersAsync(_ context.Context, _ *cid.Cid, _ int) <-chan pstore.PeerInfo {
	out := make(chan pstore.PeerInfo)
	defer close(out)
	return out
}

func (c *nilclient) Provide(_ context.Context, _ *cid.Cid, _ bool) error {
	return nil
}

func (c *nilclient) Bootstrap(_ context.Context) error {
	return nil
}

// ConstructNilRouting creates an IpfsRouting client which does nothing.
func ConstructNilRouting(_ context.Context, _ p2phost.Host, _ ds.Batching, _ record.Validator) (routing.IpfsRouting, error) {
	return &nilclient{}, nil
}

//  ensure nilclient satisfies interface
var _ routing.IpfsRouting = &nilclient{}
