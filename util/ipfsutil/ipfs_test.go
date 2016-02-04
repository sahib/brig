package ipfsutil

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	ipfsconfig "github.com/ipfs/go-ipfs/repo/config"
	"github.com/ipfs/go-ipfs/repo/fsrepo"
)

var (
	TestPath = filepath.Join(os.TempDir(), "brig_test_ipfs_repo")
)

func initRepo(t *testing.T) string {
	if err := os.MkdirAll(TestPath, 0744); err != nil {
		t.Errorf("Could not create unit test dir: %v", err)
		return ""
	}

	ipfsPath := filepath.Join(TestPath, ".ipfs")
	cfg, err := ipfsconfig.Init(ioutil.Discard, 1024)
	if err != nil {
		t.Errorf("Could not create ipfs config %v", err)
		return ""
	}

	if err := fsrepo.Init(ipfsPath, cfg); err != nil {
		t.Errorf("Could not create ipfs repo at %s: %v", TestPath, err)
		return ""
	}

	return ipfsPath
}

func TestStartDaemon(t *testing.T) {
	path := initRepo(t)

	defer func() {
		if err := os.RemoveAll(TestPath); err != nil {
			t.Errorf("Unable to remove daemon temp dir: %v", err)
		}
	}()

	node, err := StartNode(path)
	if err != nil {
		t.Errorf("Could not start ipfs daemon: %v", err)
	}

	if err := node.IpfsNode.Close(); err != nil {
		t.Errorf("Closing ipfs-daemon failed: %v", err)
	}
}

func TestAddCat(t *testing.T) {
	path := initRepo(t)

	defer func() {
		if err := os.RemoveAll(TestPath); err != nil {
			t.Errorf("Unable to remove daemon temp dir: %v", err)
		}
	}()

	node, err := StartNode(path)
	if err != nil {
		t.Errorf("Could not start ipfs daemon: %v", err)
		return
	}

	defer func() {
		if err := node.IpfsNode.Close(); err != nil {
			t.Errorf("Could not kill ipfs daemon: %v", err)
			return
		}
	}()

	// Dummy in-memory reader:
	origData := []byte("Hello World")
	buf := &bytes.Buffer{}
	buf.Write(origData)

	hash, err := Add(node, buf)
	if err != nil {
		t.Errorf("Add of a simple file failed: %v", err)
		return
	}

	reader, err := Cat(node, hash)
	if err != nil {
		t.Errorf("Could not cat simple file: %v", err)
		return
	}

	data, err := ioutil.ReadAll(reader)
	if err != nil {
		t.Errorf("Could not read back added data: %v", err)
		return
	}

	if err = reader.Close(); err != nil {
		t.Errorf("close(cat) failed: %v", err)
		return
	}

	if !bytes.Equal(data, origData) {
		t.Errorf("Data not equal: %v <- -> %v", string(data), string(origData))
	}
}
