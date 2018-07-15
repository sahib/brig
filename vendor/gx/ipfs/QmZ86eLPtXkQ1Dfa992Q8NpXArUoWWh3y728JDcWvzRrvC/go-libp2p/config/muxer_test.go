package config

import (
	"testing"

	mux "gx/ipfs/QmY9JXR3FupnYAYJWK9aMr9bCpqWKcToQ1tz8DVGTrHpHw/go-stream-muxer"
	yamux "gx/ipfs/QmcsgrV3nCAKjiHKZhKVXWc4oY3WBECJCqahXEMpHeMrev/go-smux-yamux"
	peer "gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
)

func TestMuxerSimple(t *testing.T) {
	// single
	_, err := MuxerConstructor(func(_ peer.ID) mux.Transport { return nil })
	if err != nil {
		t.Fatal(err)
	}
}

func TestMuxerByValue(t *testing.T) {
	_, err := MuxerConstructor(yamux.DefaultTransport)
	if err != nil {
		t.Fatal(err)
	}
}
func TestMuxerDuplicate(t *testing.T) {
	_, err := MuxerConstructor(func(_ peer.ID, _ peer.ID) mux.Transport { return nil })
	if err != nil {
		t.Fatal(err)
	}
}

func TestMuxerError(t *testing.T) {
	_, err := MuxerConstructor(func() (mux.Transport, error) { return nil, nil })
	if err != nil {
		t.Fatal(err)
	}
}

func TestMuxerBadTypes(t *testing.T) {
	for i, f := range []interface{}{
		func() error { return nil },
		func() string { return "" },
		func() {},
		func(string) mux.Transport { return nil },
		func(string) (mux.Transport, error) { return nil, nil },
		nil,
		"testing",
	} {

		if _, err := MuxerConstructor(f); err == nil {
			t.Fatalf("constructor %d with type %T should have failed", i, f)
		}
	}
}
