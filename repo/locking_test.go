package repo

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/disorganizer/brig/util/testutil"
	"github.com/stretchr/testify/require"
)

func mustCreate(t *testing.T, path string, size int64) []byte {
	if err := os.MkdirAll(filepath.Dir(path), 0744); err != nil {
		t.Fatalf("Failed to create necessary directories for %v: %v", path, err)
	}

	fd, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("Failed to touch `%s`: %v", path, err)
	}

	buf := testutil.CreateDummyBuf(size)
	if _, err := fd.Write(buf); err != nil {
		t.Fatalf("Failed to write dummy data to %s: %v", path, err)
	}

	if err := fd.Close(); err != nil {
		t.Fatalf("Failed to close fd for `%s`: %v", path, err)
	}
	return buf
}

func withTempDir(t *testing.T, fn func(dir string)) {
	dir, err := ioutil.TempDir("", "brig-repo-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	fn(dir)

	if err := os.RemoveAll(dir); err != nil {
		t.Fatalf("Failed to remove test directory at `%s`: %v", dir, err)
	}
}

func TestLockFile(t *testing.T) {
	withTempDir(t, func(dir string) {
		xData := mustCreate(t, filepath.Join(dir, "x"), 0)
		yData := mustCreate(t, filepath.Join(dir, "y"), 1024)
		zData := mustCreate(t, filepath.Join(dir, "z"), 1024*1024)

		// Do not lock this file:
		mustCreate(t, filepath.Join(dir, "meta.yml"), 1024)

		subDir := filepath.Join(dir, "sub")
		if err := os.Mkdir(subDir, 0744); err != nil {
			t.Fatalf("Creating test sub dir failed: %v", err)
		}

		aData := mustCreate(t, filepath.Join(subDir, "a"), 2*1024*1024)

		// Should be overwritten.
		mustCreate(t, filepath.Join(dir, "z.locked"), 2)

		if err := LockRepo(dir, "karl", "klaus", []string{"meta.yml"}); err != nil {
			t.Fatalf("Failed to lock directory: %v", err)
		}

		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if path == dir || strings.HasSuffix(path, "meta.yml") {
				return nil
			}

			if !strings.HasSuffix(path, LockPathSuffix) && path != dir {
				t.Fatalf("%s has not been locked", path)
			}

			return nil
		})

		if err != nil {
			t.Fatalf("walk failed: %v", err)
		}

		// Try with a wrong password:
		if err := UnlockRepo(dir, "karl", "klausi"); err == nil {
			t.Fatalf("unlock worked without correct password")
		}

		// Try with a wrong user:
		if err := UnlockRepo(dir, "karol", "klaus"); err == nil {
			t.Fatalf("unlock worked without correct user")
		}

		if err := UnlockRepo(dir, "karl", "klaus"); err != nil {
			t.Fatalf("unlock failed: %v", err)
		}

		err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if strings.HasSuffix(path, LockPathSuffix) && path != dir {
				t.Fatalf("%s is still locked", path)
			}

			cmp := func(expected []byte) {
				data, err := ioutil.ReadFile(path)
				if err != nil {
					t.Fatalf("Failed to read file %s: %v", path, err)
				}

				if !bytes.Equal(expected, data) {
					t.Fatalf("Content differs after lock & unlock: %s", path)
				}
			}

			switch filepath.Base(path) {
			case "a":
				cmp(aData)
			case "x":
				cmp(xData)
			case "y":
				cmp(yData)
			case "z":
				cmp(zData)
			}

			return nil
		})

		if err != nil {
			t.Fatalf("walk after unlock failed: %v", err)
		}

	})
}

func TestLockDirectory(t *testing.T) {
	withTempDir(t, func(dir string) {
		mustCreate(t, filepath.Join(dir, "a/b/x"), 1024)
		mustCreate(t, filepath.Join(dir, "a/b/empty_file"), 0)
		mustCreate(t, filepath.Join(dir, "a/x"), 1024)
		mustCreate(t, filepath.Join(dir, "x"), 1024)

		require.Nil(t, os.MkdirAll(filepath.Join(dir, "empty"), 0744))

		key := make([]byte, 32)
		require.Nil(t, lockDirectory(dir, key))
		require.Nil(t, os.RemoveAll(dir))
		require.Nil(t, unlockDirectory(dir+".tgz.locked", key))

		paths := []string{}
		walker := func(path string, child os.FileInfo, err error) error {
			require.Nil(t, err)
			path = path[len(dir):]
			if len(path) > 0 {
				paths = append(paths, path)
			}

			return nil
		}

		require.Nil(t, filepath.Walk(dir, walker))

		sort.Strings(paths)

		// Note that /empty vanished after untarring.
		// For now we do not support tarring empty directories.
		require.Equal(
			t,
			[]string{
				"/a",
				"/a/b",
				"/a/b/empty_file",
				"/a/b/x",
				"/a/x",
				"/x",
			},
			paths,
		)
	})
}
