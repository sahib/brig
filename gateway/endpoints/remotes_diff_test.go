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

		require.Equal(t, 2, len(data.Diff.Added))
		require.Equal(t, "/new_dir", data.Diff.Added[0].Path)
		require.Equal(t, "/new_file", data.Diff.Added[1].Path)

		require.Equal(t, 1, len(data.Diff.Removed))
		require.Equal(t, "/removed_file", data.Diff.Removed[0].Path)

		require.Equal(t, 1, len(data.Diff.Ignored))
		require.Equal(t, "/ignored", data.Diff.Ignored[0].Path)

		require.Equal(t, 1, len(data.Diff.Missing))
		require.Equal(t, "/missing", data.Diff.Missing[0].Path)

		require.Equal(t, 1, len(data.Diff.Conflict))
		require.Equal(t, "/conflict_src", data.Diff.Conflict[0].Src.Path)
		require.Equal(t, "/conflict_dst", data.Diff.Conflict[0].Dst.Path)

		require.Equal(t, 1, len(data.Diff.Moved))
		require.Equal(t, "/moved_src", data.Diff.Moved[0].Src.Path)
		require.Equal(t, "/moved_dst", data.Diff.Moved[0].Dst.Path)

		require.Equal(t, 1, len(data.Diff.Merged))
		require.Equal(t, "/merged_src", data.Diff.Merged[0].Src.Path)
		require.Equal(t, "/merged_dst", data.Diff.Merged[0].Dst.Path)
	})
}
