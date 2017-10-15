package repo

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func mustTouch(t *testing.T, path string) {
	fd, err := os.OpenFile(path, os.O_CREATE, 0644)
	if err != nil {
		t.Fatalf("Failed to touch `%s`: %v", path, err)
	}

	if err := fd.Close(); err != nil {
		t.Fatalf("Failed to close fd for `%s`: %v", path, err)
	}
}

func withTempDir(t *testing.T, fn func(dir string)) {
	dir, err := ioutil.TempDir("", "brig-repo-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	fn(dir)

	// if err := os.RemoveAll(dir); err != nil {
	// 	t.Fatalf("Failed to remove test directory at `%s`: %v", dir, err)
	// }
}

func TestLockFile(t *testing.T) {
	withTempDir(t, func(dir string) {
		mustTouch(t, filepath.Join(dir, "x"))
		mustTouch(t, filepath.Join(dir, "y"))
		mustTouch(t, filepath.Join(dir, "z"))

		// Should be overwritten.
		mustTouch(t, filepath.Join(dir, "z.locked"))

		if err := Lock(dir, "karl", "klaus", nil); err != nil {
			t.Fatalf("Failed to lock directory: %v", err)
		}

		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			fmt.Println(path, err)
			return nil
		})

		if err != nil {
			t.Fatalf("walk failed: %v", err)
		}

		fmt.Println(Unlock(dir, "karl", "klaus"))
	})
}
