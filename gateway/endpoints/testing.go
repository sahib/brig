package endpoints

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/sahib/brig/catfs"
	"github.com/sahib/brig/defaults"
	"github.com/sahib/brig/gateway/db"
	"github.com/sahib/config"
	"github.com/stretchr/testify/require"
)

const (
	testGwUser = "ali"
)

type testState struct {
	*State
}

func withState(t *testing.T, fn func(state *testState)) {
	tmpDir, err := ioutil.TempDir("", "brig-endpoints-test-userdb")
	require.Nil(t, err)

	defer func() {
		os.RemoveAll(tmpDir)
	}()

	cfg, err := config.Open(nil, defaults.Defaults, config.StrictnessPanic)
	require.Nil(t, err)

	fs, err := catfs.NewFilesystem(
		catfs.NewMemFsBackend(),
		filepath.Join(tmpDir, "fs"),
		testGwUser,
		false,
		cfg.Section("fs"),
	)
	require.Nil(t, err)

	userDb, err := db.NewUserDatabase(filepath.Join(tmpDir, "user"))
	require.Nil(t, err)
	require.Nil(t, userDb.Add("ali", "ila", nil))

	state, err := NewState(fs, cfg.Section("gateway"), NewEventsHandler(), userDb)
	require.Nil(t, err)

	fn(&testState{state})

	require.Nil(t, state.fs.Close())
	require.Nil(t, state.userDb.Close())
}

func mustEncodeBody(t *testing.T, v interface{}) io.Reader {
	buf := &bytes.Buffer{}
	require.Nil(t, json.NewEncoder(buf).Encode(v))
	return buf
}

func mustDecodeBody(t *testing.T, body io.Reader, v interface{}) {
	data, err := ioutil.ReadAll(body)
	require.Nil(t, err)
	require.Nil(t, json.NewDecoder(bytes.NewReader(data)).Decode(v))
}

func (s *testState) mustRun(t *testing.T, hdl http.Handler, verb, url string, jsonBody interface{}) *http.Response {
	req := httptest.NewRequest(verb, url, mustEncodeBody(t, jsonBody))
	rsw := httptest.NewRecorder()

	setSession(s.store, "ali", rsw, req)
	hdl.ServeHTTP(rsw, req)
	return rsw.Result()
}

func (s *testState) mustChangeFolders(t *testing.T, folders ...string) {
	require.Nil(t, s.userDb.Remove("ali"))
	require.Nil(t, s.userDb.Add("ali", "ila", folders))
}
