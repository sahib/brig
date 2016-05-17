package store

type Diff struct {
	A, B              *Store
	Path              []*Commit
	CanFastForward    bool
	HasCommonAncestor bool
}

// Diff shows the difference between `st` and `other`.
// It returns a Diff-struct, filled in a way that the diff contains
// the commits that need to be applied to `st` so it also contains
// the contents of `other`, with merge conflicts resolved in favour
// of `st`.
//
// - If `st` and `other` have a common ancestor, return the path
//   from the ancestor to the head of `other`.
// - If `st` and `other` have no common ancestor,
//
func (st *Store) Diff(other *Store) (*Diff, error) {
	diff := &Diff{
		A: st,
		B: other,
	}

	return diff, nil
}

func (df *Diff) Combine() (*Commit, error) {
	return nil, nil
}

// Apply applies `diff` onto `st`.
// A new merge commit will be created if fast forward was not possible.
func (st *Store) Apply(diff *Diff) error {
	return nil
}
