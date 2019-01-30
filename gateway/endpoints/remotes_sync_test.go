package endpoints

import (
	"net/http"
	"testing"

	"github.com/sahib/brig/gateway/remotesapi"
	"github.com/stretchr/testify/require"
)

func TestRemoteSyncEndpoint(t *testing.T) {
	withState(t, func(s *testState) {
		require.Nil(t, s.State.rapi.Set(remotesapi.Remote{
			Name:        "bob",
			Fingerprint: "xxx",
		}))

		resp := s.mustRun(
			t,
			NewRemotesSyncHandler(s.State),
			"POST",
			"http://localhost:5000/api/v0/remotes/sync",
			RemoteSyncRequest{
				Name: "bob",
			},
		)

		require.Equal(t, http.StatusOK, resp.StatusCode)
		data := struct {
			Success bool `json:"success"`
		}{}
		mustDecodeBody(t, resp.Body, &data)
		require.Equal(t, true, data.Success)
	})
}
