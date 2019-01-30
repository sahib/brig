package endpoints

import (
	"net/http"
	"testing"

	"github.com/sahib/brig/gateway/remotesapi"
	"github.com/stretchr/testify/require"
)

func TestRemoteDiffEndpoint(t *testing.T) {
	withState(t, func(s *testState) {
		require.Nil(t, s.State.rapi.Set(remotesapi.Remote{
			Name:        "bob",
			Fingerprint: "xxx",
		}))

		resp := s.mustRun(
			t,
			NewRemotesDiffHandler(s.State),
			"POST",
			"http://localhost:5000/api/v0/remotes/diff",
			RemoteDiffRequest{
				Name: "bob",
			},
		)

		data := &RemoteDiffResponse{}

		require.Equal(t, http.StatusOK, resp.StatusCode)
		mustDecodeBody(t, resp.Body, &data)
		require.Equal(t, true, data.Success)

		// TODO: Currently the mock backend always returns an empty diff:
		require.Equal(t, 0, len(data.Diff.Added))
		require.Equal(t, 0, len(data.Diff.Conflict))
		require.Equal(t, 0, len(data.Diff.Ignored))
		require.Equal(t, 0, len(data.Diff.Merged))
		require.Equal(t, 0, len(data.Diff.Missing))
		require.Equal(t, 0, len(data.Diff.Moved))
		require.Equal(t, 0, len(data.Diff.Removed))
	})
}
