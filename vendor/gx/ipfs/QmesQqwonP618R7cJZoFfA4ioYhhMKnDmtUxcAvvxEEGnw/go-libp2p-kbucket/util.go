package kbucket

import (
	"bytes"
	"crypto/sha256"
	"errors"

	u "gx/ipfs/QmPdKqUcHGFdeSpvjVoaTRPPstGif9GBZb5Q56RVw9o69A/go-ipfs-util"
	peer "gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
	ks "gx/ipfs/QmesQqwonP618R7cJZoFfA4ioYhhMKnDmtUxcAvvxEEGnw/go-libp2p-kbucket/keyspace"
)

// Returned if a routing table query returns no results. This is NOT expected
// behaviour
var ErrLookupFailure = errors.New("failed to find any peer in table")

// ID for IpfsDHT is in the XORKeySpace
//
// The type dht.ID signifies that its contents have been hashed from either a
// peer.ID or a util.Key. This unifies the keyspace
type ID []byte

func (id ID) equal(other ID) bool {
	return bytes.Equal(id, other)
}

func (id ID) less(other ID) bool {
	a := ks.Key{Space: ks.XORKeySpace, Bytes: id}
	b := ks.Key{Space: ks.XORKeySpace, Bytes: other}
	return a.Less(b)
}

func xor(a, b ID) ID {
	return ID(u.XOR(a, b))
}

func commonPrefixLen(a, b ID) int {
	return ks.ZeroPrefixLen(u.XOR(a, b))
}

// ConvertPeerID creates a DHT ID by hashing a Peer ID (Multihash)
func ConvertPeerID(id peer.ID) ID {
	hash := sha256.Sum256([]byte(id))
	return hash[:]
}

// ConvertKey creates a DHT ID by hashing a local key (String)
func ConvertKey(id string) ID {
	hash := sha256.Sum256([]byte(id))
	return hash[:]
}

// Closer returns true if a is closer to key than b is
func Closer(a, b peer.ID, key string) bool {
	aid := ConvertPeerID(a)
	bid := ConvertPeerID(b)
	tgt := ConvertKey(key)
	adist := xor(aid, tgt)
	bdist := xor(bid, tgt)

	return adist.less(bdist)
}
