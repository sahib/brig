package gateway

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sahib/brig/catfs"
	"github.com/sahib/brig/defaults"
	"github.com/sahib/config"
	"github.com/stretchr/testify/require"
)

func withBasicGateway(t *testing.T, fn func(gw *Gateway, fs *catfs.FS)) {
	tmpDir, err := ioutil.TempDir("", "brig-gateway-tests")
	require.Nil(t, err)

	defer func() {
		os.RemoveAll(tmpDir)
	}()

	cfg, err := config.Open(nil, defaults.Defaults, config.StrictnessPanic)
	require.Nil(t, err)

	fs, err := catfs.NewFilesystem(
		catfs.NewMemFsBackend(),
		filepath.Join(tmpDir, "fs"),
		"ali",
		false,
		cfg.Section("fs"),
	)

	require.Nil(t, err)

	cfg.SetBool("gateway.enabled", true)
	cfg.SetInt("gateway.port", 9999)
	gw, err := NewGateway(fs, cfg.Section("gateway"), filepath.Join(tmpDir, "users"))
	require.Nil(t, err)

	require.Nil(t, gw.userDb.Add("ali", "ila", []string{"/"}))

	gw.Start()

	defer func() {
		require.Nil(t, gw.Stop())
	}()

	time.Sleep(100 * time.Millisecond)
	fn(gw, fs)
}

func buildURL(gw *Gateway, suffix string) string {
	suffix = strings.TrimLeft(suffix, "/")
	return fmt.Sprintf("http://localhost:%d/%s", gw.cfg.Int("port"), suffix)
}

func ping(t *testing.T, gw *Gateway) bool {
	_, err := http.Get(buildURL(gw, ""))
	return err == nil
}

func queryWithAuth(t *testing.T, gw *Gateway, suffix, user, pass string) (int, []byte) {
	req, err := http.NewRequest("GET", buildURL(gw, suffix), nil)
	require.Nil(t, err, fmt.Sprintf("%v", err))

	req.SetBasicAuth(user, pass)
	client := &http.Client{}
	resp, err := client.Do(req)
	require.Nil(t, err, fmt.Sprintf("%v", err))
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	require.Nil(t, err, fmt.Sprintf("%v", err))

	return resp.StatusCode, data
}

func TestGatewayOK(t *testing.T) {
	withBasicGateway(t, func(gw *Gateway, fs *catfs.FS) {
		exampleData := []byte("Hello world")
		err := fs.Stage("/hello/world.png", bytes.NewReader(exampleData))
		require.Nil(t, err, fmt.Sprintf("%v", err))

		status, data := queryWithAuth(t, gw, "/get/hello/world.png", "ali", "ila")
		require.Equal(t, 200, status)
		require.Equal(t, exampleData, data)
	})
}

func TestGatewayNoSuchFile(t *testing.T) {
	withBasicGateway(t, func(gw *Gateway, fs *catfs.FS) {
		status, data := queryWithAuth(t, gw, "/get/hello/world.png", "ali", "ila")
		require.Equal(t, 404, status)
		require.Equal(t, "not found", string(bytes.TrimSpace(data)))
	})
}

func TestGatewayUnauthorizedBadFolder(t *testing.T) {
	withBasicGateway(t, func(gw *Gateway, fs *catfs.FS) {
		exampleData := []byte("Hello world")
		err := fs.Stage("/hello/world.png", bytes.NewReader(exampleData))
		require.Nil(t, err, fmt.Sprintf("%v", err))

		status, data := queryWithAuth(t, gw, "/get/hello/world.png", "ali", "ila")
		require.Equal(t, 401, status)
		require.Equal(t, "not authorized", string(bytes.TrimSpace(data)))
	})
}

func TestGatewayUnauthorizedBadUser(t *testing.T) {
	withBasicGateway(t, func(gw *Gateway, fs *catfs.FS) {
		exampleData := []byte("Hello world")
		err := fs.Stage("/hello/world.png", bytes.NewReader(exampleData))
		require.Nil(t, err, fmt.Sprintf("%v", err))

		status, data := queryWithAuth(t, gw, "/get/hello/world.png", "resu", "pass")
		require.Equal(t, 401, status)
		require.Equal(t, "not authorized", string(bytes.TrimSpace(data)))
	})
}

func TestGatewayUnauthorizedBadPass(t *testing.T) {
	withBasicGateway(t, func(gw *Gateway, fs *catfs.FS) {
		exampleData := []byte("Hello world")
		err := fs.Stage("/hello/world.png", bytes.NewReader(exampleData))
		require.Nil(t, err)

		status, data := queryWithAuth(t, gw, "/get/hello/world.png", "user", "ssap")
		require.Equal(t, 401, status)
		require.Equal(t, "not authorized", string(bytes.TrimSpace(data)))
	})
}

func TestGatewayConfigChangeEnabled(t *testing.T) {
	withBasicGateway(t, func(gw *Gateway, fs *catfs.FS) {
		exampleData := []byte("Hello world")
		err := fs.Stage("/hello/world.png", bytes.NewReader(exampleData))
		require.Nil(t, err)

		require.True(t, ping(t, gw))
		status, data := queryWithAuth(t, gw, "/get/hello/world.png", "ali", "ila")
		require.Equal(t, 200, status)
		require.Equal(t, exampleData, data)

		gw.cfg.SetBool("enabled", false)
		time.Sleep(10 * time.Millisecond)

		require.False(t, ping(t, gw))
	})
}

func TestGatewayConfigChangePort(t *testing.T) {
	t.Skip("TODO: This triggers some badger db bug. Investigate later.")

	withBasicGateway(t, func(gw *Gateway, fs *catfs.FS) {
		exampleData := []byte("Hello world")
		err := fs.Stage("/hello/world.png", bytes.NewReader(exampleData))
		require.Nil(t, err)

		require.True(t, ping(t, gw))
		status, data := queryWithAuth(t, gw, "/get/hello/world.png", "ali", "ila")
		require.Equal(t, 200, status)
		require.Equal(t, exampleData, data)

		gw.cfg.SetInt("port", 8888)
		time.Sleep(1 * time.Second)

		// should still work, the port changed.
		status, data = queryWithAuth(t, gw, "/get/hello/world.png", "ali", "ila")
		require.Equal(t, 200, status)
		require.Equal(t, exampleData, data)
	})
}
