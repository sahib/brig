package db

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/disorganizer/brig/util"
	"github.com/peterbourgon/diskv"
)

// DiskvDatabase is a database that simply uses the filesystem as storage.
// Each bucket is one directory. Leaf keys are simple files.
// The exported form of the database is simply a gzipped .tar of the directory.
type DiskvDatabase struct {
	db *diskv.Diskv
}

// NewDiskvDatabase creates a new database at `basePath`.
func NewDiskvDatabase(basePath string) (*DiskvDatabase, error) {
	return &DiskvDatabase{
		db: diskv.New(
			diskv.Options{
				BasePath: basePath,
				Transform: func(s string) []string {
					return []string{}
				},
			},
		),
	}, nil
}

// Get a single value from `bucket` by `key`.
func (db *DiskvDatabase) Get(bucket string, key string) ([]byte, error) {
	data, err := db.db.Read(path.Join(bucket, key))
	if os.IsNotExist(err) {
		return nil, ErrNoSuchKey
	}

	return data, err
}

// Set stores a new `val` under `key` at `bucket`.
// Implementation detail: `key` may contain slashes (/). If used, those keys
// will result in a nested directory structure.
func (db *DiskvDatabase) Set(bucket string, key string, val []byte) error {
	if err := os.MkdirAll(filepath.Join(db.db.BasePath, bucket), 0700); err != nil {
		return err
	}
	return db.db.Write(path.Join(bucket, key), val)
}

// Export writes all key/valeus into a gzipped .tar that is written to `w`.
func (db *DiskvDatabase) Export(w io.Writer) error {
	gzw := gzip.NewWriter(w)
	gzw.Name = "brigmeta.gz"
	gzw.Comment = "brig metadata db"
	gzw.ModTime = time.Now()

	tw := tar.NewWriter(gzw)
	walker := func(path string, info os.FileInfo, err error) error {
		if !info.Mode().IsRegular() {
			return nil
		}

		hdr := &tar.Header{
			Name: path[len(db.db.BasePath):],
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

	if err := filepath.Walk(db.db.BasePath, walker); err != nil {
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
		fullPath := filepath.Join(db.db.BasePath, hdr.Name)
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
func (db *DiskvDatabase) Close() error {
	return nil
}
