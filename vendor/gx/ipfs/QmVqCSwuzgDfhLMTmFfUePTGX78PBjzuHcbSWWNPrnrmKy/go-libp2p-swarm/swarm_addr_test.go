package swarm_test

import (
	"context"
	"testing"

	ma "gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
	pstore "gx/ipfs/QmZR2XWVVBCtbgBWnQhWk2xcQfaR3W8faQPriAiaaj7rsr/go-libp2p-peerstore"
	testutil "gx/ipfs/QmcW4FGAt24fdK1jBgWQn3yP4R9ZLyWQqjozv9QK7epRhL/go-testutil"
)

func TestDialBadAddrs(t *testing.T) {

	m := func(s string) ma.Multiaddr {
		maddr, err := ma.NewMultiaddr(s)
		if err != nil {
			t.Fatal(err)
		}
		return maddr
	}

	ctx := context.Background()
	s := makeSwarms(ctx, t, 1)[0]

	test := func(a ma.Multiaddr) {
		p := testutil.RandPeerIDFatal(t)
		s.Peerstore().AddAddr(p, a, pstore.PermanentAddrTTL)
		if _, err := s.DialPeer(ctx, p); err == nil {
			t.Errorf("swarm should not dial: %s", p)
		}
	}

	test(m("/ip6/fe80::1"))                // link local
	test(m("/ip6/fe80::100"))              // link local
	test(m("/ip4/127.0.0.1/udp/1234/utp")) // utp
}
