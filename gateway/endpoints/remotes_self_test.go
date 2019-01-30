package endpoints

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRemoteSelfEndpoint(t *testing.T) {
	withState(t, func(s *testState) {
		resp := s.mustRun(
			t,
			NewRemotesSelfHandler(s.State),
			"POST",
			"http://localhost:5000/api/v0/remotes/self",
			nil,
		)

		require.Equal(t, http.StatusOK, resp.StatusCode)

		data := &RemoteSelfResponse{}
		mustDecodeBody(t, resp.Body, &data)
		require.Equal(t, true, data.Success)
		require.Equal(t, "ali", data.Self.Name)
		require.Equal(t, "alisfingerprint", data.Self.Fingerprint)
	})
}
