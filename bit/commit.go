package bit

import (
	multihash "github.com/jbenet/go-multihash"
	"time"
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

// Commit groups a changese
type Commit struct {
	// Optional commit message
	Message string

	// Time at this commit was conceived.
	ModTime time.Time

	// Set of files that were changed.
	Changes map[File]ChangeType

	// Parent commit (only nil for initial commit)
	Parent *Commit
}

func (c *Commit) Hash() multihash.Multihash {
	// TODO
	return nil
}

func (c *Commit) Contains(file File) ChangeType {
	if c == nil {
		return ChangeInvalid
	}

	if changeType, ok := c.Changes[file]; ok {
		return changeType
	}

	return c.Parent.Contains(file)
}

func NewCommit(parent *Commit, msg string, diffMap map[File]ChangeType) *Commit {
	return &Commit{
		Message: msg,
		ModTime: time.Now(),
		Parent:  parent,
		Changes: diffMap,
	}
}
