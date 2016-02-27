package store

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/disorganizer/brig/util/testwith"
)

var TestPath = filepath.Join(os.TempDir(), "brig-store-test")

func withEmptyStore(t *testing.T, f func(*Store)) {
	ipfsPath := filepath.Join(TestPath, "ipfs")

	testwith.WithIpfsRepo(t, ipfsPath, func(ipfsRepoPath string) {
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
		store, err := Open(TestPath)
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
	fmt.Println(string(exportData))

	withEmptyStore(t, func(store *Store) {
		if err := store.Import(TestPath, bytes.NewReader(exportData)); err != nil {
			t.Errorf("Could not import data: %v", err)
			return
		}
	})
}
