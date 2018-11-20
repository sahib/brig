package ipfs

import (
	"io/ioutil"
	"testing"

	"github.com/sahib/brig/util/testutil"
)

// WithIpfs creates a new ipfs node and passes it to `f`.
func WithIpfs(t *testing.T, f func(*Node)) {
	tmpDir, err := ioutil.TempDir("", "brig-ipfs-")
	if err != nil {
		t.Errorf("Cannot create temp dir %v", err)
		return
	}

	WithIpfsAtPath(t, tmpDir, f)
}

// WithIpfsAtPath creates a new ipfs node at `path` and passes it to `f`.
func WithIpfsAtPath(t *testing.T, root string, f func(*Node)) {
	WithIpfsAtPathAndPort(t, root, 4001, f)
}

// WithIpfsAtPort is the same as WithIpfs with the ability to change `port`.
func WithIpfsAtPort(t *testing.T, port int, f func(*Node)) {
	tmpDir, err := ioutil.TempDir("", "brig-ipfs")
	if err != nil {
		t.Errorf("cannot create temp dir %v", err)
		return
	}

	WithIpfsAtPathAndPort(t, tmpDir, port, f)
}

// WithIpfsAtPathAndPort is the same as WithIpfs with the ability to change
// `port` and `path`.
func WithIpfsAtPathAndPort(t *testing.T, root string, port int, f func(*Node)) {
	withIpfsRepo(t, root, func(path string) {
		nd, err := NewWithPort(path, nil, port)
		if err != nil {
			t.Fatalf("with ipfs: %v", err)
		}

		f(nd)

		if err := nd.Close(); err != nil {
			t.Errorf("Closing ipfs-daemon failed: %v", err)
		}
	})
}

func withIpfsRepo(t *testing.T, root string, f func(repoPath string)) {
	if err := Init(root, 1024); err != nil {
		t.Errorf("Could not create ipfs repo: %v", err)
		return
	}

	defer testutil.Remover(t, root)

	f(root)
}
