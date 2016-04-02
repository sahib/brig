package testwith

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/disorganizer/brig/util/ipfsutil"
	"github.com/disorganizer/brig/util/testutil"
)

var (
	TestDir = filepath.Join(os.TempDir(), ".brig_ipfs_tests")
)

func WithIpfs(t *testing.T, f func(*ipfsutil.Node)) {
	WithIpfsAtPath(t, TestDir, f)
}

func WithIpfsAtPort(t *testing.T, port int, f func(*ipfsutil.Node)) {
	WithIpfsAtPathAndPort(t, TestDir, port, f)
}

func WithIpfsAtPath(t *testing.T, root string, f func(*ipfsutil.Node)) {
	WithIpfsAtPathAndPort(t, root, 4001, f)
}

func WithIpfsAtPathAndPort(t *testing.T, root string, port int, f func(*ipfsutil.Node)) {
	WithIpfsRepo(t, root, func(path string) {
		node := ipfsutil.NewWithPorts(path, 1000+port, port)
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
