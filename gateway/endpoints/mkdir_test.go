package endpoints

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

type mkdirResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func TestMkdirEndpointSuccess(t *testing.T) {
	withState(t, func(s *testState) {
		resp := s.mustRun(
			t,
			NewMkdirHandler(s.State),
			"POST",
			"http://localhost:5000/api/v0/mkdir",
			&MkdirRequest{
				Path: "/test",
			},
		)

		require.Equal(t, http.StatusOK, resp.StatusCode)

		mkdirResp := &mkdirResponse{}
		mustDecodeBody(t, resp.Body, &mkdirResp)
		require.Equal(t, true, mkdirResp.Success)

		info, err := s.fs.Stat("/test")
		require.Nil(t, err)
		require.Equal(t, "/test", info.Path)
	})
}

func TestMkdirEndpointInvalidPath(t *testing.T) {
	withState(t, func(s *testState) {
		s.mustChangeFolders(t, "/something/else")
		resp := s.mustRun(
			t,
			NewMkdirHandler(s.State),
			"POST",
			"http://localhost:5000/api/v0/mkdir",
			&MkdirRequest{
				Path: "/test",
			},
		)

		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}
