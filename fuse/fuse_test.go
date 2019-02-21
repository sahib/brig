// +build !windows

package fuse

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/sahib/brig/catfs"
	"github.com/sahib/brig/defaults"
	"github.com/sahib/brig/util/testutil"
	"github.com/sahib/config"
	"github.com/stretchr/testify/require"
)

func init() {
	log.SetLevel(log.DebugLevel)
}

func withDummyFS(t *testing.T, fn func(fs *catfs.FS)) {
	backend := catfs.NewMemFsBackend()
	owner := "alice"

	dbPath, err := ioutil.TempDir("", "brig-fs-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	defer os.RemoveAll(dbPath)

	cfg, err := config.Open(nil, defaults.Defaults, config.StrictnessPanic)
	require.Nil(t, err)

	fs, err := catfs.NewFilesystem(backend, dbPath, owner, false, cfg.Section("fs"))
	if err != nil {
		t.Fatalf("Failed to create filesystem: %v", err)
	}

	fn(fs)

	if err := fs.Close(); err != nil {
		t.Fatalf("Failed to close filesystem: %v", err)
	}
}

func withMountFromFs(t *testing.T, opts MountOptions, fs *catfs.FS, f func(mount *Mount)) {
	mntPath := filepath.Join(os.TempDir(), "brig-fuse-mountdir")

	// Make sure to unmount any mounts that are there.
	// Possibly there are some leftovers from previous failed runs.
	lazyUnmount(mntPath)

	if err := os.MkdirAll(mntPath, 0777); err != nil {
		t.Fatalf("Unable to create empty mount dir: %v", err)
	}

	defer testutil.Remover(t, mntPath)

	mount, err := NewMount(fs, mntPath, nil, opts)
	if err != nil {
		t.Fatalf("Cannot create mount: %v", err)
	}

	f(mount)

	if err := mount.Close(); err != nil {
		t.Fatalf("Closing mount failed: %v", err)
	}
}

func withMount(t *testing.T, opts MountOptions, f func(mount *Mount)) {
	withDummyFS(t, func(fs *catfs.FS) {
		withMountFromFs(t, opts, fs, f)
	})

}

func checkForCorrectFile(t *testing.T, path string, data []byte) {
	// Try to read it over fuse:
	helloBuffer := &bytes.Buffer{}
	fd, err := os.Open(path)
	if err != nil {
		t.Fatalf("Unable to open simple file over fuse: %v", err)
	}

	defer func() {
		if err := fd.Close(); err != nil {
			t.Fatalf("Unable to close simple file over fuse: %v", err)
		}
	}()

	n, err := io.CopyBuffer(helloBuffer, fd, make([]byte, 128*1024))
	if err != nil {
		t.Fatalf("Unable to read full simple file over fuse: %v", err)
	}

	if n != int64(len(data)) {
		t.Fatalf("Data differs over fuse: got %d, should be %d bytes", n, len(data))
	}

	if !bytes.Equal(helloBuffer.Bytes(), data) {
		t.Errorf(
			"Data from simple file does not match source. Len: %d %d",
			len(data),
			helloBuffer.Len(),
		)

		require.Equal(t, data, helloBuffer.Bytes())
	}
}

var (
	DataSizes = []int64{
		0, 1, 2, 4, 8, 16, 32, 64, 1024,
		2048, 4095, 4096, 4097, 147611,
	}
)

func TestRead(t *testing.T) {
	withMount(t, MountOptions{}, func(mount *Mount) {
		for _, size := range DataSizes {
			t.Run(fmt.Sprintf("%d", size), func(t *testing.T) {
				helloData := testutil.CreateDummyBuf(size)

				// Add a simple file:
				name := fmt.Sprintf("hello_%d", size)
				reader := bytes.NewReader(helloData)
				if err := mount.filesys.m.fs.Stage("/"+name, reader); err != nil {
					t.Fatalf("Adding simple file from reader failed: %v", err)
				}

				path := filepath.Join(mount.Dir, name)
				checkForCorrectFile(t, path, helloData)
			})
		}
	})
}

func TestWrite(t *testing.T) {
	withMount(t, MountOptions{}, func(mount *Mount) {
		for _, size := range DataSizes {
			helloData := testutil.CreateDummyBuf(size)
			path := filepath.Join(mount.Dir, fmt.Sprintf("hello_%d", size))

			// Write a simple file via the fuse layer:
			err := ioutil.WriteFile(path, helloData, 0644)
			if err != nil {
				t.Fatalf("Could not write simple file via fuse layer: %v", err)
			}

			checkForCorrectFile(t, path, helloData)
		}
	})
}

// Regression test for copying larger file to the mount.
func TestTouchWrite(t *testing.T) {
	withMount(t, MountOptions{}, func(mount *Mount) {
		for _, size := range DataSizes {
			name := fmt.Sprintf("/empty_%d", size)
			if err := mount.filesys.m.fs.Touch(name); err != nil {
				t.Fatalf("Could not touch an empty file: %v", err)
			}

			path := filepath.Join(mount.Dir, name)

			// Write a simple file via the fuse layer:
			helloData := testutil.CreateDummyBuf(size)
			err := ioutil.WriteFile(path, helloData, 0644)
			if err != nil {
				t.Fatalf("Could not write simple file via fuse layer: %v", err)
			}

			checkForCorrectFile(t, path, helloData)
		}
	})
}

// Regression test for copying a file to a subdirectory.
func TestTouchWriteSubdir(t *testing.T) {
	withMount(t, MountOptions{}, func(mount *Mount) {
		subDirPath := filepath.Join(mount.Dir, "sub")
		require.Nil(t, os.Mkdir(subDirPath, 0644))

		expected := []byte{1, 2, 3}
		filePath := filepath.Join(subDirPath, "donald.png")
		require.Nil(t, ioutil.WriteFile(filePath, expected, 0644))

		got, err := ioutil.ReadFile(filePath)
		require.Nil(t, err)
		require.Equal(t, expected, got)
	})
}

func TestReadOnlyFs(t *testing.T) {
	opts := MountOptions{
		ReadOnly: true,
	}
	withMount(t, opts, func(mount *Mount) {
		cfs := mount.filesys.m.fs
		cfs.Stage("/x.png", bytes.NewReader([]byte{1, 2, 3}))

		// Do some allowed io to check if the fs is actually working.
		// The test does not check on the kind of errors otherwise.
		xPath := filepath.Join(mount.Dir, "x.png")
		data, err := ioutil.ReadFile(xPath)
		require.Nil(t, err)
		require.Equal(t, []byte{1, 2, 3}, data)

		// Try creating a new file:
		yPath := filepath.Join(mount.Dir, "y.png")
		require.NotNil(t, ioutil.WriteFile(yPath, []byte{4, 5, 6}, 0600))

		// Try modifying an existing file:
		require.NotNil(t, ioutil.WriteFile(xPath, []byte{4, 5, 6}, 0600))

		dirPath := filepath.Join(mount.Dir, "sub")
		require.NotNil(t, os.Mkdir(dirPath, 0644))
	})
}

func TestWithRoot(t *testing.T) {
	opts := MountOptions{
		Root: "/a/b",
	}

	withDummyFS(t, func(fs *catfs.FS) {
		require.Nil(t, fs.Mkdir("/a/b", true))
		require.Nil(t, fs.Mkdir("/a/b/c", true))
		require.Nil(t, fs.Stage("/u.png", bytes.NewReader([]byte{1, 2, 3})))
		require.Nil(t, fs.Stage("/a/x.png", bytes.NewReader([]byte{2, 3, 4})))
		require.Nil(t, fs.Stage("/a/b/y.png", bytes.NewReader([]byte{3, 4, 5})))
		require.Nil(t, fs.Stage("/a/b/c/z.png", bytes.NewReader([]byte{4, 5, 6})))

		withMountFromFs(t, opts, fs, func(mount *Mount) {
			yPath := filepath.Join(mount.Dir, "y.png")
			data, err := ioutil.ReadFile(yPath)
			require.Nil(t, err)
			require.Equal(t, []byte{3, 4, 5}, data)

			newPath := filepath.Join(mount.Dir, "new.png")
			require.Nil(t, ioutil.WriteFile(newPath, []byte{5, 6, 7}, 0644))

			data, err = ioutil.ReadFile(newPath)
			require.Nil(t, err)
			require.Equal(t, []byte{5, 6, 7}, data)
		})
	})
}
