package store

import (
	"fmt"
	"path"
	"time"

	"github.com/disorganizer/brig/id"
	"github.com/disorganizer/brig/store/wire"
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
		With: string(mg.With),
		Hash: mg.Hash.Bytes(),
	}, nil
}

func (mg *Merge) FromProto(protoMerge *wire.Merge) error {
	ID, err := id.Cast(protoMerge.With)
	if err != nil {
		return err
	}

	hash, err := multihash.Cast(protoMerge.Hash)
	if err != nil {
		return err
	}

	mg.With = ID
	mg.Hash = &Hash{hash}
	return nil
}

func (mg *Merge) String() string {
	return fmt.Sprintf("merge(%s:%s)", mg.With, mg.Hash.B58String())
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

func StageAuthor() *Author {
	return &Author{"unknown", EmptyHash.Clone()}
}

func (a *Author) FromProto(pa *wire.Author) error {
	ident, err := id.Cast(pa.Name)
	if err != nil {
		return err
	}

	mh, err := multihash.FromB58String(pa.Hash)
	if err != nil {
		return err
	}

	a.ident = ident
	a.hash = &Hash{mh}
	return nil
}

func (a *Author) ToProto() (*wire.Author, error) {
	// Author might be nil for the staging commit:
	if a == nil {
		return StageAuthor().ToProto()
	}

	return &wire.Author{
		Name: string(a.ident),
		Hash: a.hash.B58String(),
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

func (cm *Commit) String() string {
	return fmt.Sprintf(
		"commit %s/%s <%s>",
		cm.hash.B58String(),
		cm.root.B58String(),
		cm.message,
	)
}

func (cm *Commit) GetChangeset() []*CheckpointLink {
	return cm.changeset
}

func (cm *Commit) FromProto(pnd *wire.Node) error {
	pcm := pnd.Commit
	if pcm == nil {
		return fmt.Errorf("No commit attr in protobuf. Probably not a commit.")
	}

	author := &Author{}
	if err := author.FromProto(pcm.Author); err != nil {
		return err
	}

	modTime := time.Time{}
	if err := modTime.UnmarshalBinary(pnd.ModTime); err != nil {
		return err
	}

	hash, err := multihash.Cast(pnd.Hash)
	if err != nil {
		return err
	}

	root, err := multihash.Cast(pcm.Root)
	if err != nil {
		return err
	}

	var parent multihash.Multihash
	if len(pcm.Parent) > 0 {
		parent, err = multihash.Cast(pcm.Parent)
		if err != nil {
			return err
		}
	}

	var changeset []*CheckpointLink

	for _, pcl := range pcm.Changeset {
		cl := &CheckpointLink{}
		if err := cl.FromProto(pcl); err != nil {
			return err
		}

		changeset = append(changeset, cl)
	}

	protoMergeInfo := pcm.Merge
	if protoMergeInfo != nil {
		mergeInfo := &Merge{}
		if err := mergeInfo.FromProto(protoMergeInfo); err != nil {
			return err
		}

		cm.merge = mergeInfo
	}

	// Set commit data if everything worked:
	cm.id = pnd.ID
	cm.message = pcm.Message
	cm.author = author
	cm.modTime = modTime
	cm.hash = &Hash{hash}
	cm.root = &Hash{root}
	cm.changeset = changeset

	if parent != nil {
		cm.parent = &Hash{parent}
	}
	return nil
}

func (cm *Commit) ExpandProto(pnd *wire.Node) error {
	for _, link := range cm.changeset {
		ckp, err := link.Resolve(cm.fs)
		if err != nil {
			return err
		}

		pckp, err := ckp.ToProto()
		if err != nil {
			return err
		}

		if err := ckp.ExpandProto(cm.fs, pckp); err != nil {
			return err
		}

		pnd.Commit.Checkpoints = append(pnd.Commit.Checkpoints, pckp)
	}

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

	pcm.Message = cm.message
	pcm.Author = pauthor
	pcm.Root = cm.root.Bytes()
	pcm.Changeset = changeset

	// Check if it's the initial commit:
	var parentHash []byte
	if cm.parent != nil {
		parentHash = cm.parent.Bytes()
	}

	pcm.Parent = parentHash

	hashBytes := cm.hash.Bytes()
	if len(hashBytes) == 0 {
		hashBytes = EmptyHash.Bytes()
	}

	// TODO: Store something more meaningful in 'name':
	return &wire.Node{
		Type:     wire.NodeType_COMMIT,
		NodeSize: 0,
		ModTime:  modTime,
		Hash:     hashBytes,
		Name:     "commit",
		ID:       cm.id,
		Commit:   pcm,
	}, nil
}

/////////////////// METADATA INTERFACE ///////////////////

func (cm *Commit) Name() string {
	return cm.hash.B58String()
}

func (cm *Commit) Path() string {
	return prefixSlash(path.Join(".snapshots", cm.Name()))
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

func (cm *Commit) Finalize(author *Author, message string, parent *Commit) error {
	cm.message = message

	if parent != nil {
		if err := cm.SetParent(parent); err != nil {
			return err
		}
	}

	// This is inefficient, but is supposed to be easy to understand
	// while this is still playground stuff.
	s := ""
	s += fmt.Sprintf("Message: %s\n", cm.message)
	s += fmt.Sprintf("Author:  %s\n", cm.author)
	if parent != nil {
		s += fmt.Sprintf("Parent:  %s\n", parent.Hash().B58String())
	}

	hash := cm.root.Clone()
	if err := hash.MixIn([]byte(s)); err != nil {
		return err
	}

	cm.hash = hash
	cm.modTime = time.Now()
	return nil
}
