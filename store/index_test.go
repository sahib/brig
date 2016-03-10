package store

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/disorganizer/brig/util/ipfsutil"
	"github.com/disorganizer/brig/util/testutil"
	"github.com/disorganizer/brig/util/testwith"
)

var TestPath = filepath.Join(os.TempDir(), "brig-store-test")

func withEmptyStore(t *testing.T, f func(*Store)) {
	ipfsPath := filepath.Join(TestPath, "ipfs")

	testwith.WithIpfs(t, ipfsPath, func(node *ipfsutil.Node) {
		if err := os.MkdirAll(TestPath, 0744); err != nil {
			t.Errorf("Could not create store dir at %s: %v", TestPath, err)
			return
		}

		defer func() {
			if err := os.RemoveAll(TestPath); err != nil {
				t.Errorf("Could not remove temp dir for empty store.")
				return
			}
		}()

		// We need the filesystem for ipfs here:
		store, err := Open(TestPath, node)
		if err != nil {
			t.Errorf("Could not open empty store at %s: %v", TestPath, err)
			return
		}

		f(store)

		if err := store.Close(); err != nil {
			t.Errorf("Unable to close empty store: %v", err)
			return
		}
	})
}

func TestAddCat(t *testing.T) {
	sizes := []int64{
		0, 1, 2, 4, 1024, 4096, 16 * 1024,
	}

	for _, size := range sizes {
		data := testutil.CreateDummyBuf(size)
		path := fmt.Sprintf("dummy_%d", size)

		withEmptyStore(t, func(st *Store) {
			if err := st.AddFromReader(path, bytes.NewReader(data), size); err != nil {
				t.Errorf("Adding of `%s` failed: %v", path, err)
			}

			recvBuf := &bytes.Buffer{}
			if err := st.Cat(path, recvBuf); err != nil {
				t.Errorf("Catting of `%s` failed: %v", path, err)
			}
		})
	}
}

func TestListEntries(t *testing.T) {
	paths := []string{
		"/a", "/b/b1", "/b/b2", "/c/cc/ccc",
	}

	tests := []struct {
		root  string
		depth int
		want  []string
	}{
		{"/", +0, []string{"/"}},
		{"/", +1, []string{"/", "/a", "/b", "/c"}},
		{"/", -1, []string{"/", "/a", "/b", "/b/b1", "/b/b2", "/c", "/c/cc", "/c/cc/ccc"}},
		{"/c", -1, []string{"/c", "/c/cc", "/c/cc/ccc"}},
		{"/a", -1, []string{"/a"}},
	}

	withEmptyStore(t, func(st *Store) {
		// Build the tree:
		for _, path := range paths {
			if err := st.AddFromReader(path, bytes.NewReader(nil), 0); err != nil {
				t.Errorf("Adding of `%s` failed: %v", path, err)
				break
			}
		}

		// Run the actual tests on it:
		for _, test := range tests {
			dirlist, err := st.ListEntries(test.root, test.depth)
			if err != nil {
				t.Errorf("Listing `%s` failed: %v", "/", err)
				break
			}

			sorted := []string{}
			for _, e := range dirlist.Entries {
				sorted = append(sorted, e.GetPath())
			}

			sort.Strings(sorted)

			if len(sorted) != len(test.want) {
				t.Errorf(
					"Length of want (%d) and got (%d) differs.",
					len(test.want),
					len(sorted),
				)
				break
			}

			for idx := range sorted {
				if sorted[idx] != test.want[idx] {
					t.Errorf("List order differs at index %d", idx)
					break
				}
			}
		}
	})
}

func TestExport(t *testing.T) {
	paths := []string{
		"/root", "/pics/me.png", "/pics/him.png",
	}

	exportBuf := &bytes.Buffer{}

	withEmptyStore(t, func(store *Store) {
		for _, path := range paths {
			if err := store.Touch(path); err != nil {
				t.Errorf("Touching file `%s` failed: %v", path, err)
				return
			}
		}

		if err := store.Export(exportBuf); err != nil {
			t.Errorf("store-export failed: %v", err)
			return
		}
	})

	exportData := exportBuf.Bytes()

	if len(exportData) == 0 {
		t.Errorf("Exported data is empty.")
		return
	}

	withEmptyStore(t, func(store *Store) {
		if err := store.Import(bytes.NewReader(exportData)); err != nil {
			t.Errorf("Could not import data: %v", err)
			return
		}
	})
}
