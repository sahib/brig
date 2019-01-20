package endpoints

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

type copyResponse struct {
	Success bool `json:"success"`
}

func TestCopySuccess(t *testing.T) {
	withState(t, func(s *TestState) {
		require.Nil(t, s.fs.Mkdir("/hinz", true))
		resp := s.mustRun(
			t,
			NewCopyHandler(s.State),
			"POST",
			"http://localhost:5000/api/v0/copy",
			&CopyRequest{
				Source:      "/hinz",
				Destination: "/kunz",
			},
		)

		require.Equal(t, http.StatusOK, resp.StatusCode)

		copyResp := &copyResponse{}
		mustDecodeBody(t, resp.Body, copyResp)
		require.Equal(t, true, copyResp.Success)

		hinzInfo, err := s.fs.Stat("/hinz")
		require.Nil(t, err)
		require.Equal(t, "/hinz", hinzInfo.Path)

		kunzInfo, err := s.fs.Stat("/kunz")
		require.Nil(t, err)
		require.Equal(t, "/kunz", kunzInfo.Path)
	})
}

func TestCopyDisallowedSource(t *testing.T) {
	withState(t, func(s *TestState) {
		s.mustChangeFolders(t, "/kunz")
		require.Nil(t, s.fs.Mkdir("/hinz", true))

		resp := s.mustRun(
			t,
			NewCopyHandler(s.State),
			"POST",
			"http://localhost:5000/api/v0/copy",
			&CopyRequest{
				Source:      "/hinz",
				Destination: "/kunz",
			},
		)

		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestCopyDisallowedDest(t *testing.T) {
	withState(t, func(s *TestState) {
		s.mustChangeFolders(t, "/hinz")
		require.Nil(t, s.fs.Mkdir("/hinz", true))

		resp := s.mustRun(
			t,
			NewCopyHandler(s.State),
			"POST",
			"http://localhost:5000/api/v0/copy",
			&CopyRequest{
				Source:      "/hinz",
				Destination: "/kunz",
			},
		)

		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}
