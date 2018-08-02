package db

import (
	"encoding/gob"
	"io"
	"path"
	"sort"
	"strings"
)

// MemoryDatabase is a purely in memory database.
type MemoryDatabase struct {
	data       map[string][]byte
	haveWrites bool
	refCount   int
}

// NewMemoryDatabase allocates a new empty MemoryDatabase
func NewMemoryDatabase() *MemoryDatabase {
	return &MemoryDatabase{
		data: make(map[string][]byte),
	}
}

// Batch is a no-op for a memory database.
func (mdb *MemoryDatabase) Batch() Batch {
	mdb.refCount++
	return mdb
}

// Flush is a no-op for a memory database.
func (mdb *MemoryDatabase) Flush() error {
	mdb.refCount--

	if mdb.refCount == 0 {
		mdb.haveWrites = false
	}
	return nil
}

// Rollback is a no-op for a memory database
func (mdb *MemoryDatabase) Rollback() {}

// Get returns `key` of `bucket`.
func (mdb *MemoryDatabase) Get(key ...string) ([]byte, error) {
	data, ok := mdb.data[path.Join(key...)]
	if !ok {
		return nil, ErrNoSuchKey
	}

	return data, nil
}

// Put sets `key` in `bucket` to `data`.
func (mdb *MemoryDatabase) Put(data []byte, key ...string) {
	mdb.haveWrites = true
	mdb.data[path.Join(key...)] = data
}

// Clear removes all keys includin and below `key`.
func (mdb *MemoryDatabase) Clear(key ...string) {
	mdb.haveWrites = true
	joinedKey := path.Join(key...)
	for mapKey := range mdb.data {
		if strings.HasPrefix(mapKey, joinedKey) {
			delete(mdb.data, mapKey)
		}
	}
}

func (mdb *MemoryDatabase) Erase(key ...string) {
	fullKey := path.Join(key...)
	mdb.haveWrites = true
	delete(mdb.data, fullKey)
}

// Keys will return all keys currently stored in the memory map
func (mdb *MemoryDatabase) Keys(fn func(key []string) error, prefix ...string) error {
	prefixPath := path.Join(prefix...)
	for key := range mdb.data {
		if strings.HasPrefix(key, prefixPath) {
			if err := fn(strings.Split(key, "/")); err != nil {
				return err
			}
		}
	}

	return nil
}

func (mdb *MemoryDatabase) HaveWrites() bool {
	return mdb.haveWrites
}

func (mdb *MemoryDatabase) Glob(prefix []string) ([][]string, error) {
	prefixKey := path.Join(prefix...)

	var result [][]string

	err := mdb.Keys(func(key []string) error {
		fullKey := path.Join(key...)
		if strings.HasPrefix(fullKey, prefixKey) {
			// Filter "directories":
			suffix := fullKey[len(prefixKey):]
			if !strings.Contains(suffix, "/") {
				result = append(result, strings.Split(fullKey, "/"))
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	sort.Slice(result, func(i, j int) bool {
		return path.Join(result[i]...) < path.Join(result[j]...)
	})

	return result, nil
}

// Export encodes the internal memory map to a gob structure,
// and writes it to `w`.
func (mdb *MemoryDatabase) Export(w io.Writer) error {
	return gob.NewEncoder(w).Encode(mdb.data)
}

// Import imports a previously exported dump and decodes the gob structure.
func (mdb *MemoryDatabase) Import(r io.Reader) error {
	return gob.NewDecoder(r).Decode(&mdb.data)
}

// Close the memory - a no op.
func (mdb *MemoryDatabase) Close() error {
	return nil
}
