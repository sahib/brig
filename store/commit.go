package store

import (
	"fmt"
	"time"

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

type Author struct {
	ident id.ID
	hash  *Hash
}

func (a *Author) ID() id.ID {
	return a.ident
}

func (a *Author) Hash() string {
	return a.hash.B58String()
}

func (a *Author) FromProto(pa *wire.Author) error {
	ident, err := id.Cast(pa.GetName())
	if err != nil {
		return err
	}

	mh, err := multihash.FromB58String(pa.GetHash())
	if err != nil {
		return err
	}

	a.ident = ident
	a.hash = &Hash{mh}
	return nil
}

func (a *Author) ToProto() (*wire.Author, error) {
	return &wire.Author{
		Name: proto.String(string(a.ident)),
		Hash: proto.String(a.hash.B58String()),
	}, nil
}

////////////////////////

// Commit groups a change set
type Commit struct {
	// Commit message (might be auto-generated)
	message string

	// Author is the id of the committer.
	author *Author

	// Time at this commit was conceived.
	modTime time.Time

	// Checkpoints is the bag of actual changes.
	changeset []*CheckpointLink

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

func newEmptyCommit(fs *FS) (*Commit, error) {
	id, err := fs.NextID()
	if err != nil {
		return nil, err
	}

	return &Commit{
		id:      id,
		fs:      fs,
		modTime: time.Now(),
	}, nil
}

func (cm *Commit) FromProto(pnd *wire.Node) error {
	pcm := pnd.GetCommit()
	if pcm == nil {
		return fmt.Errorf("No commit attr in protobuf. Probably not a commit.")
	}

	author := &Author{}
	if err := author.FromProto(pcm.GetAuthor()); err != nil {
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

	root, err := multihash.Cast(pcm.GetRoot())
	if err != nil {
		return err
	}

	parent, err := multihash.Cast(pcm.GetParentHash())
	if err != nil {
		return err
	}

	var changeset []*CheckpointLink

	for _, pcl := range pcm.GetChangeset() {
		cl := &CheckpointLink{}
		if err := cl.FromProto(pcl); err != nil {
			return err
		}

		changeset = append(changeset, cl)
	}

	protoMergeInfo := pcm.GetMerge()
	if protoMergeInfo != nil {
		mergeInfo := &Merge{}
		if err := mergeInfo.FromProto(protoMergeInfo); err != nil {
			return err
		}

		cm.merge = mergeInfo
	}

	// Set commit data if everything worked:
	cm.id = pnd.GetID()
	cm.message = pcm.GetMessage()
	cm.parent = &Hash{pcm.GetParentHash()}
	cm.author = author
	cm.modTime = modTime
	cm.hash = &Hash{hash}
	cm.root = &Hash{root}
	cm.parent = &Hash{parent}
	return nil
}

func (cm *Commit) ToProto() (*wire.Node, error) {
	pcm := &wire.Commit{}
	modTime, err := cm.modTime.MarshalBinary()
	if err != nil {
		return nil, err
	}

	var changeset []*wire.CheckpointLink

	for _, link := range cm.changeset {
		plink, err := link.ToProto()
		if err != nil {
			return nil, err
		}

		changeset = append(changeset, plink)
	}

	if cm.merge != nil {
		pmerge, err := cm.merge.ToProto()
		if err != nil {
			return nil, err
		}

		pcm.Merge = pmerge
	}

	pauthor, err := cm.author.ToProto()
	if err != nil {
		return nil, err
	}

	pcm.Message = proto.String(cm.message)
	pcm.Author = pauthor
	pcm.ModTime = modTime
	pcm.Hash = cm.hash.Bytes()
	pcm.Root = cm.root.Bytes()
	pcm.Changeset = changeset

	// Check if it's the initial commit:
	if cm.parent != nil {
		pcm.ParentHash = cm.parent.Bytes()
	}

	return &wire.Node{
		ID:     proto.Uint64(cm.id),
		Commit: pcm,
	}, nil
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

func (cm *Commit) ID() uint64 {
	return cm.id
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

func (cm *Commit) GetType() NodeType {
	return NodeTypeCommit
}

///////////////////////////////
/// OWN COMMIT FUNCTIONALITY //
///////////////////////////////

func (cm *Commit) Root() *Hash {
	return cm.root
}

func (cm *Commit) AddCheckpointLink(cl *CheckpointLink) {
	cm.changeset = append(cm.changeset, cl)
}

func (cm *Commit) SetRoot(root *Hash) error {
	cm.root = root.Clone()
	return nil
}

func (cm *Commit) Finalize(author id.Peer, message string, parent *Commit) error {
	cm.message = message
	if err := cm.SetParent(parent); err != nil {
		return err
	}

	// This is inefficient, but is supposed to be easy to understand
	// while this is still playground stuff.
	s := ""
	s += fmt.Sprintf("Parent:  %s\n", parent.Hash().B58String())
	s += fmt.Sprintf("Message: %s\n", cm.message)
	s += fmt.Sprintf("Author:  %s\n", cm.author)

	hash := cm.root.Clone()
	fmt.Printf("tree %v\nhash %v\n", cm.root, hash)
	if err := hash.MixIn([]byte(s)); err != nil {
		return err
	}

	cm.modTime = time.Now()
	return nil
}

// TODO: is this needed?
// Commits is a list of single commits.
// It is used to enable chronological sorting of a bunch of commits.
type Commits []*Commit

func (cs *Commits) Len() int {
	return len(*cs)
}

func (cs *Commits) Less(i, j int) bool {
	return (*cs)[i].modTime.Before((*cs)[j].modTime)
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

		protoCmts.Commits = append(protoCmts.Commits, protoCmt.GetCommit())
	}

	return protoCmts, nil
}

func (cs *Commits) Unmarshal(data []byte) error {
	protoCmts := &wire.Commits{}
	if err := proto.Unmarshal(data, protoCmts); err != nil {
		return err
	}

	return cs.FromProto(protoCmts)
}

func (cs *Commits) FromProto(protoCmts *wire.Commits) error {
	for _, protoCmt := range protoCmts.GetCommits() {
		cmt, err := newEmptyCommit(nil)
		if err != nil {
			return err
		}

		pnode := &wire.Node{
			Type:   wire.NodeType_COMMIT.Enum(),
			Commit: protoCmt,
		}

		if err := cmt.FromProto(pnode); err != nil {
			return err
		}

		*cs = append(*cs, cmt)
	}

	return nil
}
