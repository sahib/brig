package endpoints

import (
	"bytes"
	"context"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func mustDoUpload(t *testing.T, s *testState, name string, data []byte) *http.Response {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", path.Base(name))
	require.Nil(t, err)

	_, err = part.Write(data)
	require.Nil(t, err)
	require.Nil(t, writer.Close())

	req := httptest.NewRequest(
		"POST",
		"/api/v0/upload?root="+url.QueryEscape(path.Dir(name)),
		body,
	)
	user, err := s.userDb.Get("ali")
	require.Nil(t, err)
	req = req.WithContext(context.WithValue(req.Context(), dbUserKey("brig.db_user"), user))

	req.Header.Set("Content-Type", writer.FormDataContentType())
	rsw := httptest.NewRecorder()
	setSession(s.store, "ali", rsw, req)
	NewUploadHandler(s.State).ServeHTTP(rsw, req)
	return rsw.Result()
}

func TestUploadSuccess(t *testing.T) {
	withState(t, func(s *testState) {
		require.Nil(t, s.fs.Mkdir("/sub", true))
		resp := mustDoUpload(t, s, "/sub/new_file.png", []byte("hello"))

		require.Equal(t, http.StatusOK, resp.StatusCode)
		entries, err := s.fs.List("/sub", 1)
		require.Nil(t, err)
		require.Len(t, entries, 1)

		stream, err := s.fs.Cat("/sub/new_file.png")
		require.Nil(t, err)

		data, err := ioutil.ReadAll(stream)
		require.Nil(t, err)
		require.Equal(t, []byte("hello"), data)
	})
}

func TestUploadForbidden(t *testing.T) {
	withState(t, func(s *testState) {
		s.mustChangeFolders(t, "/public")
		resp := mustDoUpload(t, s, "/sub/new_file.png", []byte("hello"))

		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}
