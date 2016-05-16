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

// Commit groups a change set
type Commit struct {
	// Commit message (might be auto-generated)
	Message string

	// Author is the id of the committer.
	Author id.ID

	// Time at this commit was conceived.
	ModTime time.Time

	// Checkpoints is the bag of actual changes.
	Checkpoints []*Checkpoint

	// Hash of this commit (== hash of the root node)
	Hash *Hash

	// Parent commit (only nil for initial commit)
	Parent *Commit

	// store is needed to marshal/unmarshal properly
	store *Store
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

	// Set commit data if everything worked:
	cm.Message = c.GetMessage()
	cm.Author = author
	cm.ModTime = modTime
	cm.Checkpoints = checkpoints
	cm.Hash = &Hash{hash}
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

	pcm.Message = proto.String(cm.Message)
	pcm.Author = proto.String(string(cm.Author))
	pcm.ModTime = modTime
	pcm.Hash = cm.Hash.Bytes()
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
	s += fmt.Sprintf("ModTime: %s\n", current.ModTime.String())
	s += fmt.Sprintf("Author:  %s\n", current.Author)
	s += fmt.Sprintf("Message: %s\n", current.Message)

	hash := st.Root.Hash().Clone()
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
