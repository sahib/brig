package endpoints

import (
	"net/http"
	"testing"

	"github.com/sahib/brig/gateway/remotesapi"
	"github.com/stretchr/testify/require"
)

func TestRemoteListEndpoint(t *testing.T) {
	withState(t, func(s *testState) {
		require.Nil(t, s.State.rapi.Set(remotesapi.Remote{
			Name:        "bob",
			Fingerprint: "xxx",
		}))

		require.Nil(t, s.State.rapi.Set(remotesapi.Remote{
			Name:              "charlie",
			Fingerprint:       "yyy",
			AcceptAutoUpdates: true,
			Folders:           []string{"/public"},
		}))

		resp := s.mustRun(
			t,
			NewRemotesListHandler(s.State),
			"POST",
			"http://localhost:5000/api/v0/remotes/list",
			nil,
		)

		require.Equal(t, http.StatusOK, resp.StatusCode)

		data := &RemoteListResponse{}
		mustDecodeBody(t, resp.Body, &data)

		require.Equal(t, true, data.Success)
		require.Equal(t, 2, len(data.Remotes))
		require.Equal(t, "bob", data.Remotes[0].Name)
		require.Equal(t, "xxx", data.Remotes[0].Fingerprint)
		require.Equal(t, false, data.Remotes[0].AcceptAutoUpdates)

		require.Equal(t, "charlie", data.Remotes[1].Name)
		require.Equal(t, "yyy", data.Remotes[1].Fingerprint)
		require.Equal(t, true, data.Remotes[1].AcceptAutoUpdates)
	})
}
