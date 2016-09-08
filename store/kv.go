package store

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/boltdb/bolt"
)

var (
	ErrEmptyPath    = errors.New("Bucket path is empty")
	ErrPathTooShort = errors.New("Path needs at least two elements for this")
)

type KV interface {
	Bucket(path []string) (Bucket, error)
	Close() error
}

type Bucket interface {
	Get(key string) ([]byte, error)
	Put(key string, data []byte) error
	Bucket(path []string) (Bucket, error)

	// TODO:
	// Clear() error
	// CopyTo(b Bucket) error
}

func findBucket(kv KV, path string) (Bucket, string, error) {
	elems := strings.Split(path, "/")
	if len(elems) == 0 {
		return nil, "", ErrEmptyPath
	}

	if elems[0] == "" {
		elems = elems[1:]
	}

	if len(elems) == 0 {
		return nil, "", ErrPathTooShort
	}

	// Get the parent bucket:
	bkt, err := kv.Bucket(elems[:len(elems)-1])
	if err != nil {
		return nil, "", err
	}

	return bkt, elems[len(elems)-1], nil
}

func getPath(kv KV, path string) ([]byte, error) {
	bkt, key, err := findBucket(kv, path)
	if err != nil {
		return nil, err
	}

	data, err := bkt.Get(key)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func putPath(kv KV, path string, data []byte) error {
	bkt, key, err := findBucket(kv, path)
	fmt.Println("putPath", path, bkt, err, key, data)
	if err != nil {
		return err
	}

	return bkt.Put(key, data)
}

//////////////////////////////////

type BoltKV struct {
	db *bolt.DB
}

func NewBoltKV(path string) (KV, error) {
	options := &bolt.Options{Timeout: 1 * time.Second}

	if err := os.MkdirAll(path, 0777); err != nil {
		return nil, err
	}

	db, err := bolt.Open(filepath.Join(path, "index.bolt"), 0600, options)
	if err != nil {
		return nil, err
	}

	return &BoltKV{db}, nil
}

func (kv *BoltKV) Bucket(path []string) (Bucket, error) {
	if len(path) == 0 {
		return nil, fmt.Errorf("Empty path given to Bucket()")
	}

	err := kv.db.Update(func(tx *bolt.Tx) error {
		curr, err := tx.CreateBucketIfNotExists([]byte(path[0]))
		if err != nil {
			return err
		}

		for _, name := range path[1:] {
			curr, err = curr.CreateBucketIfNotExists([]byte(name))
			if err != nil {
				return err
			}
		}
		return err
	})

	if err != nil {
		return nil, err
	}

	return &BoltBucket{kv.db, path}, nil
}

func (kv *BoltKV) Close() error {
	return kv.db.Close()
}

type BoltBucket struct {
	db   *bolt.DB
	path []string
}

// dig to the correct bucket at bb.path
func (bb *BoltBucket) dig(path []string, writable bool, fn func(bucket *bolt.Bucket) error) error {
	if len(path) == 0 {
		return nil
	}

	tx, err := bb.db.Begin(writable)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	// Iterate down to the right bucket:
	buck := tx.Bucket([]byte(path[0]))
	for iter := path[1:]; len(iter) > 0; iter = iter[1:] {
		buck = buck.Bucket([]byte(iter[0]))
	}

	if err := fn(buck); err != nil {
		return err
	}

	if writable {
		// Commit the transaction and check for error.
		if err := tx.Commit(); err != nil {
			return err
		}
	}

	return nil
}

// TODO: Transactions? GetMany? PutMany?

func (bb *BoltBucket) Get(key string) (data []byte, err error) {
	err = bb.dig(bb.path, false, func(bucket *bolt.Bucket) error {
		bdata := bucket.Get([]byte(key))
		if bdata == nil {
			return nil
		}

		// We need to copy the data, since it's only valid for this transaction:
		data = make([]byte, len(bdata))
		copy(data, bdata)
		return nil
	})

	// Silence `data` if some error happened:
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (bb *BoltBucket) Put(key string, data []byte) error {
	return bb.dig(bb.path, true, func(bucket *bolt.Bucket) error {
		return bucket.Put([]byte(key), data)
	})
}

func (bb *BoltBucket) Bucket(path []string) (Bucket, error) {
	err := bb.dig(bb.path, true, func(bucket *bolt.Bucket) error {
		curr := bucket
		var err error

		for _, name := range path {
			curr, err = curr.CreateBucketIfNotExists([]byte(name))
			if err != nil {
				return err
			}
		}

		return nil
	})

	// Silence `data` if some error happened:
	if err != nil {
		return nil, err
	}

	return &BoltBucket{
		db:   bb.db,
		path: append(bb.path, path...),
	}, nil
}
