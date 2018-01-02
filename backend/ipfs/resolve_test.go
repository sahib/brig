package ipfs

import (
	"fmt"
	"testing"
	"time"

	"github.com/sahib/brig/net/peer"
	"github.com/stretchr/testify/require"
)

func TestLocateUsers(t *testing.T) {
	fmt.Println("Starting alice node...")
	WithIpfs(t, func(alice *Node) {
		fmt.Println("Starting bob node...")
		WithIpfs(t, func(bob *Node) {
			time.Sleep(60 * time.Second)

			fmt.Println("Starting publish of alice...")
			err := alice.PublishName(peer.Name("alice@wonderland.org/res"))
			require.Nil(t, err)

			fmt.Println("Starting publish of bob...")
			err = bob.PublishName(peer.Name("bob@wonderland.org/home"))
			require.Nil(t, err)

			fmt.Println("Starting alice resolve of bob...")
			peers, err := alice.ResolveName("bob@wonderland.org/home")
			require.Nil(t, err)
			fmt.Println(peers)
		})
	})
}
