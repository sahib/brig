package catfs

import (
	c "github.com/disorganizer/brig/catfs/core"
	ie "github.com/disorganizer/brig/catfs/errors"
	n "github.com/disorganizer/brig/catfs/nodes"
	h "github.com/disorganizer/brig/util/hashlib"
)

// parseRev resolves a base58 to a commit or if it looks like a refname
// it tries to resolve that (HEAD, CURR, INIT e.g.).
func parseRev(lkr *c.Linker, rev string) (*n.Commit, error) {
	var cmt *n.Commit
	cmtNd, err := lkr.ResolveRef(rev)

	if err != nil {
		hash, err := h.FromB58String(rev)
		if err != nil {
			return nil, err
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

// validaRefname will return an error if `name` is an invalid refname.
func validateRefname(name string) error {
	return nil
}
