package store

import (
	"fmt"

	"github.com/boltdb/bolt"
)

type ErrNoSuchBucket struct {
	Name string
}

func (e ErrNoSuchBucket) Error() string {
	return fmt.Sprintf("bolt: no bucket named `%s`", e.Name)
}
