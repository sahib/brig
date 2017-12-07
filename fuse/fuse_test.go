package fuse

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/sahib/brig/catfs"
	"github.com/sahib/brig/util/testutil"
)

func init() {
	// NOTE: This is useful for debugging.
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

	fs, err := catfs.NewFilesystem(backend, dbPath, owner, nil)
	if err != nil {
		t.Fatalf("Failed to create filesystem: %v", err)
	}

	fn(fs)

	if err := fs.Close(); err != nil {
		t.Fatalf("Failed to close filesystem: %v", err)
	}
}

func withMount(t *testing.T, f func(mount *Mount)) {
	mntPath := filepath.Join(os.TempDir(), "brig-fuse-mountdir")

	if err := os.MkdirAll(mntPath, 0777); err != nil {
		t.Fatalf("Unable to create empty mount dir: %v", err)
	}

	defer testutil.Remover(t, mntPath)

	withDummyFS(t, func(fs *catfs.FS) {
		mount, err := NewMount(fs, mntPath)
		if err != nil {
			t.Fatalf("Cannot create mount: %v", err)
		}

		f(mount)

		if err := mount.Close(); err != nil {
			t.Fatalf("Closing mount failed: %v", err)
		}
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
		t.Errorf("Data from simple file does not match source. Len: %d", len(data))
		t.Errorf("\tExpected: %v", data)
		t.Fatalf("\tGot:      %v", helloBuffer.Bytes())
	}
}

var (
	DataSizes = []int64{
		0, 1, 2, 4, 8, 16, 32, 64, 1024,
		2048, 4095, 4096, 4097, 147611,
	}
)

func TestRead(t *testing.T) {
	withMount(t, func(mount *Mount) {
		for _, size := range DataSizes {
			helloData := testutil.CreateDummyBuf(size)

			// Add a simple file:
			name := fmt.Sprintf("hello_%d", size)
			reader := bytes.NewReader(helloData)
			if err := mount.filesys.cfs.Stage("/"+name, reader); err != nil {
				t.Fatalf("Adding simple file from reader failed: %v", err)
			}

			path := filepath.Join(mount.Dir, name)
			checkForCorrectFile(t, path, helloData)
		}
	})
}

func TestWrite(t *testing.T) {
	withMount(t, func(mount *Mount) {
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
	withMount(t, func(mount *Mount) {
		for _, size := range DataSizes {
			name := fmt.Sprintf("/empty_%d", size)
			if err := mount.filesys.cfs.Touch(name); err != nil {
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
