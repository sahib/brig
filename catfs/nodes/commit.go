package nodes

import (
	"bytes"
	"fmt"
	"path"
	"time"

	"github.com/multiformats/go-multihash"
	capnp_model "github.com/sahib/brig/catfs/nodes/capnp"
	h "github.com/sahib/brig/util/hashlib"
	capnp "zombiezen.com/go/capnproto2"
)

const (
	// AuthorOfStage is the Person that is displayed for the stage commit.
	// Currently this is just an empty hash Person that will be set later.
	AuthorOfStage = "unknown"
)

// Commit groups a set of changes
type Commit struct {
	Base

	// Commit message (might be auto-generated)
	message string

	// Author is the id of the committer.
	author string

	// root is the tree hash of the root directory
	root h.Hash

	// Parent hash (only nil for initial commit)
	parent h.Hash

	merge struct {
		// With indicates with which person we merged.
		with string

		// head is a reference to the commit we merged with on
		// the remote side.
		head h.Hash
	}
}

// NewEmptyCommit creates a new commit after the commit referenced by `parent`.
// `parent` might be nil for the very first commit.
func NewEmptyCommit(inode uint64) (*Commit, error) {
	return &Commit{
		Base: Base{
			nodeType: NodeTypeCommit,
			inode:    inode,
			modTime:  time.Now(),
		},
		author: AuthorOfStage,
	}, nil
}

// ToCapnp will convert all commit internals to a capnp message.
func (c *Commit) ToCapnp() (*capnp.Message, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, err
	}

	capNd, err := capnp_model.NewRootNode(seg)
	if err != nil {
		return nil, err
	}

	return msg, c.ToCapnpNode(seg, capNd)
}

func (c *Commit) ToCapnpNode(seg *capnp.Segment, capNd capnp_model.Node) error {
	if err := c.setBaseAttrsToNode(capNd); err != nil {
		return err
	}

	capCmt, err := c.setCommitAttrs(seg)
	if err != nil {
		return err
	}

	return capNd.SetCommit(*capCmt)
}

func (c *Commit) setCommitAttrs(seg *capnp.Segment) (*capnp_model.Commit, error) {
	capCmt, err := capnp_model.NewCommit(seg)
	if err != nil {
		return nil, err
	}

	if err := capCmt.SetMessage(c.message); err != nil {
		return nil, err
	}
	if err := capCmt.SetRoot(c.root); err != nil {
		return nil, err
	}
	if err := capCmt.SetAuthor(c.author); err != nil {
		return nil, err
	}
	if err := capCmt.SetParent(c.parent); err != nil {
		return nil, err
	}

	// Store merge infos:
	capmerge := capCmt.Merge()

	if err := capmerge.SetWith(c.merge.with); err != nil {
		return nil, err
	}

	if err := capmerge.SetHead(c.merge.head); err != nil {
		return nil, err
	}

	return &capCmt, nil
}

// FromCapnp will set the content of `msg` into the commit,
// overwriting any previous state.
func (c *Commit) FromCapnp(msg *capnp.Message) error {
	capNd, err := capnp_model.ReadRootNode(msg)
	if err != nil {
		return err
	}

	return c.FromCapnpNode(capNd)
}

func (c *Commit) FromCapnpNode(capNd capnp_model.Node) error {
	if err := c.parseBaseAttrsFromNode(capNd); err != nil {
		return err
	}

	c.nodeType = NodeTypeCommit
	capCmt, err := capNd.Commit()
	if err != nil {
		return err
	}

	return c.readCommitAttrs(capCmt)
}

func (c *Commit) readCommitAttrs(capCmt capnp_model.Commit) error {
	var err error

	c.author, err = capCmt.Author()
	if err != nil {
		return err
	}

	c.message, err = capCmt.Message()
	if err != nil {
		return err
	}

	c.root, err = capCmt.Root()
	if err != nil {
		return err
	}

	c.parent, err = capCmt.Parent()
	if err != nil {
		return err
	}

	capmerge := capCmt.Merge()
	c.merge.head, err = capmerge.Head()
	if err != nil {
		return err
	}

	c.merge.with, err = capmerge.With()
	return err
}

