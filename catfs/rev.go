package catfs

import (
	c "github.com/disorganizer/brig/catfs/core"
	ie "github.com/disorganizer/brig/catfs/errors"
	n "github.com/disorganizer/brig/catfs/nodes"
)

// parseRev resolves a base58 to a commit or if it looks like a refname
// it tries to resolve that (HEAD, CURR, INIT e.g.).
func parseRev(lkr *c.Linker, rev string) (*n.Commit, error) {
	var cmt *n.Commit
	cmtNd, err := lkr.ResolveRef(rev)

	if err != nil {
		// Expand possible abbreviations:
		hash, err := lkr.ExpandAbbrev(rev)
		if err != nil {
			// If the file was not a valid refname
			// and the hash conversion failed it's probably invalid.
			return nil, ie.ErrNoSuchRef(rev)
		}

		cmt, err := lkr.CommitByHash(hash)
		if err != nil {
			return nil, err
		}

		return cmt, nil
	}

	cmt, ok := cmtNd.(*n.Commit)
	if !ok {
		return nil, ie.ErrBadNode
	}

	return cmt, nil
}
