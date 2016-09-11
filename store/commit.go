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

////////////////////////

// Commit groups a change set
type Commit struct {
	// Commit message (might be auto-generated)
	message string

	// Author is the id of the committer.
	author id.ID

	// Time at this commit was conceived.
	modTime time.Time

	// Checkpoints is the bag of actual changes.
	heckpoints Checkpoints

	// Hash of this commit
	hash *Hash

	// TreeHash is the hash of the root node at this point in time
	root *Hash

	// Parent hash (only nil for initial commit)
	parent *Hash

	// store is needed to marshal/unmarshal properly
	fs *FS

	// Merge is set if this is a merge commit (nil otherwise)
	merge *Merge

	id uint64
}

func newEmptyCommit(fs *FS, author id.ID) (*Commit, error) {
	id, err := fs.NextID()
	if err != nil {
		return nil, err
	}

	return &Commit{
		id: id,
		fs:      fs,
		modTime: time.Now(),
		author:  author,
	}, nil
}

func (cm *Commit) FromProto(pnd *wire.Node) error {
	pcm := pnd.GetCommit()
	if pcm == nil {
		return fmt.Errorf("No commit attr in protobuf. Probably not a commit.")
	}

	author, err := id.Cast(pcm.GetAuthor())
	if err != nil {
		return err
	}

	modTime := time.Time{}
	if err := modTime.UnmarshalBinary(pcm.GetModTime()); err != nil {
		return err
	}

	hash, err := multihash.Cast(pcm.GetHash())
	if err != nil {
		return err
	}

	treeHash, err := multihash.Cast(pcm.GetTreeHash())
	if err != nil {
		return err
	}

	var checkpoints []*Checkpoint

	for _, protoCheckpoint := range pcm.GetCheckpoints() {
		checkpoint := &Checkpoint{}
		if err := checkpoint.FromProto(protoCheckpoint); err != nil {
			return err
		}

		checkpoints = append(checkpoints, checkpoint)
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
	cm.id = pnd.GetID()
	cm.message = pcm.GetMessage()
	cm.parent = &Hash{pcm.GetParentHash()}
	cm.author = author
	cm.modTime = modTime
	cm.checkpoints = checkpoints
	cm.hash = &Hash{hash}
	cm.treeHash = &Hash{treeHash}
	cm.parent = parentCommit
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

	pcm.ID = proto.Uint64(cm.id)
	pcm.Message = proto.String(cm.message)
	pcm.Author = proto.String(string(cm.author))
	pcm.ModTime = modTime
	pcm.Hash = cm.hash.Bytes()
	pcm.Root = cm.root.Bytes()
	pcm.Checkpoints = checkpoints

	// Check if it's the initial commit:
	if cm.parent != nil {
		pcm.ParentHash = cm.parent.Hash().Bytes()
	}

	return pcm, nil
}

/////////////////// METADATA INTERFACE ///////////////////

func (cm *Commit) Name() string {
	return cm.hash.B58String()
}

func (cm *Commit) Size() uint64 {
	root, err := cm.fs.DirectoryByHash(cm.root)
	if err != nil {
		return 0
	}

	return root.Size()
}

func (cm *Commit) Hash() *Hash {
	return cm.hash
}

func (cm *Commit) ModTime() time.Time {
	return cm.modTime
}

/////////////// HIERARCHY INTERFACE ///////////////

func (cm *Commit) NChildren() int {
	return 1
}

func (cm *Commit) Child(name string) (Node, error) {
	return cm.fs.DirectoryByHash(cm.root)
}

func (cm *Commit) Parent() (Node, error) {
	return cm.fs.CommitByHash(cm.parent)
}

func (cm *Commit) SetParent(nd Node) error {
	cm.parent = nd.Hash()
	return nil
}

return (cm *Commit) GetType() NodeType {
	return NodeTypeCommit
}

///////////////////////////////////
/// STORE METHOD IMPLEMENTATION ///
///////////////////////////////////

// Status shows how a Commit would look like if Commit() would be called.
func (st *Store) Status() (*Commit, error) {
	return st.status()
}

func (cm *Commit) Finalize(message string, parent *Commit) error {
	cm.message = message
	if err := cm.SetParent(parent); err != nil {
		return err
	}

	// This is inefficient, but is supposed to be easy to understand
	// while this is still playground stuff.
	s := ""
	s += fmt.Sprintf("Parent:  %s\n", parent.Hash.B58String())
	s += fmt.Sprintf("Message: %s\n", current.Message)
	s += fmt.Sprintf("Author:  %s\n", current.Author)

	hash := current.Root.Clone()

	fmt.Printf("tree %v\nhash %v\n", current.Root, hash)
	if err := hash.MixIn([]byte(s)); err != nil {
		return err
	}

	return nil
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
}

// TODO: respect from/to ranges
func (fs *FS) Log() (*Commits, error) {
	var cmts Commits

	head, err := fs.Head()
	if err != nil {
		return nil, err
	}

	for curr := head; curr != nil; curr = curr.ParentCommit() {
		cmts = append(cmts, curr)
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
