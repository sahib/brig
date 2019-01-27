package endpoints

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDeletedPathSuccess(t *testing.T) {
	withState(t, func(s *testState) {
		s.mustChangeFolders(t, "/a")
		require.Nil(t, s.fs.Touch("/a/b/c1"))
		require.Nil(t, s.fs.Touch("/a/b/c2"))
		require.Nil(t, s.fs.Touch("/d/e/f1"))
		require.Nil(t, s.fs.Touch("/d/e/f2"))

		require.Nil(t, s.fs.MakeCommit("add"))

		require.Nil(t, s.fs.Remove("/a/b/c1"))
		require.Nil(t, s.fs.Remove("/a/b/c2"))
		require.Nil(t, s.fs.Remove("/d/e/f1"))
		require.Nil(t, s.fs.Remove("/d/e/f2"))
		require.Nil(t, s.fs.MakeCommit("rm"))

		resp := s.mustRun(
			t,
			NewDeletedPathsHandler(s.State),
			"POST",
			"http://localhost:5000/api/v0/deleted",
			nil,
		)

		require.Equal(t, http.StatusOK, resp.StatusCode)

		deletedResp := &DeletedPathsResponse{}
		mustDecodeBody(t, resp.Body, deletedResp)
		require.Equal(t, true, deletedResp.Success)
		require.Equal(t, 2, len(deletedResp.Entries))

		paths := []string{}
		for _, entry := range deletedResp.Entries {
			paths = append(paths, entry.Path)
		}

		require.Equal(
			t,
			[]string{"/a/b/c1", "/a/b/c2"},
			paths,
		)
	})
}
