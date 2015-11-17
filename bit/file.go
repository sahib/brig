package bit

import (
	"time"

	multihash "github.com/jbenet/go-multihash"
)

type File interface {
	// Path relative to the repo root
	Path() string

	// File size of the file in bytes
	Size() int

	// Modification timestamp (with timezone)
	Mtime() time.Time

	// Hash of the unencrypted file
	Hash() multihash.Multihash

	// Hash of the encrypted file from IPFS
	IpfsHash() multihash.Multihash
}

func NewFile(path string) (*File, error) {
	// TODO:
	return nil, nil
}
