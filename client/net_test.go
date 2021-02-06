package client_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/sahib/brig/client"
	"github.com/stretchr/testify/require"
)

func TestPush(t *testing.T) {
	withDaemonPair(t, "ali", "bob", func(aliCtl, bobCtl *client.Client) {
		require.Nil(t, aliCtl.StageFromReader("/ali-file", bytes.NewReader([]byte{1, 2, 3})))

		err := aliCtl.Push("bob", true)
		require.True(t, strings.HasSuffix(err.Error(), "remote does not allow it"))

		aliRmt, err := bobCtl.RemoteByName("ali")
		require.Nil(t, err)
		aliRmt.AcceptPush = true
		require.Nil(t, bobCtl.RemoteAddOrUpdate(aliRmt))

		err = aliCtl.Push("bob", true)
		require.Nil(t, err)

		err = aliCtl.Push("bob", false)
		require.Nil(t, err)

		// There is a possible race condition here:
		// ``brig push`` only triggers the sync, but
		// waits only until the network message was sent.
		// It might take a small amount of time till the other
		// side managed to do the sync.
		time.Sleep(250 * time.Millisecond)

		// bob should have ali file without him syncing explicitly.
		_, err = bobCtl.Stat("/ali-file")
		require.Nil(t, err)

	})
}
