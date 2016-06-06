package store

import (
	"fmt"
	"sort"
	"time"

	"github.com/boltdb/bolt"
	"github.com/disorganizer/brig/id"
	"github.com/disorganizer/brig/store/wire"
	"github.com/gogo/protobuf/proto"
	"github.com/jbenet/go-multihash"
)

var (
	ErrEmptyStage         = fmt.Errorf("Nothing staged. No commit done")
	ErrEmptyCommitMessage = fmt.Errorf("Not doing a commit due to missing messsage")
)

// Merge describes the merge of two stores at one point in history.
type Merge struct {
	// With is the store owner of the store we merged with.
	With id.ID

	// Hash of the commit in the other store we merged with.
	Hash *Hash
}

func (mg *Merge) ToProto() (*wire.Merge, error) {
	return &wire.Merge{
		With: proto.String(string(mg.With)),
		Hash: mg.Hash.Bytes(),
	}, nil
}

func (mg *Merge) FromProto(protoMerge *wire.Merge) error {
	ID, err := id.Cast(protoMerge.GetWith())
	if err != nil {
		return err
	}

	hash, err := multihash.Cast(protoMerge.GetHash())
	if err != nil {
		return err
	}

	mg.With = ID
	mg.Hash = &Hash{hash}
	return nil
}

// Commit groups a change set
type Commit struct {
	// Commit message (might be auto-generated)
	Message string

	// Author is the id of the committer.
	Author id.ID

	// Time at this commit was conceived.
	ModTime time.Time

	// Checkpoints is the bag of actual changes.
	Checkpoints Checkpoints

	// Hash of this commit
	Hash *Hash

	// TreeHash is the hash of the root node at this point in time
	TreeHash *Hash

	// Parent commit (only nil for initial commit)
	Parent *Commit

	// store is needed to marshal/unmarshal properly
	store *Store

	// Merge is set if this is a merge commit (nil otherwise)
	Merge *Merge
}

func NewEmptyCommit(store *Store, author id.ID) *Commit {
	return &Commit{
		store:   store,
		ModTime: time.Now(),
		Author:  author,
	}
}

func (cm *Commit) FromProto(c *wire.Commit) error {
	author, err := id.Cast(c.GetAuthor())
	if err != nil {
		return err
	}

	modTime := time.Time{}
	if err := modTime.UnmarshalBinary(c.GetModTime()); err != nil {
		return err
	}

	hash, err := multihash.Cast(c.GetHash())
	if err != nil {
		return err
	}

	treeHash, err := multihash.Cast(c.GetTreeHash())
	if err != nil {
		return err
	}

	var checkpoints []*Checkpoint

	for _, protoCheckpoint := range c.GetCheckpoints() {
		checkpoint := &Checkpoint{}
		if err := checkpoint.FromProto(protoCheckpoint); err != nil {
			return err
		}

		checkpoints = append(checkpoints, checkpoint)
	}

	var parentCommit *Commit

	if c.GetParentHash() != nil && cm.store != nil {
		err = cm.store.viewWithBucket(
			"commits",
			func(tx *bolt.Tx, bckt *bolt.Bucket) error {
				parentData := bckt.Get(c.GetParentHash())
				if parentData == nil {
					return fmt.Errorf("No commit with hash `%x`", c.GetParentHash())
				}

				protoCommit := &wire.Commit{}
				if err := proto.Unmarshal(parentData, protoCommit); err != nil {
					return err
				}

				return NewEmptyCommit(cm.store, "").FromProto(protoCommit)
			},
		)

		if err != nil {
			return err
		}
	}

	protoMergeInfo := c.GetMerge()
	if protoMergeInfo != nil {
		mergeInfo := &Merge{}
		if err := mergeInfo.FromProto(protoMergeInfo); err != nil {
			return err
		}

		cm.Merge = mergeInfo
	}

	// Set commit data if everything worked:
	cm.Message = c.GetMessage()
	cm.Author = author
	cm.ModTime = modTime
	cm.Checkpoints = checkpoints
	cm.Hash = &Hash{hash}
	cm.TreeHash = &Hash{treeHash}
	cm.Parent = parentCommit
	return nil
}

func (cm *Commit) ToProto() (*wire.Commit, error) {
	pcm := &wire.Commit{}
	modTime, err := cm.ModTime.MarshalBinary()
	if err != nil {
		return nil, err
	}

	var checkpoints []*wire.Checkpoint

	for _, checkpoint := range cm.Checkpoints {
		protoCheckpoint, err := checkpoint.ToProto()
		if err != nil {
			return nil, err
		}

		checkpoints = append(checkpoints, protoCheckpoint)
	}

	if cm.Merge != nil {
		protoMergeInfo, err := cm.Merge.ToProto()
		if err != nil {
			return nil, err
		}

		pcm.Merge = protoMergeInfo
	}

	pcm.Message = proto.String(cm.Message)
	pcm.Author = proto.String(string(cm.Author))
	pcm.ModTime = modTime
	pcm.Hash = cm.Hash.Bytes()
	pcm.TreeHash = cm.TreeHash.Bytes()
	pcm.Checkpoints = checkpoints

	if cm.Parent != nil {
		pcm.ParentHash = cm.Parent.Hash.Bytes()
	}

	return pcm, nil
}

func (cm *Commit) MarshalProto() ([]byte, error) {
	protoCmt, err := cm.ToProto()
	if err != nil {
		return nil, err
	}

	return proto.Marshal(protoCmt)
}

func (cm *Commit) UnmarshalProto(data []byte) error {
	protoCmt := &wire.Commit{}
	if err := proto.Unmarshal(data, protoCmt); err != nil {
		return err
	}

	return cm.FromProto(protoCmt)
}

