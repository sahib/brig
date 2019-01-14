package endpoints

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/sahib/brig/catfs"
	"github.com/sahib/brig/defaults"
	"github.com/sahib/config"
	"github.com/stretchr/testify/require"
)

const (
	testDbPath  = "/tmp/gw-standalone.db"
	testGwUser  = "ali"
	testCsrfKey = "00000000000000000000000000000000"
)

func withFsAndCfg(t *testing.T, fn func(cfg *config.Config, fs *catfs.FS)) {
	defer func() {
		os.RemoveAll(testDbPath)
	}()

	cfg, err := config.Open(nil, defaults.Defaults, config.StrictnessPanic)
	require.Nil(t, err)

	cfg.SetBool("gateway.enabled", true)
	cfg.SetInt("gateway.port", 5000)
	cfg.SetBool("gateway.cert.redirect.enabled", false)

	cfg.SetStrings("gateway.folders", []string{"/"})

	// TODO: Do we really need https for tests?
	cfg.SetString("gateway.cert.domain", "nwzmlh4iouqikobq.myfritz.net")
	cfg.SetString("gateway.cert.certfile", "/tmp/fullchain.pem")
	cfg.SetString("gateway.cert.keyfile", "/tmp/privkey.pem")

	fs, err := catfs.NewFilesystem(
		catfs.NewMemFsBackend(),
		testDbPath,
		testGwUser,
		false,
		cfg.Section("fs"),
	)
	require.Nil(t, err)

	exampleData := bytes.NewReader([]byte("Hello world"))
	err = fs.Stage("/hello/world.png", exampleData)
	require.Nil(t, err)

	fn(cfg.Section("gateway"), fs)
	require.Nil(t, fs.Close())
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

func mustCreateRequest(t *testing.T, verb string, url string, jsonBody interface{}) *http.Request {
	req := httptest.NewRequest(verb, url, mustEncodeBody(t, jsonBody))
	// TODO: rewrite this to use sessions:
	// encoded, err := cookieHandler.Encode("session", map[string]string{"name": "ali"})
	// require.Nil(t, err)

	// req.AddCookie(&http.Cookie{
	// 	Name:  "session",
	// 	Value: encoded,
	// 	Path:  "/",
	// })

	return req
}
