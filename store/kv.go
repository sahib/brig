package store

import (
	"errors"
	"fmt"
	"io"
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
	Export(w io.Writer) error
	Import(r io.Reader) error
	Close() error
}

type Bucket interface {
	Get(key string) ([]byte, error)
	Put(key string, data []byte) error
	Bucket(path []string) (Bucket, error)
	Foreach(fn func(key string, value []byte) error) error
	Clear() error
	Last() ([]byte, error)
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

	// fmt.Printf("getPath %s -> %x\n", path, data)
	return data, nil
}

func putPath(kv KV, path string, data []byte) error {
	bkt, key, err := findBucket(kv, path)
	// fmt.Printf("putPath %s <- %x\n", path, data)
	if err != nil {
		return err
	}

	return bkt.Put(key, data)
}

///////// BOLT KEY/VALUE IMPLEMENTATION //////////

type BoltKV struct {
	db *bolt.DB
}

func NewBoltKV(path string) (*BoltKV, error) {
	options := &bolt.Options{Timeout: 1 * time.Second}

	if err := os.MkdirAll(path, 0777); err != nil {
		return nil, err
	}

	db, err := bolt.Open(filepath.Join(path, "index.bolt"), 0600, options)
	if err != nil {
		return nil, err
	}
	db.Path()

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

func (kv *BoltKV) Export(w io.Writer) error {
	return kv.db.View(func(tx *bolt.Tx) error {
		_, err := tx.WriteTo(w)
		return err
	})
}

func (kv *BoltKV) Import(r io.Reader) error {
	path := kv.db.Path()

	if err := kv.Close(); err != nil {
		return err
	}

	fd, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	defer fd.Close()

	if _, err := io.Copy(fd, r); err != nil {
		return err
	}

	newKv, err := NewBoltKV(filepath.Dir(path))
	if err != nil {
		return err
	}

	*kv = *newKv

	return nil
}

////////////

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

func (bb *BoltBucket) Foreach(fn func(key string, value []byte) error) error {
	return bb.dig(bb.path, false, func(bkt *bolt.Bucket) error {
		cur := bkt.Cursor()
		for key, val := cur.First(); key != nil; key, val = cur.Next() {
			var copyValue []byte

			if val != nil {
				copyValue = make([]byte, len(val))
				copy(copyValue, val)
			}

			if err := fn(string(key), copyValue); err != nil {
				return err
			}
		}

		return nil
	})
}

func (bb *BoltBucket) Clear() error {
	if len(bb.path) < 1 {
		return ErrEmptyPath
	}

	dirname := bb.path[:len(bb.path)-1]
	basename := bb.path[len(bb.path)-1]

	return bb.dig(dirname, true, func(parBkt *bolt.Bucket) error {
		return parBkt.DeleteBucket([]byte(basename))
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

func (bb *BoltBucket) Last() ([]byte, error) {
	var data []byte

	err := bb.dig(bb.path, false, func(bucket *bolt.Bucket) error {
		cursor := bucket.Cursor()
		_, data = cursor.Last()
		return nil
	})

	if err != nil {
		return nil, err
	}

	return data, nil
}
