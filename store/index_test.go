package store

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/disorganizer/brig/id"
	"github.com/disorganizer/brig/util/ipfsutil"
	"github.com/disorganizer/brig/util/testutil"
	"github.com/disorganizer/brig/util/testwith"
)

var TestPath = filepath.Join(os.TempDir(), "brig-store-test")

func withEmptyStore(t *testing.T, f func(*Store)) {
	ipfsPath := filepath.Join(TestPath, "ipfs")

	testwith.WithIpfsAtPath(t, ipfsPath, func(node *ipfsutil.Node) {
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
		store, err := Open(TestPath, id.ID("alice@nullcat.de/desktop"), node)
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
				return
			}

			recvBuf := &bytes.Buffer{}
			if err := st.Cat(path, recvBuf); err != nil {
				t.Errorf("Catting of `%s` failed: %v", path, err)
				return
			}

			if !bytes.Equal(recvBuf.Bytes(), data) {
				t.Errorf("Data differs between add and cat")
				return
			}
		})
	}
}

func TestList(t *testing.T) {
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
			dirlist, err := st.List(test.root, test.depth)
			if err != nil {
				t.Errorf("Listing `%s` failed: %v", "/", err)
				break
			}

			sorted := []string{}
			for _, e := range dirlist {
				sorted = append(sorted, e.Path())
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

func createDirAndFile(t *testing.T, st *Store, data []byte) error {
	if err := st.AddFromReader("/dummy", bytes.NewReader(data), int64(len(data))); err != nil {
		t.Errorf("Could not add dummy file for move: %v", err)
		return err
	}

	if _, err := st.Mkdir("/dir"); err != nil {
		t.Errorf("Mkdir(/dir) failed: %v", err)
		return err
	}

	if err := st.Touch("/dir/a"); err != nil {
		t.Errorf("Touch(/dir/a) failed: %v", err)
		return err
	}

	if err := st.Touch("/dir/b"); err != nil {
		t.Errorf("Touch(/dir/b) failed: %v", err)
		return err
	}

	return nil
}

func TestMove(t *testing.T) {
	data := testutil.CreateDummyBuf(1024)

	withEmptyStore(t, func(st *Store) {
		if err := createDirAndFile(t, st, data); err != nil {
			return
		}

		check := func(path string, expect []byte) {
			recvBuf := &bytes.Buffer{}
			if err := st.Cat(path, recvBuf); err != nil {
				t.Errorf("Catting of `%s` failed: %v", path, err)
				return
			}

			if !bytes.Equal(recvBuf.Bytes(), expect) {
				t.Errorf("Data differs between add/move/cat")
				return
			}
		}

		check("/dummy", data)

		if err := st.Move("/dummy", "/new_dummy"); err != nil {
			t.Errorf("Move failed: %v", err)
			return
		}

		if err := st.Cat("/dummy", &bytes.Buffer{}); err != ErrNoSuchFile {
			t.Errorf("Move: dummy still reachable")
			return
		}

		check("/new_dummy", data)

		if err := st.Move("/dummy", "/new_dummy"); err != ErrNoSuchFile {
			t.Errorf("Move could move dead file: %v", err)
			return
		}

		if err := st.Move("/dir", "/other"); err != nil {
			t.Errorf("Move could not move dir: %v", err)
			return
		}

		check("/other/a", []byte{})
		check("/other/b", []byte{})

		if err := st.Cat("/dir/a", &bytes.Buffer{}); err != ErrNoSuchFile {
			t.Errorf("Move: /dir/a still reachable")
			return
		}

		if err := st.Cat("/dir/b", &bytes.Buffer{}); err != ErrNoSuchFile {
			t.Errorf("Move: /dir/b still reachable")
			return
		}
	})
}

func TestRemove(t *testing.T) {
	data := testutil.CreateDummyBuf(1024)

	withEmptyStore(t, func(st *Store) {
		if err := createDirAndFile(t, st, data); err != nil {
			return
		}

		if _, err := st.Mkdir("/empty_dir"); err != nil {
			t.Errorf("Could not mkdir /empty_dir: %v", err)
			return
		}

		if err := st.Remove("/dummy", false); err != nil {
			t.Errorf("Could not remove /dummy: %v", err)
			return
		}

		if err := st.Remove("/dummy", false); err != ErrNoSuchFile {
			t.Errorf("Could remove /dummy twice: %v", err)
			return
		}

		if err := st.Remove("/dir", false); err != ErrNotEmpty {
			t.Errorf("Remove did not deny removing non-empty dir: %v", err)
			return
		}

		if err := st.Remove("/dir", true); err != nil {
			t.Errorf("Could not remove /dir recursively: %v", err)
			return
		}

		if err := st.Remove("/empty_dir", false); err != nil {
			t.Errorf("Could not remove /empty_dir non-recursively: %v", err)
			return
		}
	})
}
