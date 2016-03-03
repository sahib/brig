package testwith

import (
	"testing"

	"github.com/disorganizer/brig/util/ipfsutil"
	"github.com/disorganizer/brig/util/testutil"
)

func WithIpfs(t *testing.T, root string, f func(*ipfsutil.Node)) {
	WithIpfsRepo(t, root, func(path string) {
		node := ipfsutil.New(path)
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
