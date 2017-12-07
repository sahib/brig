package net

import (
	"fmt"
	"testing"
	"time"

	"github.com/sahib/brig/net/mock"
	"github.com/stretchr/testify/require"
)

func TestPinger(t *testing.T) {
	nbk := mock.NewNetBackend()
	pmap := NewPingMap(nbk)

	require.Nil(t, pmap.Sync([]string{
		"alice-addr",
		"bob-addr",
		"vincent-addr",
		"charlie-addr-wrong",
		"something-not-there-addr",
	}))

	// Give it a bit of time to send the first pings.
	time.Sleep(50 * time.Millisecond)

	for _, addr := range []string{"alice-addr", "bob-addr"} {
		pinger, err := pmap.For(addr)
		require.Nil(t, err)
		require.Nil(t, pinger.Err())

		// TODO: Actually assert that some correct values are here:
		fmt.Println(pinger.LastSeen())
		fmt.Println(pinger.Roundtrip())
	}

	pinger, err := pmap.For("charlie-addr-wrong")
	require.Nil(t, err)
	require.Nil(t, pinger)
}
