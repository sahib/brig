package store

/*

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

// diff returns the commits a has, that b does not.
// Oldest commit is first.
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

	// Reverse the order, so that older commits are first:
	for i := 0; i < len(path)/2; i++ {
		end := len(path) - i - 1
		path[i], path[end] = path[end], path[i]
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
// It looks at both AWants and BWants, where commits from `A` win.
// The merge commit also has the `Merge` attribute filled correctly,
// but not the TreeHash and Hash attribute, since the merge commit
// was not applied yet onto the store.
func (df *Diff) Squash() (*Commit, error) {
	if len(df.AWants) == 0 {
		return nil, fmt.Errorf("Cannot squash empty diff")
	}

	aHead, err := df.A.Head()
	if err != nil {
		return nil, err
	}

	merged := NewEmptyCommit(df.A, df.A.ID)
	merged.ModTime = time.Now()
	merged.Parent = aHead
	merged.Merge = &Merge{
		With: df.B.ID,
		Hash: df.AWants[len(df.AWants)-1].Hash, // Newest commit that we want.
	}

	changeset := make(map[string]*Checkpoint)

	message := fmt.Sprintf("Merged with `%s`:\n", df.B.ID)

	// Only take the last checkpoint of a file:
	for _, cmt := range df.AWants {
		for _, ckpnt := range cmt.Checkpoints {
			curr, ok := changeset[ckpnt.Path]
			if !ok || ckpnt.ModTime.After(curr.ModTime) {
				changeset[ckpnt.Path] = ckpnt
			}
		}
	}

	// If we had modifications on our own in the meantime,
	// let them win over other's modifications.
	for _, cmt := range df.BWants {
		for _, ckpnt := range cmt.Checkpoints {
			curr, ok := changeset[ckpnt.Path]
			if ok && !curr.Hash.Equal(ckpnt.Hash) {
				// Uh-oh, both sides have modified this path.
				// Keep our change, but backup their change as ''
				conflictPath := fmt.Sprintf("%s.%s", ckpnt.Path, df.B.ID.AsPath())

				log.Warningf(
					"%s was changed by both differently; keeping %s's version as %s",
					ckpnt.Path, df.B.ID, conflictPath,
				)

				changeset[ckpnt.Path] = ckpnt
				curr.Path = conflictPath
				changeset[conflictPath] = curr
			}
		}
	}

	// Convert map to checkpoint slice:
	for _, ckpnt := range changeset {
		merged.Checkpoints = append(merged.Checkpoints, ckpnt)
	}

	// Sort by path:
	sort.Stable(&merged.Checkpoints)

	// Be nice and create a commit summary:
	for _, ckpnt := range merged.Checkpoints {
		message += fmt.Sprintf("  %s %s\n", ckpnt.Change.String(), ckpnt.Path)
	}

	merged.Message = message
	return merged, nil
}

// Apply applies `diff` onto `st`.
func (st *Store) Apply(diff *Diff) error {
	merged, err := diff.Squash()
	if err != nil {
		return err
	}

	return st.ApplyMergeCommit(merged)
}

// ApplyMergeCommit will apply the checkpoints in `cmt` onto `st`.
// It will also fill in the Hash and TreeHash attribute of `cmt`
// since it will (likely) have changed after merging.
func (st *Store) ApplyMergeCommit(cmt *Commit) error {
	st.mu.Lock()
	defer st.mu.Unlock()

	var err error

	for _, ckpnt := range cmt.Checkpoints {
		switch ckpnt.Change {
		case ChangeAdd:
			file, err := NewFile(st, ckpnt.Path)
			if err != nil {
				return err
			}

			err = st.insertMetadata(file, ckpnt.Path, ckpnt.Hash, true, ckpnt.Size)
		case ChangeModify:
			file := st.Root.Lookup(ckpnt.Path)
			if file == nil {
				return NoSuchFile(ckpnt.Path)
			}

			err = st.insertMetadata(file, ckpnt.Path, ckpnt.Hash, false, ckpnt.Size)
		case ChangeRemove:
			err = st.remove(ckpnt.Path, false)
		case ChangeMove:
			err = st.move(ckpnt.OldPath, ckpnt.Path)
		default:
			return fmt.Errorf("Invalid change type `%d`", ckpnt.Change)
		}
	}

	cmt.TreeHash = st.Root.Hash().Clone()
	hash, err := st.makeCommitHash(cmt, cmt.Parent)

	if err != nil {
		log.Errorf("Unable to create commit hash of merge commit: %v", err)
		return err
	}

	cmt.Hash = hash

	// Update HEAD - TODO: finally neeed proper refs.
	return st.updateHEAD(cmt)
}
*/