// IsBoxed will return True if the ommit was already boxed
// (i.e. is a finished commit and no staging commit)
func (c *Commit) IsBoxed() bool {
	return c.tree != nil
}

// padHash will take a Hash and pad it's representation to 2048 bytes.
// This is done so we can support different hash sizes later on.
// We need fixed lengths for the hash calculation of a commit.
func padHash(hash h.Hash) []byte {
	padded := make([]byte, 2048)
	copy(padded, hash.Bytes())
	return padded
}

// Root returns the current root hash
// You shall not modify the returned hash.
func (c *Commit) Root() h.Hash {
	return c.root
}

// SetRoot sets the root directory of this commit.
func (c *Commit) SetRoot(hash h.Hash) {
	c.root = hash.Clone()
}

// BoxCommit takes all currently filled data and calculates the final hash.
// It also will update the modification time.
// Only a boxed commit should be
func (c *Commit) BoxCommit(author string, message string) error {
	if c.root == nil {
		return fmt.Errorf("Cannot box commit: root directory is empty")
	}

	c.author = author

	buf := &bytes.Buffer{}

	// If parent == nil, this will be EmptyHash.
	buf.Write(padHash(c.parent))

	// Write the root hash.
	buf.Write(padHash(c.root))

	// Write the author hash. Different author -> different content.
	buf.Write(padHash(h.Sum([]byte(c.author))))

	// Write the message last, it may be arbitary length.
	buf.Write([]byte(message))

	mh, err := multihash.Sum(
		buf.Bytes(),
		multihash.BLAKE2B_MAX,
		multihash.DefaultLengths[multihash.BLAKE2B_MAX],
	)

	if err != nil {
		return err
	}

	c.message = message
	c.tree = h.Hash(mh)
	return nil
}

// String will return a nice representation of a commit.
func (c *Commit) String() string {
	return fmt.Sprintf(
		"<commit %s (%s)>",
		c.tree.B58String(),
		c.message,
	)
}

// SetMergeMarker remembers that we merged with the user `with`
// at this commit at `remoteHead`.
func (c *Commit) SetMergeMarker(with string, remoteHead h.Hash) {
	c.merge.with = with
	c.merge.head = remoteHead.Clone()
}

// MergeMarker returns the merge info for this commit, if any.
func (c *Commit) MergeMarker() (string, h.Hash) {
	return c.merge.with, c.merge.head
}

// /////////////////// METADATA INTERFACE ///////////////////

// Name will return the hash of the commit.
func (c *Commit) Name() string {
	return c.tree.B58String()
}

// Message will return the commit message of this commit
func (c *Commit) Message() string {
	return c.message
}

// Path will return the path of the commit, which will
func (c *Commit) Path() string {
	return prefixSlash(path.Join(".snapshots", c.Name()))
}

// Size will always return 0 since a commit has no defined size.
// If you're interested in the size of the snapshot, check the size
// of the root directory.
func (c *Commit) Size() uint64 {
	return 0
}

/////////////// HIERARCHY INTERFACE ///////////////

// NChildren will always return 1, since a commit has always exactly one
// root dir attached.
func (c *Commit) NChildren(lkr Linker) int {
	return 1
}

// Child will return the root directory, no matter what name is given.
func (c *Commit) Child(lkr Linker, _ string) (Node, error) {
	// Return the root directory, no matter what name was passed.
	return lkr.NodeByHash(c.root)
}

// Parent will return the parent commit of this node or nil
// if it is the first commit ever made.
func (c *Commit) Parent(lkr Linker) (Node, error) {
	if c.parent == nil {
		return nil, nil
	}

	return lkr.NodeByHash(c.parent)
}

// SetParent sets the parent of the commit to `nd`.
func (c *Commit) SetParent(lkr Linker, nd Node) error {
	c.parent = nd.TreeHash().Clone()
	return nil
}

// Assert that Commit follows the Node interface:
var _ Node = &Commit{}
