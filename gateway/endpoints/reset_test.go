package endpoints

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

type resetResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func TestResetSuccess(t *testing.T) {
	withState(t, func(s *testState) {
		require.Nil(t, s.fs.Stage("/file", bytes.NewReader([]byte("hello"))))
		require.Nil(t, s.fs.MakeCommit("add"))
		require.Nil(t, s.fs.Stage("/file", bytes.NewReader([]byte("world"))))
		require.Nil(t, s.fs.MakeCommit("modify"))

		resp := s.mustRun(
			t,
			NewResetHandler(s.State),
			"POST",
			"http://localhost:5000/api/v0/reset",
			&ResetRequest{
				Path:     "/file",
				Revision: "init",
			},
		)

		require.Equal(t, http.StatusOK, resp.StatusCode)

		resetResp := &resetResponse{}
		mustDecodeBody(t, resp.Body, &resetResp)
		require.Equal(t, true, resetResp.Success)

		stream, err := s.fs.Cat("/file")
		require.Nil(t, err)

		data, err := ioutil.ReadAll(stream)
		require.Nil(t, err)
		require.Equal(t, []byte("hello"), data)
	})
}

func TestResetForbidden(t *testing.T) {
	withState(t, func(s *testState) {
		s.mustChangeFolders(t, "/public")
		resp := s.mustRun(
			t,
			NewResetHandler(s.State),
			"POST",
			"http://localhost:5000/api/v0/reset",
			&ResetRequest{
				Path:     "/file",
				Revision: "init",
			},
		)

		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)

		resetResp := &resetResponse{}
		mustDecodeBody(t, resp.Body, &resetResp)
		require.Equal(t, false, resetResp.Success)
	})
}
