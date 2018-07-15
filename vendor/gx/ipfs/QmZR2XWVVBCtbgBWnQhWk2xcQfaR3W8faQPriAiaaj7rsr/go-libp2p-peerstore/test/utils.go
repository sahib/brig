package testutil

import (
	"io"
	"math/rand"
	"time"

	mh "gx/ipfs/QmPnFwZ2JXKnXgMw8CdBPxn7FWh6LLdjUjxV1fKHuJnkr8/go-multihash"
	peer "gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
	ci "gx/ipfs/Qme1knMqwt1hKZbc1BmQFmnm9f36nyQGwXxPGVpVJ9rMK5/go-libp2p-crypto"
)

func timeSeededRand() io.Reader {
	return rand.New(rand.NewSource(time.Now().UnixNano()))
}

func RandPeerID() (peer.ID, error) {
	buf := make([]byte, 16)
	if _, err := io.ReadFull(timeSeededRand(), buf); err != nil {
		return "", err
	}
	h, err := mh.Sum(buf, mh.SHA2_256, -1)
	if err != nil {
		return "", err
	}

	return peer.ID(h), nil
}

func RandTestKeyPair(bits int) (ci.PrivKey, ci.PubKey, error) {
	return ci.GenerateKeyPairWithReader(ci.RSA, bits, timeSeededRand())
}

func SeededTestKeyPair(seed int64) (ci.PrivKey, ci.PubKey, error) {
	return ci.GenerateKeyPairWithReader(ci.RSA, 512, rand.New(rand.NewSource(seed)))
}
