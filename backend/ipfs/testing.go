package ipfsutil

import (
	"io/ioutil"
	"testing"

	"github.com/disorganizer/brig/util/ipfsutil"
	"github.com/disorganizer/brig/util/testutil"
)

func WithIpfs(t *testing.T, f func(*ipfsutil.Node)) {
	tmpDir, err := ioutil.TempDir("", "brig-ipfs")
	if err != nil {
		t.Errorf("Cannot create temp dir %v", err)
		return
	}
	WithIpfsAtPath(t, tmpDir, f)
}

func WithIpfsAtPort(t *testing.T, port int, f func(*ipfsutil.Node)) {
	tmpDir, err := ioutil.TempDir("", "brig-ipfs")
	if err != nil {
		t.Errorf("Cannot create temp dir %v", err)
		return
	}
	WithIpfsAtPathAndPort(t, tmpDir, port, f)
}

func WithIpfsAtPath(t *testing.T, root string, f func(*ipfsutil.Node)) {
	WithIpfsAtPathAndPort(t, root, 4001, f)
}

func WithIpfsAtPathAndPort(t *testing.T, root string, port int, f func(*ipfsutil.Node)) {
	WithIpfsRepo(t, root, func(path string) {
		node := ipfsutil.NewWithPort(path, port)
		f(node)

		if err := node.Close(); err != nil {
			t.Errorf("Closing ipfs-daemon failed: %v", err)
		}
	})
}

func WithIpfsRepo(t *testing.T, root string, f func(string)) {
	path, err := ipfsutil.InitRepo(root, 1024)
	if err != nil {
		t.Errorf("Could not create ipfs repo: %v", err)
		return
	}

	defer testutil.Remover(t, path)

	f(path)
}
