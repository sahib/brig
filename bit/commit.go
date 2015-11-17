package bit

import (
	"time"

	mapset "github.com/deckarep/golang-set"
	multihash "github.com/jbenet/go-multihash"
)

const (
	ChangeInvalid = iota

	// The file was newly added.
	ChangeAdd

	// The file was modified
	ChangeModify

	// The file was removed.
	ChangeRemove
)

type ChangeType byte

// Change describes the details of modification of a File.
type Change struct {
	Subject File
	Type    ChangeType
}

// Commit groups a changese
type Commit struct {
	// Optional commit message
	Message string

	// Time at this commit was conceived.
	ModTime time.Time

	// Set of files that were changed.
	Changes mapset.Set

	// Parent commit (only nil for initial commit)
	Parent *Commit
}

func (c *Commit) Hash() multihash.Multihash {
	// TODO
	return nil
}
