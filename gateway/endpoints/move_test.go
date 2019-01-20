package endpoints

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

type moveResponse struct {
	Success bool `json:"success"`
}

func TestMoveSuccess(t *testing.T) {
	withState(t, func(s *testState) {
		require.Nil(t, s.fs.Mkdir("/hinz", true))

		resp := s.mustRun(
			t,
			NewMoveHandler(s.State),
			"POST",
			"http://localhost:5000/api/v0/move",
			&MoveRequest{
				Source:      "/hinz",
				Destination: "/kunz",
			},
		)

		require.Equal(t, http.StatusOK, resp.StatusCode)

		moveResp := &moveResponse{}
		mustDecodeBody(t, resp.Body, moveResp)
		require.Equal(t, true, moveResp.Success)

		_, err := s.fs.Stat("/hinz")
		require.NotNil(t, err)

		kunzInfo, err := s.fs.Stat("/kunz")
		require.Nil(t, err)
		require.Equal(t, "/kunz", kunzInfo.Path)
	})
}

func TestMoveDisallowedSource(t *testing.T) {
	withState(t, func(s *testState) {
		s.mustChangeFolders(t, "/kunz")
		require.Nil(t, s.fs.Mkdir("/hinz", true))

		resp := s.mustRun(
			t,
			NewMoveHandler(s.State),
			"POST",
			"http://localhost:5000/api/v0/move",
			&MoveRequest{
				Source:      "/hinz",
				Destination: "/kunz",
			},
		)

		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestMoveDisallowedDest(t *testing.T) {
	withState(t, func(s *testState) {
		s.mustChangeFolders(t, "/hinz")
		require.Nil(t, s.fs.Mkdir("/hinz", true))

		resp := s.mustRun(
			t,
			NewMoveHandler(s.State),
			"POST",
			"http://localhost:5000/api/v0/move",
			&MoveRequest{
				Source:      "/hinz",
				Destination: "/kunz",
			},
		)

		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}
