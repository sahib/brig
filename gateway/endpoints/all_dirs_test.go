package endpoints

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAllDirsSuccess(t *testing.T) {
	withState(t, func(s *testState) {
		s.mustChangeFolders(t, "/a")
		require.Nil(t, s.fs.Mkdir("/a/b/c", true))
		require.Nil(t, s.fs.Mkdir("/d/e/f", true))

		resp := s.mustRun(
			t,
			NewAllDirsHandler(s.State),
			"POST",
			"http://localhost:5000/api/v0/all_dirs",
			nil,
		)

		require.Equal(t, http.StatusOK, resp.StatusCode)

		allDirsResp := &AllDirsResponse{}
		mustDecodeBody(t, resp.Body, allDirsResp)
		require.Equal(t, true, allDirsResp.Success)
		require.Equal(t, []string{"/a", "/a/b", "/a/b/c"}, allDirsResp.Paths)
	})
}
