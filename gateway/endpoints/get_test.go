package endpoints

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetEndpointSuccess(t *testing.T) {
	withState(t, func(s *testState) {
		fileData := []byte("HelloWorld")
		require.Nil(t, s.fs.Stage("/file", bytes.NewReader(fileData)))

		resp := s.mustRun(
			t,
			NewGetHandler(s.State),
			"GET",
			"http://localhost:5000/get/file",
			nil,
		)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		data, err := ioutil.ReadAll(resp.Body)
		require.Nil(t, err)
		require.Equal(t, fileData, data)
	})
}

func TestGetEndpointDisallowed(t *testing.T) {
	withState(t, func(s *testState) {
		fileData := []byte("HelloWorld")
		require.Nil(t, s.fs.Stage("/file", bytes.NewReader(fileData)))
		s.mustChangeFolders(t, "/public")

		resp := s.mustRun(
			t,
			NewGetHandler(s.State),
			"GET",
			"http://localhost:5000/get/file",
			nil,
		)

		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}
