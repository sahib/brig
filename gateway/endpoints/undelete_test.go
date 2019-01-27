package endpoints

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

type undeleteResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func TestUndeleteEndpointSuccess(t *testing.T) {
	withState(t, func(s *testState) {
		require.Nil(t, s.fs.Touch("/test"))
		require.Nil(t, s.fs.Touch("/dir"))
		require.Nil(t, s.fs.MakeCommit("create"))
		require.Nil(t, s.fs.Remove("/test"))
		require.Nil(t, s.fs.Remove("/dir"))
		require.Nil(t, s.fs.MakeCommit("remove"))

		resp := s.mustRun(
			t,
			NewUndeleteHandler(s.State),
			"POST",
			"http://localhost:5000/api/v0/undelete",
			&UndeleteRequest{
				Path: "/test",
			},
		)

		require.Equal(t, http.StatusOK, resp.StatusCode)

		undeleteResp := &undeleteResponse{}
		mustDecodeBody(t, resp.Body, &undeleteResp)
		require.Equal(t, true, undeleteResp.Success)

		info, err := s.fs.Stat("/test")
		require.Nil(t, err)
		require.Equal(t, "/test", info.Path)
		require.Equal(t, false, info.IsDir)
	})
}

func TestUndeleteEndpointInvalidPath(t *testing.T) {
	withState(t, func(s *testState) {
		s.mustChangeFolders(t, "/something/else")
		resp := s.mustRun(
			t,
			NewUndeleteHandler(s.State),
			"POST",
			"http://localhost:5000/api/v0/undelete",
			&UndeleteRequest{
				Path: "/test",
			},
		)

		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}
