package store

import (
	"fmt"
	"time"
)

type Diff struct {
	// A is the store we will merge to.
	A *Store

	// B is the store we will merge from.
	B *Store

	// AWants is a slice of commits with all commits
	// B has that A has not, starting with the earlier commits.
	AWants []*Commit

	// BWants is a slice of commits with all commits
	// A has that B has not, starting with the earlier commits.
	BWants []*Commit

	// MergedAlready is true when A merged with B earlier.
	AMergedAlready bool
	BMergedAlready bool
}

func diff(a, b *Store) ([]*Commit, bool, error) {
	curr, err := a.Head()
	if err != nil {
		return nil, false, err
	}

	if curr == nil {
		return nil, false, fmt.Errorf("bug: no head in `b` -> initial commit missing.")
	}

	path := []*Commit{}
	mergedAlready := false

	for curr != nil {
		if curr.Merge != nil && curr.Merge.With == b.ID {
			// `b` merged with `a` in this commit.
			// diff is therefore 'prev..head_a'
			mergedAlready = true
			break
		}

		path = append(path, curr)
		curr = curr.Parent
	}

	return path, mergedAlready, nil
}

// Diff returns the commits that `other` has and `st` doesn't.
func (st *Store) Diff(other *Store) (*Diff, error) {
	awants, aMergedAlready, err := diff(st, other)
	if err != nil {
		return nil, err
	}

	bwants, bMergedAlready, err := diff(other, st)
	if err != nil {
		return nil, err
	}

	return &Diff{
		A:              st,
		B:              other,
		AWants:         awants,
		BWants:         bwants,
		AMergedAlready: aMergedAlready,
		BMergedAlready: bMergedAlready,
	}, nil
}

// Squash combines all commits in a diff into one single merge commit.
// The merge commit also has the `Merge` attribute filled correctly.
func (df *Diff) Squash() (*Commit, error) {
	if len(df.AWants) == 0 {
		return nil, fmt.Errorf("Cannot squash empty diff")
	}

	aHead, err := df.A.Head()
	if err != nil {
		return nil, err
	}

	merged := NewEmptyCommit(df.A, df.B.ID)
	merged.ModTime = time.Now()
	merged.Parent = aHead
	merged.Merge = &Merge{
		With: df.B.ID,
		Hash: df.AWants[0].Hash,
	}

	for _, cmt := range df.AWants {
		// TODO: Filter duplicate checkpoints?
		for _, ckpnt := range cmt.Checkpoints {
			merged.Checkpoints = append(merged.Checkpoints, ckpnt)
		}
	}

	return nil, nil
}

// Apply applies `diff` onto `st`.
// A new merge commit will be created if fast forward was not possible.
func (st *Store) Apply(diff *Diff) error {
	return nil
}
