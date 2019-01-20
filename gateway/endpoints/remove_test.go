package endpoints

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

type removeResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func TestRemoveEndpointSuccess(t *testing.T) {
	withState(t, func(s *TestState) {
		require.Nil(t, s.fs.Touch("/file"))
		require.Nil(t, s.fs.Mkdir("/dir", true))

		resp := s.mustRun(
			t,
			NewRemoveHandler(s.State),
			"POST",
			"http://localhost:5000/api/v0/remove",
			&RemoveRequest{
				Paths: []string{"/file", "/dir"},
			},
		)

		require.Equal(t, http.StatusOK, resp.StatusCode)

		removeResp := &removeResponse{}
		mustDecodeBody(t, resp.Body, &removeResp)
		require.Equal(t, true, removeResp.Success)

		_, err := s.fs.Stat("/file")
		require.NotNil(t, err)

		_, err = s.fs.Stat("/dir")
		require.NotNil(t, err)
	})
}

func TestRemoveEndpointForbidden(t *testing.T) {
	withState(t, func(s *TestState) {
		s.mustChangeFolders(t, "/public")
		resp := s.mustRun(
			t,
			NewRemoveHandler(s.State),
			"POST",
			"http://localhost:5000/api/v0/remove",
			&RemoveRequest{
				Paths: []string{"/file", "/dir"},
			},
		)

		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		removeResp := &removeResponse{}
		mustDecodeBody(t, resp.Body, &removeResp)
		require.Equal(t, false, removeResp.Success)
	})
}
