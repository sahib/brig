package endpoints

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

type pinResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func TestPinEndpointSuccess(t *testing.T) {
	withState(t, func(s *testState) {
		require.Nil(t, s.fs.Touch("/file"))
		require.Nil(t, s.fs.Mkdir("/dir", true))

		resp := s.mustRun(
			t,
			NewPinHandler(s.State),
			"POST",
			"http://localhost:5000/api/v0/pin",
			&PinRequest{
				Path:     "/file",
				Revision: "curr",
			},
		)

		require.Equal(t, http.StatusOK, resp.StatusCode)

		pinResp := &pinResponse{}
		mustDecodeBody(t, resp.Body, &pinResp)
		require.Equal(t, true, pinResp.Success)

		stat, err := s.fs.Stat("/file")
		require.Nil(t, err)
		require.True(t, stat.IsPinned)
		require.True(t, stat.IsExplicit)

		stat, err = s.fs.Stat("/dir")
		require.Nil(t, err)
		require.True(t, stat.IsPinned)
		require.True(t, stat.IsExplicit)
	})
}

func TestPinEndpointForbidden(t *testing.T) {
	withState(t, func(s *testState) {
		s.mustChangeFolders(t, "/public")
		resp := s.mustRun(
			t,
			NewPinHandler(s.State),
			"POST",
			"http://localhost:5000/api/v0/pin",
			&PinRequest{
				Path:     "/file",
				Revision: "curr",
			},
		)

		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		pinResp := &pinResponse{}
		mustDecodeBody(t, resp.Body, &pinResp)
		require.Equal(t, false, pinResp.Success)
	})
}
