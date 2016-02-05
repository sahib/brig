package ipfsutil

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/disorganizer/brig/util/testutil"
)

var (
	TestPath = filepath.Join(os.TempDir(), "brig_test_ipfs_repo")
)

func initRepo(t *testing.T) string {
	path, err := InitRepo(TestPath, 1024)
	if err != nil {
		t.Errorf("Could not create ipfs repo: %v", err)
		return ""
	}

	return path
}

func withIpfs(t *testing.T, f func(*Node)) {
	path, err := InitRepo(TestPath, 1024)
	if err != nil {
		t.Errorf("Could not create ipfs repo: %v", err)
		return
	}

	defer testutil.Remover(t, path)

	node, err := StartNode(path)
	if err != nil {
		t.Errorf("")
	}

	f(node)

	if err := node.Close(); err != nil {
		t.Errorf("Closing ipfs-daemon failed: %v", err)
	}
}

func TestStartDaemon(t *testing.T) {
	withIpfs(t, func(node *Node) {
		if node.IpfsNode == nil {
			t.Errorf("withIpfs created an invalid Node.")
		}
	})
}

func TestAddCat(t *testing.T) {
	withIpfs(t, func(node *Node) {
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
	})
}
