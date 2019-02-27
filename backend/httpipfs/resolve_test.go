package httpipfs

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPublishResolve(t *testing.T) {
	t.Skip("needs work")

	// Only use one ipfs instance, for test performance.
	WithDoubleIpfs(t, 1, func(t *testing.T, apiPortA, apiPortB int) {
		ndA, err := NewNode(apiPortA)
		require.Nil(t, err)

		ndB, err := NewNode(apiPortB)
		require.Nil(t, err)

		// self, err := ndA.Identity()
		// require.Nil(t, err)

		require.Nil(t, ndA.PublishName("alice"))
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		infos, err := ndB.ResolveName(ctx, "alice")
		require.Nil(t, err)

		// TODO: This test doesn't produce results yet,
		// most likely because of time issues (would need to run longer?)
		fmt.Println(infos)
	})
}
