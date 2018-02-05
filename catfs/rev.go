package catfs

import (
	"fmt"
	"unicode"

	c "github.com/sahib/brig/catfs/core"
	ie "github.com/sahib/brig/catfs/errors"
	n "github.com/sahib/brig/catfs/nodes"
)

// validateRev check is a rev spec looks like it's valid
// from a syntactiv point of view.
//
// A valid ref may contain only letters or numbers, but might end with an
// arbitary number of '^' at the end. Unicode is allowed.
//
// If any violation is dected, an error is returned.
func validateRev(rev string) error {
	foundUp := false
	for _, c := range rev {
		if unicode.IsLetter(c) || unicode.IsNumber(c) {
			if foundUp {
				return fmt.Errorf("normal character after ^")
			}
			continue
		}

		switch c {
		case '^':
			foundUp = true
		default:
			return fmt.Errorf("invalid character in ref: `%v`", c)
		}
	}

	return nil
}

// parseRev resolves a base58 to a commit or if it looks like a refname
// it tries to resolve that (HEAD, CURR, INIT e.g.).
func parseRev(lkr *c.Linker, rev string) (*n.Commit, error) {
	if err := validateRev(rev); err != nil {
		return nil, err
	}

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
