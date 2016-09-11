package id

import (
	"bytes"
	"fmt"

	multihash "github.com/jbenet/go-multihash"
)

// TODO: Proposal:
// Parse the domain and register a ipfs block for that too.
// auto discovery can limit the search to those then.
// Also: register a common "brig" block, making discovery of all nodes
// possible.

type Peer interface {
	ID() ID
	Hash() string
}

func MarshalPeer(p Peer) ([]byte, error) {
	return p.Hash() + "=" + p.ID()
}

func UnmarshalPeer(data []byte) (Peer, error) {
	split := bytes.SplitN(data, []byte("="), 1)

	if len(split) < 2 {
		return nil, fmt.Errorf("Marshalled peer has no `=` in it")
	}

	id, err := Cast(split[1])
	if err != nil {
		return nil, err
	}

	mh, err := multihash.FromB58String(split[0])
	if err != nil {
		return nil, err
	}

	return NewPeer(id, mh.B58String())
}

type ipfsPeer struct {
	id   ID
	hash string
}

func NewPeer(id ID, hash string) Peer {
	return &ipfsPeer{id, hash}
}

func (ip *ipfsPeer) ID() ID {
	return ip.id
}

func (ip *ipfsPeer) Hash() string {
	return ip.hash
}
