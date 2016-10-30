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
	"github.com/disorganizer/brig/repo"
	"github.com/disorganizer/brig/store/compress"
	"github.com/disorganizer/brig/util/testutil"
)

func withMount(t *testing.T, f func(mount *Mount)) {
	mntPath := filepath.Join(os.TempDir(), "brig_fuse_mountdir")

	// NOTE: This is useful for debugging.
	log.SetLevel(log.WarnLevel)
	// log.SetLevel(log.DebugLevel)

	if err := os.MkdirAll(mntPath, 0777); err != nil {
		t.Errorf("Unable to create empty mount dir: %v", err)
		return
	}

	defer testutil.Remover(t, mntPath)

	repo.WithAliceRepo(t, func(rep *repo.Repository) {
		mount, err := NewMount(rep.OwnStore, mntPath)
		if err != nil {
			t.Errorf("Cannot create mount: %v", err)
			return
		}

		f(mount)

		if err := mount.Close(); err != nil {
			t.Errorf("Closing mount failed: %v", err)
		}
	})
}

func checkForCorrectFile(t *testing.T, path string, data []byte) bool {
	// Try to read it over fuse:
	helloBuffer := &bytes.Buffer{}
	fd, err := os.Open(path)
	if err != nil {
		t.Errorf("Unable to open simple file over fuse: %v", err)
		return false
	}

	defer func() {
		if err := fd.Close(); err != nil {
			t.Errorf("Unable to close simple file over fuse: %v", err)
		}
	}()

	n, err := io.CopyBuffer(helloBuffer, fd, make([]byte, 128*1024))
	if err != nil {
		t.Errorf("Unable to read full simple file over fuse: %v", err)
		return false
	}

	if n != int64(len(data)) {
		t.Errorf("Data differs over fuse: got %d, should be %d bytes", n, len(data))
		return false
	}

	if !bytes.Equal(helloBuffer.Bytes(), data) {
		t.Errorf("Data from simple file does not match source. Len: %d", len(data))
		t.Errorf("\tExpected: %v", data)
		t.Errorf("\tGot:      %v", helloBuffer.Bytes())
		return false
	}

	return true
}

var (
	DataSizes = []int64{
		// 0, 1, 2, 4, 8, 16, 32, 64, 1024, 2048, 4095, 4096, 4097,
		147611,
	}
)

func TestRead(t *testing.T) {
	withMount(t, func(mount *Mount) {
		for _, size := range DataSizes {
			helloData := testutil.CreateDummyBuf(size)

			// Add a simple file:
			name := fmt.Sprintf("hello_%d", size)
			reader := bytes.NewReader(helloData)
			if err := mount.Store.StageFromReader("/"+name, reader, compress.AlgoNone); err != nil {
				t.Errorf("Adding simple file from reader failed: %v", err)
				return
			}

			path := filepath.Join(mount.Dir, name)
			if !checkForCorrectFile(t, path, helloData) {
				break
			}
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
				t.Errorf("Could not write simple file via fuse layer: %v", err)
				return
			}

			if !checkForCorrectFile(t, path, helloData) {
				break
			}
		}
	})
}

// Regression test for copying larger file to the mount.
func TestTouchWrite(t *testing.T) {
	withMount(t, func(mount *Mount) {
		for _, size := range DataSizes {
			name := fmt.Sprintf("/empty_%d", size)
			if err := mount.Store.Touch(name); err != nil {
				t.Errorf("Could not touch an empty file: %v", err)
				return
			}

			path := filepath.Join(mount.Dir, name)

			// Write a simple file via the fuse layer:
			helloData := testutil.CreateDummyBuf(size)
			err := ioutil.WriteFile(path, helloData, 0644)
			if err != nil {
				t.Errorf("Could not write simple file via fuse layer: %v", err)
				return
			}

			if !checkForCorrectFile(t, path, helloData) {
				break
			}
		}
	})
}
