package db

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/disorganizer/brig/util"
)

// DiskDatabase is a database that simply uses the filesystem as storage.
// Each bucket is one directory. Leaf keys are simple files.
// The exported form of the database is simply a gzipped .tar of the directory.
//
// Note that this database backends was written for easy debugging.
// It is currently by no means optimized for fast reads and writes and
// could be probably made a lot faster if we ever need that.
type DiskDatabase struct {
	basePath string
	lock     sync.RWMutex
	cache    map[string][]byte
	ops      []func() error
	refs     int64
}

// NewDiskDatabase creates a new database at `basePath`.
func NewDiskDatabase(basePath string) (*DiskDatabase, error) {
	return &DiskDatabase{
		basePath: basePath,
		cache:    make(map[string][]byte),
	}, nil
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

func (db *DiskDatabase) Flush() error {
	db.lock.Lock()
	defer db.lock.Unlock()

	// Currently no revertible operations are implemented. If something goes
	// wrong on the filesystem, chances are high that we're not able to revert
	// previous ops anyways.
	for _, op := range db.ops {
		if err := op(); err != nil {
			return err
		}
	}

	db.refs--
	db.cache = make(map[string][]byte)
	return nil
}

// Get a single value from `bucket` by `key`.
func (db *DiskDatabase) Get(key ...string) ([]byte, error) {
	db.lock.Lock()
	defer db.lock.Unlock()

	data, ok := db.cache[path.Join()]
	if ok {
		return data, nil
	}

	// We have to go to the disk to find the right key:
	filePath := filepath.Join(db.basePath, fixDirectoryKeys(key))
	data, err := ioutil.ReadFile(filePath)

	if os.IsNotExist(err) {
		return nil, ErrNoSuchKey
	}

	return data, err
}

func (db *DiskDatabase) Batch() Batch {
	db.lock.Lock()
	defer db.lock.Unlock()

	db.refs++
	return db
}

// Put stores a new `val` under `key` at `bucket`.
// Implementation detail: `key` may contain slashes (/). If used, those keys
// will result in a nested directory structure.
func (db *DiskDatabase) Put(val []byte, key ...string) {
	db.lock.Lock()
	defer db.lock.Unlock()

	db.ops = append(db.ops, func() error {
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
	})

	db.cache[path.Join(key...)] = val
}

// Clear removes all keys below and including `key`.
func (db *DiskDatabase) Clear(key ...string) {
	db.lock.Lock()
	defer db.lock.Unlock()

	// Cache the real modification for later:
	db.ops = append(db.ops, func() error {
		prefix := filepath.Join(db.basePath, fixDirectoryKeys(key))
		walker := func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() && strings.HasPrefix(path, prefix) {
				if err := os.Remove(path); err != nil {
					return err
				}
			}

			return nil
		}
		return filepath.Walk(db.basePath, walker)
	})

	prefix := path.Join(key...)
	for key := range db.cache {
		if strings.HasPrefix(key, prefix) {
			delete(db.cache, key)
		}
	}
}

func (db *DiskDatabase) Erase(key ...string) {
	db.lock.Lock()
	defer db.lock.Unlock()

	db.ops = append(db.ops, func() error {
		fullPath := filepath.Join(db.basePath, fixDirectoryKeys(key))
		err := os.Remove(fullPath)
		fmt.Println("ERASE", fullPath)
		if os.IsNotExist(err) {
			return ErrNoSuchKey
		}

		return err
	})

	delete(db.cache, path.Join(key...))
}

func (db *DiskDatabase) Keys(fn func(key []string) error, prefix ...string) error {
	db.lock.Lock()
	defer db.lock.Unlock()

	fullPath := filepath.Join(db.basePath, fixDirectoryKeys(prefix))
	return filepath.Walk(fullPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			if fn(reverseDirectoryKeys(path[len(db.basePath):])); err != nil {
				return err
			}
		}
		return nil
	})
}

// Export writes all key/valeus into a gzipped .tar that is written to `w`.
func (db *DiskDatabase) Export(w io.Writer) error {
	db.lock.Lock()
	defer db.lock.Unlock()

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
func (db *DiskDatabase) Import(r io.Reader) error {
	db.lock.Lock()
	defer db.lock.Unlock()

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

// Close the database
func (db *DiskDatabase) Close() error {
	return nil
}
