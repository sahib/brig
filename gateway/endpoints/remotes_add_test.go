package endpoints

import (
	"net/http"
	"testing"

	"github.com/sahib/brig/gateway/remotesapi"
	"github.com/stretchr/testify/require"
)

func TestRemoteAddEndpoint(t *testing.T) {
	withState(t, func(s *testState) {
		resp := s.mustRun(
			t,
			NewRemotesAddHandler(s.State),
			"POST",
			"http://localhost:5000/api/v0/remotes/add",
			&RemoteAddRequest{
				Name:              "bob",
				Folders:           nil,
				Fingerprint:       "bobsfingerprint",
				AcceptAutoUpdates: true,
			},
		)

		require.Equal(t, http.StatusOK, resp.StatusCode)

		data := struct {
			Success bool `json:"success"`
		}{}
		mustDecodeBody(t, resp.Body, &data)
		require.Equal(t, true, data.Success)

		rmt, err := s.State.rapi.Get("bob")
		require.Nil(t, err)
		require.Equal(t, "bob", rmt.Name)
		require.Equal(t, "bobsfingerprint", rmt.Fingerprint)
		require.Equal(t, true, rmt.AcceptAutoUpdates)
	})
}

func TestRemoteModifyEndpoint(t *testing.T) {
	withState(t, func(s *testState) {
		require.Nil(t, s.State.rapi.Set(remotesapi.Remote{
			Name:        "bob",
			Fingerprint: "oldfingerprint",
			Folders:     []string{"/public"},
		}))

		resp := s.mustRun(
			t,
			NewRemotesModifyHandler(s.State),
			"POST",
			"http://localhost:5000/api/v0/remotes/modify",
			&RemoteAddRequest{
				Name:              "bob",
				Folders:           nil,
				Fingerprint:       "bobsfingerprint",
				AcceptAutoUpdates: true,
			},
		)

		require.Equal(t, http.StatusOK, resp.StatusCode)

		data := struct {
			Success bool `json:"success"`
		}{}

		mustDecodeBody(t, resp.Body, &data)
		require.Equal(t, true, data.Success)

		rmt, err := s.State.rapi.Get("bob")
		require.Nil(t, err)
		require.Equal(t, "bob", rmt.Name)
		require.Equal(t, "bobsfingerprint", rmt.Fingerprint)
		require.Equal(t, true, rmt.AcceptAutoUpdates)
	})
}
