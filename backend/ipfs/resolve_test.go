package ipfs

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLocateUsers(t *testing.T) {
	t.Skip("This test needs work")

	fmt.Println("Starting alice node...")
	WithIpfs(t, func(alice *Node) {
		fmt.Println("Starting bob node...")
		WithIpfs(t, func(bob *Node) {
			time.Sleep(60 * time.Second)

			fmt.Println("Starting publish of alice...")
			err := alice.PublishName("alice@wonderland.org/res")
			require.Nil(t, err)

			fmt.Println("Starting publish of bob...")
			err = bob.PublishName("bob@wonderland.org/home")
			require.Nil(t, err)

			fmt.Println("Starting alice resolve of bob...")
			ctx := context.Background()
			peers, err := alice.ResolveName(ctx, "bob@wonderland.org/home")
			require.Nil(t, err)
			fmt.Println(peers)
		})
	})
}
