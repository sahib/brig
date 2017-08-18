package db

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/disorganizer/brig/util"
)

// DiskvDatabase is a database that simply uses the filesystem as storage.
// Each bucket is one directory. Leaf keys are simple files.
// The exported form of the database is simply a gzipped .tar of the directory.
// TODO: rename, doesn't use diskv anymore
type DiskvDatabase struct {
	basePath string
}

// NewDiskvDatabase creates a new database at `basePath`.
func NewDiskvDatabase(basePath string) (*DiskvDatabase, error) {
	return &DiskvDatabase{basePath: basePath}, nil
}

func fixDirectoryKeys(key []string) string {
	if len(key) == 0 {
		return ""
	}

	switch lastPart := key[len(key)-1]; {
	case lastPart == "DOT":
		return filepath.Join(key[:len(key)-1]...) + "/__NO_DOT__"
	case lastPart == "." || strings.HasSuffix(lastPart, "/."):
		return filepath.Join(key[:len(key)-1]...) + "/DOT"
	default:
		return filepath.Join(key...)
	}
}

func reverseDirectoryKeys(key string) []string {
	parts := strings.Split(key, string(filepath.Separator))
	switch parts[len(parts)-1] {
	case "DOT":
		parts[len(parts)-1] = "."
	case "__NO_DOT__":
		parts[len(parts)-1] = "DOT"
	}

	return parts
}

// Get a single value from `bucket` by `key`.
func (db *DiskvDatabase) Get(key ...string) ([]byte, error) {
	filePath := filepath.Join(db.basePath, fixDirectoryKeys(key))
	data, err := ioutil.ReadFile(filePath)

	if os.IsNotExist(err) {
		return nil, ErrNoSuchKey
	}

	return data, err
}

// Put stores a new `val` under `key` at `bucket`.
// Implementation detail: `key` may contain slashes (/). If used, those keys
// will result in a nested directory structure.
func (db *DiskvDatabase) Put(val []byte, key ...string) error {
	filePath := filepath.Join(db.basePath, fixDirectoryKeys(key))
	if err := os.MkdirAll(filepath.Dir(filePath), 0700); err != nil {
		return err
	}

	fmt.Println("SET", key)
	// It is allowed to set a key over an existing one.
	// i.e. set "a/b" over "a/b/c". This requires us to potentially
	// delete nested directories (c).
	info, err := os.Stat(filePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if info != nil && info.IsDir() {
		if err := os.RemoveAll(filePath); err != nil {
			return err
		}
	}

	return ioutil.WriteFile(filePath, val, 0600)
}

// Clear removes all keys below and including `key`.
func (db *DiskvDatabase) Clear(key ...string) error {
	prefix := filepath.Join(db.basePath, fixDirectoryKeys(key))
	return filepath.Walk(db.basePath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && strings.HasPrefix(path, prefix) {
			if err := os.Remove(path); err != nil {
				return err
			}
		}

		return nil
	})
}

// Export writes all key/valeus into a gzipped .tar that is written to `w`.
func (db *DiskvDatabase) Export(w io.Writer) error {
	gzw := gzip.NewWriter(w)
	gzw.Name = fmt.Sprintf("brigmeta-%s.gz", time.Now().Format(time.RFC3339))
	gzw.Comment = "compressed brig metadata database"
	gzw.ModTime = time.Now()

	tw := tar.NewWriter(gzw)
	walker := func(path string, info os.FileInfo, err error) error {
		if !info.Mode().IsRegular() {
			return nil
		}

		hdr := &tar.Header{
			Name: path[len(db.basePath):],
			Mode: 0600,
			Size: info.Size(),
		}

		if werr := tw.WriteHeader(hdr); err != nil {
			return werr
		}

		fd, err := os.Open(path)
		if err != nil {
			return err
		}

		defer util.Closer(fd)

		if _, err := io.Copy(tw, fd); err != nil {
			return err
		}

		return nil
	}

	if err := filepath.Walk(db.basePath, walker); err != nil {
		return err
	}

	if err := tw.Close(); err != nil {
		return err
	}

	return gzw.Close()
}

// Import a gzipped tar from `r` into the current database.
func (db *DiskvDatabase) Import(r io.Reader) error {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}

	tr := tar.NewReader(gzr)
	for {
		hdr, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		// Create the necessary directory if necessary.
		fullPath := filepath.Join(db.basePath, hdr.Name)
		if oerr := os.MkdirAll(filepath.Dir(fullPath), 0700); err != nil {
			return oerr
		}

		// Overwrite the file in the target directory
		fd, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
		if err != nil {
			return err
		}

		if _, err := io.Copy(fd, tr); err != nil {
			return fd.Close()
		}

		if err := fd.Close(); err != nil {
			return err
		}
	}

	return gzr.Close()
}

func (db *DiskvDatabase) Keys(prefix ...string) (<-chan []string, error) {
	ch := make(chan []string)
	fullPath := filepath.Join(db.basePath, fixDirectoryKeys(prefix))

	go func() {
		filepath.Walk(fullPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() {
				ch <- reverseDirectoryKeys(path[len(db.basePath):])
			}
			return nil
		})

		close(ch)
	}()

	return ch, nil
}

func (db *DiskvDatabase) Erase(key ...string) error {
	fullPath := filepath.Join(db.basePath, fixDirectoryKeys(key))
	err := os.Remove(fullPath)
	fmt.Println("ERASE", fullPath)
	if os.IsNotExist(err) {
		return ErrNoSuchKey
	}

	return err
}

// Close the database
func (db *DiskvDatabase) Close() error {
	return nil
}
