package catfs

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	e "github.com/pkg/errors"
	c "github.com/sahib/brig/catfs/core"
	ie "github.com/sahib/brig/catfs/errors"
	n "github.com/sahib/brig/catfs/nodes"
)

var (
	indexCommitPattern = regexp.MustCompile(`^commit\[([0-9]+)\]$`)
)

// validateRev check is a rev spec looks like it's valid
// from a syntactic point of view.
//
// A valid ref may contain only letters or numbers, but might end with an
// arbitary number of '^' at the end. Unicode is allowed.
// As special case it might also match indexCommitPattern.
//
// If any violation is dected, an error is returned.
func validateRev(rev string) error {
	if indexCommitPattern.Match([]byte(rev)) {
		return nil
	}

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
		return nil, e.Wrapf(err, "validate")
	}

	lowerRev := strings.ToLower(rev)
	matches := indexCommitPattern.FindSubmatch([]byte(lowerRev))
	if len(matches) >= 2 {
		index, err := strconv.ParseInt(string(matches[1]), 10, 64)
		if err != nil {
			return nil, e.Wrapf(err, "failed to parse commit index spec")
		}

		return lkr.CommitByIndex(index)
	}

	var cmt *n.Commit
	cmtNd, err := lkr.ResolveRef(lowerRev)

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
