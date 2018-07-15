// Package exchange defines the IPFS exchange interface
package exchange

import (
	"context"
	"io"

	blocks "gx/ipfs/QmTRCUvZLiir12Qr6MV3HKfKMHX8Nf1Vddn6t2g5nsQSb9/go-block-format"
	cid "gx/ipfs/QmapdYm1b22Frv3k17fqrBYTFRxwiaVJkB299Mfn33edeB/go-cid"
)

// Interface defines the functionality of the IPFS block exchange protocol.
type Interface interface { // type Exchanger interface
	Fetcher

	// TODO Should callers be concerned with whether the block was made
	// available on the network?
	HasBlock(blocks.Block) error

	IsOnline() bool

	io.Closer
}

// Fetcher is an object that can be used to retrieve blocks
type Fetcher interface {
	// GetBlock returns the block associated with a given key.
	GetBlock(context.Context, *cid.Cid) (blocks.Block, error)
	GetBlocks(context.Context, []*cid.Cid) (<-chan blocks.Block, error)
}

// SessionExchange is an exchange.Interface which supports
// sessions.
type SessionExchange interface {
	Interface
	NewSession(context.Context) Interface
}
