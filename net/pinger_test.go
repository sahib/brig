package net

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPinger(t *testing.T) {
	withNetPair(t, func(a, b testUnit) {
		apmap := a.srv.PingMap()
		apmap.Sync([]string{"bob@9999"})

		bpmap := a.srv.PingMap()
		bpmap.Sync([]string{"alice@9998"})

		// Give it a bit of time to send the first pings.
		time.Sleep(100 * time.Millisecond)

		aliPinger, err := apmap.For("alice@9998")
		require.Nil(t, err)
		require.Nil(t, aliPinger.Err())

		require.True(t, aliPinger.Roundtrip() < 1*time.Millisecond)

		bobPinger, err := bpmap.For("bob@9999")
		require.Nil(t, err)
		require.Nil(t, bobPinger.Err())

		require.True(t, bobPinger.Roundtrip() < 1*time.Millisecond)

		charliePinger, err := bpmap.For("charlie@9999")
		require.Nil(t, charliePinger)
		require.NotNil(t, err)
	})
}
