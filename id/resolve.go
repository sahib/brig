package id

// TODO: Proposal:
// Parse the domain and register a ipfs block for that too.
// auto discovery can limit the search to those then.
// Also: register a common "brig" block, making discovery of all nodes
// possible.

type Peer interface {
	ID() ID
	Hash() string
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
