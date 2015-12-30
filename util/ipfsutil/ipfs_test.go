package ipfsutil

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/disorganizer/brig/repo"
)

const (
	TEST_PATH = "/tmp/brig_test_ipfs_repo"
)

func initRepo(t *testing.T) *Context {
	if err := os.MkdirAll(TEST_PATH, 0744); err != nil {
		t.Errorf("Could not create unit test dir: %v", err)
		return nil
	}

	ipfsPath := filepath.Join(TEST_PATH, ".ipfs")
	if err := repo.CreateIpfsRepo(ipfsPath); err != nil {
		t.Errorf("Could not create ipfs repo at %s: %v", TEST_PATH, err)
		return nil
	}

	return &Context{Path: ipfsPath}
}

func TestStartDaemon(t *testing.T) {
	ctx := initRepo(t)
	if ctx == nil {
		return
	}

	defer os.RemoveAll(TEST_PATH)

	cmd, err := StartDaemon(ctx)
	if err != nil {
		t.Errorf("Could not start ipfs daemon: %v", err)
		return
	}

	if err := cmd.Process.Kill(); err != nil {
		t.Errorf("Could not kill ipfs daemon: %v", err)
		return
	}
}

func TestAddCat(t *testing.T) {
	ctx := initRepo(t)
	if ctx == nil {
		return
	}

	defer os.RemoveAll(TEST_PATH)

	// Dummy in-memory reader:
	origData := []byte("Hello World")
	buf := &bytes.Buffer{}
	buf.Write(origData)

	hash, err := Add(ctx, buf)
	if err != nil {
		t.Errorf("Add of a simple file failed: %v", err)
		return
	}

	reader, err := Cat(ctx, hash)
	if err != nil {
		t.Errorf("Could not cat simple file: %v", err)
		return
	}

	data, err := ioutil.ReadAll(reader)
	if err != nil {
		t.Errorf("Could not read back added data: %v", err)
		return
	}

	if !bytes.Equal(data, origData) {
		t.Errorf("Data not equal: %v <-> %v", string(data), string(origData))
	}
}
