package routinghelpers

import (
	"context"
	"testing"

	routing "gx/ipfs/QmPpdpS9fknTBM3qHDcpayU6nYPZQeVjia2fbNrD8YWDe6/go-libp2p-routing"
	peert "gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer/test"
)

func TestGetPublicKey(t *testing.T) {
	d := Parallel{
		Parallel{
			&Compose{
				ValueStore: &LimitedValueStore{
					ValueStore: new(dummyValueStore),
					Namespaces: []string{"other"},
				},
			},
		},
		Tiered{
			&Compose{
				ValueStore: &LimitedValueStore{
					ValueStore: new(dummyValueStore),
					Namespaces: []string{"pk"},
				},
			},
		},
		&Compose{
			ValueStore: &LimitedValueStore{
				ValueStore: new(dummyValueStore),
				Namespaces: []string{"other", "pk"},
			},
		},
		&struct{ Compose }{Compose{ValueStore: &LimitedValueStore{ValueStore: Null{}}}},
		&struct{ Compose }{},
	}

	pid, _ := peert.RandPeerID()

	ctx := context.Background()
	if _, err := d.GetPublicKey(ctx, pid); err != routing.ErrNotFound {
		t.Fatal(err)
	}
}
