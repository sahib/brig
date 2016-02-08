package store

import (
	"time"

	"github.com/boltdb/bolt"
)

// import (
// 	multihash "github.com/jbenet/go-multihash"
// 	"time"
// )
//
const (
	// ChangeInvalid indicates a bug.
	ChangeInvalid = iota

	// The file was newly added.
	ChangeAdd

	// The file was modified
	ChangeModify

	// The file was removed.
	ChangeRemove
)

type ChangeType byte

// Commit groups a change set
type Commit struct {
	// Optional commit message
	Message string

	// Author is the jabber id of the committer.
	Author string

	// Time at this commit was conceived.
	ModTime time.Time

	// Set of files that were changed.
	Changes map[string]ChangeType

	// Parent commit (only nil for initial commit)
	Parent *Commit
}

func (s *Store) MakeCommit(old, curr *File) error {
	// change := ChangeInvalid
	// if old == nil {
	// 	change = ChangeAdd
	// } else if curr == nil {
	// 	change = ChangeRemove
	// } else {
	// 	// TODO: Check if something changed:
	// 	change = ChangeModify
	// }

	return s.db.Update(withBucket("commits", func(tx *bolt.Tx, bucket *bolt.Bucket) error {
		return nil
	}))

	return nil
}
