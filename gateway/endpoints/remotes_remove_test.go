package endpoints

import (
	"net/http"
	"testing"

	"github.com/sahib/brig/gateway/remotesapi"
	"github.com/stretchr/testify/require"
)

func TestRemoteRemoveEndpoint(t *testing.T) {
	withState(t, func(s *testState) {
		require.Nil(t, s.State.rapi.Set(remotesapi.Remote{
			Name:        "bob",
			Fingerprint: "xxx",
		}))

		resp := s.mustRun(
			t,
			NewRemotesRemoveHandler(s.State),
			"POST",
			"http://localhost:5000/api/v0/remotes/remove",
			RemoteRemoveRequest{
				Name: "bob",
			},
		)

		require.Equal(t, http.StatusOK, resp.StatusCode)
		data := struct {
			Success bool `json:"success"`
		}{}
		mustDecodeBody(t, resp.Body, &data)
		require.Equal(t, true, data.Success)

		resp = s.mustRun(
			t,
			NewRemotesRemoveHandler(s.State),
			"POST",
			"http://localhost:5000/api/v0/remotes/remove",
			RemoteRemoveRequest{
				Name: "bob",
			},
		)

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}
