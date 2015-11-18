package bit

import (
	mapset "github.com/deckarep/golang-set"
	multihash "github.com/jbenet/go-multihash"
)

type Index struct {
	// Most recent commit
	Head *Commit

	// Initial commit
	Root *Commit

	// A set of all commits:
	Commits mapset.Set
}

func (i *Index) Load() error {
	// Load from .brig/index
	return nil
}

func (i *Index) Save() error {
	// Save to .brig/index
	return nil
}

// LookupCommit returns nil if the commit does not exist.
func (i *Index) LookupCommit(multihash.Multihash) *Commit {
	return nil
}