///////////////////////////////////
/// STORE METHOD IMPLEMENTATION ///
///////////////////////////////////

// Head returns the most recent commit.
// Commit will be always non-nil if error is nil,
// the initial commit has no changes.
func (st *Store) Head() (*Commit, error) {
	st.mu.Lock()
	defer st.mu.Unlock()

	return st.head()
}

// Unlocked version of Head()
func (st *Store) head() (*Commit, error) {
	cmt := NewEmptyCommit(st, st.ID)

	err := st.viewWithBucket("refs", func(tx *bolt.Tx, bkt *bolt.Bucket) error {
		data := bkt.Get([]byte("HEAD"))
		if data == nil {
			return fmt.Errorf("No HEAD in database")
		}

		return cmt.UnmarshalProto(data)
	})

	if err != nil {
		return nil, err
	}

	return cmt, nil
}

// Status shows how a Commit would look like if Commit() would be called.
func (st *Store) Status() (*Commit, error) {
	st.mu.Lock()
	defer st.mu.Unlock()

	return st.status()
}

func (st *Store) makeCommitHash(current, parent *Commit) (*Hash, error) {
	// This is inefficient, but is supposed to be easy to understand
	// while this is still playground stuff.
	s := ""
	s += fmt.Sprintf("Parent:  %s\n", parent.Hash.B58String())
	s += fmt.Sprintf("Message: %s\n", current.Message)
	s += fmt.Sprintf("Author:  %s\n", current.Author)

	hash := current.TreeHash.Clone()

	fmt.Printf("tree %v\nhash %v\n", current.TreeHash, hash)
	if err := hash.MixIn([]byte(s)); err != nil {
		return nil, err
	}

	return hash, nil
}

// Unlocked version of Status()
func (st *Store) status() (*Commit, error) {
	head, err := st.head()
	if err != nil {
		return nil, err
	}

	cmt := NewEmptyCommit(st, st.ID)
	cmt.Parent = head
	cmt.Message = "Uncommitted changes"
	cmt.TreeHash = st.Root.Hash().Clone()

	hash, err := st.makeCommitHash(cmt, head)
	if err != nil {
		return nil, err
	}

	cmt.Hash = hash

	err = st.viewWithBucket("stage", func(tx *bolt.Tx, bkt *bolt.Bucket) error {
		return bkt.ForEach(func(bpath, bckpnt []byte) error {
			checkpoint := &Checkpoint{}
			if err := checkpoint.Unmarshal(bckpnt); err != nil {
				return err
			}

			cmt.Checkpoints = append(cmt.Checkpoints, checkpoint)
			return nil
		})
	})

	if err != nil {
		return nil, err
	}

	return cmt, nil
}

// Commit saves a commit in the store history.
func (st *Store) MakeCommit(msg string) error {
	st.mu.Lock()
	defer st.mu.Unlock()

	if msg == "" {
		return ErrEmptyCommitMessage
	}

	cmt, err := st.status()
	if err != nil {
		return err
	}

	return st.db.Update(func(tx *bolt.Tx) error {
		// Check if the stage area contains something:
		stage := tx.Bucket([]byte("stage"))
		if stage == nil {
			return ErrNoSuchBucket{"stage"}
		}

		if stage.Stats().KeyN == 0 {
			return ErrEmptyStage
		}

		// Flush the staging area:
		if err := tx.DeleteBucket([]byte("stage")); err != nil {
			return err
		}

		if _, err := tx.CreateBucket([]byte("stage")); err != nil {
			return err
		}

		cmts := tx.Bucket([]byte("commits"))
		if cmts == nil {
			return ErrNoSuchBucket{"commits"}
		}

		// Put the new commit in the commits bucket:
		cmt.Message = msg
		data, err := cmt.MarshalProto()
		if err != nil {
			return err
		}

		if err := cmts.Put(cmt.Hash.Bytes(), data); err != nil {
			return err
		}

		// Update HEAD:
		refs := tx.Bucket([]byte("refs"))
		if refs == nil {
			return ErrNoSuchBucket{"refs"}
		}

		return refs.Put([]byte("HEAD"), data)
	})
}

// TODO: respect from/to ranges
func (st *Store) Log() (*Commits, error) {
	var cmts Commits

	st.mu.Lock()
	defer st.mu.Unlock()

	err := st.viewWithBucket("commits", func(tx *bolt.Tx, bkt *bolt.Bucket) error {
		return bkt.ForEach(func(k, v []byte) error {
			cmt := NewEmptyCommit(st, st.ID)
			if err := cmt.UnmarshalProto(v); err != nil {
				return err
			}

			cmts = append(cmts, cmt)
			return nil
		})
	})

	if err != nil {
		return nil, err
	}

	sort.Sort(&cmts)
	return &cmts, nil
}

// Commits is a list of single commits.
// It is used to enable chronological sorting of a bunch of commits.
type Commits []*Commit

func (cs *Commits) Len() int {
	return len(*cs)
}

func (cs *Commits) Less(i, j int) bool {
	return (*cs)[i].ModTime.Before((*cs)[j].ModTime)
}

func (cs *Commits) Swap(i, j int) {
	(*cs)[i], (*cs)[j] = (*cs)[j], (*cs)[i]
}

func (cs *Commits) ToProto() (*wire.Commits, error) {
	protoCmts := &wire.Commits{}

	for _, cmt := range *cs {
		protoCmt, err := cmt.ToProto()
		if err != nil {
			return nil, err
		}

		protoCmts.Commits = append(protoCmts.Commits, protoCmt)
	}

	return protoCmts, nil
}
