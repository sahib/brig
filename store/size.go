package store

import (
	"github.com/dustin/go-humanize"
)

// FileSize is a large enough integer for file sizes, offering a few util methods.
type FileSize int64

func (s FileSize) String() string {
	return humanize.Bytes(uint64(s))
}
