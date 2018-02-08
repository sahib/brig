package db

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/sahib/brig/util"
)

const (
	debug = false
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
	db.refs--
	if db.refs < 0 {
		db.refs = 0
	}

	if db.refs > 0 {
		return nil
	}

	if debug {
		fmt.Println("FLUSH")
	}

	// Clear the cache first, if any of the next step fail,
	// we have at least the current state.
	db.cache = make(map[string][]byte)

	// Make sure that db.ops is nil, even if Flush failed.
	ops := db.ops
	db.ops = nil

	// Currently no revertible operations are implemented. If something goes
	// wrong on the filesystem, chances are high that we're not able to revert
	// previous ops anyways.
	for _, op := range ops {
		if err := op(); err != nil {
			return err
		}
	}

	return nil
}

func (db *DiskDatabase) Rollback() {
	if debug {
		fmt.Println("ROLLBACK")
	}

	db.refs = 0
	db.ops = nil
	db.cache = make(map[string][]byte)
}

// Get a single value from `bucket` by `key`.
func (db *DiskDatabase) Get(key ...string) ([]byte, error) {
	if debug {
		fmt.Println("GET", key)
	}

	data, ok := db.cache[path.Join(key...)]
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
	db.refs++
	return db
}

func removeNonDirs(path string) error {
	if path == "/" || path == "" {
		return nil
	}

	info, err := os.Stat(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if info != nil && !info.IsDir() {
		return os.Remove(path)
	}

	return removeNonDirs(filepath.Dir(path))
}

// Put stores a new `val` under `key` at `bucket`.
// Implementation detail: `key` may contain slashes (/). If used, those keys
// will result in a nested directory structure.
func (db *DiskDatabase) Put(val []byte, key ...string) {
	if debug {
		fmt.Println("SET", key)
	}

	db.ops = append(db.ops, func() error {
		filePath := filepath.Join(db.basePath, fixDirectoryKeys(key))

		// If any of the parent are non-directories,
		// we need to remove them, since more nesting is requested.
		// (e.g. set /a/b/c/d over /a/b/c, where c is a file)
		parentDir := filepath.Dir(filePath)
		if err := removeNonDirs(parentDir); err != nil {
			return err
		}

		if err := os.MkdirAll(parentDir, 0700); err != nil {
			return err
		}

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
	if debug {
		fmt.Println("CLEAR", key)
	}

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

	// Make sure we also modify the currently cached objects:
	prefix := path.Join(key...)
	for key := range db.cache {
		if strings.HasPrefix(key, prefix) {
			delete(db.cache, key)
		}
	}
}

func (db *DiskDatabase) Erase(key ...string) {
	if debug {
		fmt.Println("ERASE", key)
	}

	db.ops = append(db.ops, func() error {
		fullPath := filepath.Join(db.basePath, fixDirectoryKeys(key))
		err := os.Remove(fullPath)
		if os.IsNotExist(err) {
			return ErrNoSuchKey
		}

		return err
	})

	delete(db.cache, path.Join(key...))
}

func (db *DiskDatabase) Keys(fn func(key []string) error, prefix ...string) error {
	fullPath := filepath.Join(db.basePath, fixDirectoryKeys(prefix))
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil
	}

	// TODO: error reporting is correct here?
	return filepath.Walk(fullPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			if err := fn(reverseDirectoryKeys(path[len(db.basePath):])); err != nil {
				return err
			}
		}

		return nil
	})
}

func (db *DiskDatabase) Glob(prefix []string) ([][]string, error) {
	fullPrefix := filepath.Join(db.basePath, filepath.Join(prefix...))
	matches, err := filepath.Glob(fullPrefix + "*")
	if err != nil {
		return nil, err
	}

	results := [][]string{}
	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil {
			return nil, err
		}

		if !info.IsDir() {
			result := strings.Split(match[len(db.basePath)+1:], string(filepath.Separator))
			results = append(results, result)
		}
	}

	return results, nil
}

// Export writes all key/values into a gzipped .tar that is written to `w`.
func (db *DiskDatabase) Export(w io.Writer) error {
	archiveName := fmt.Sprintf("brigmeta-%s.gz", time.Now().Format(time.RFC3339))
	return util.Tar(db.basePath, archiveName, w)
}

// Import a gzipped tar from `r` into the current database.
func (db *DiskDatabase) Import(r io.Reader) error {
	return util.Untar(r, db.basePath)
}

// Close the database
func (db *DiskDatabase) Close() error {
	return nil
}
