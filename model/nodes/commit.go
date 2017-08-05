package nodes

import (
	"bytes"
	"fmt"
	"path"
	"time"

	capnp_model "github.com/disorganizer/brig/model/nodes/capnp"
	h "github.com/disorganizer/brig/util/hashlib"
	"github.com/multiformats/go-multihash"
	capnp "zombiezen.com/go/capnproto2"
)

// Commit groups a set of changes
type Commit struct {
	Base

	// Commit message (might be auto-generated)
	message string

	// Author is the id of the committer.
	author *Person

	// TreeHash is the hash of the root node at this point in time
	root h.Hash

	// Parent hash (only nil for initial commit)
	parent h.Hash
}

// NewCommit creates a new commit after the commit referenced by `parent`.
// `parent` might be nil for the very first commit.
func NewCommit(lkr Linker, parent h.Hash) (*Commit, error) {
	return &Commit{
		Base: Base{
			nodeType: NodeTypeCommit,
			uid:      lkr.NextUID(),
			modTime:  time.Now(),
		},
		author: AuthorOfStage(),
		parent: parent,
	}, nil
}

func (c *Commit) ToCapnp() (*capnp.Message, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, err
	}

	node, err := capnp_model.NewRootNode(seg)
	if err != nil {
		return nil, err
	}

	if err := c.setBaseAttrsToNode(node); err != nil {
		return nil, err
	}

	capcmt, err := capnp_model.NewCommit(seg)
	if err != nil {
		return nil, err
	}

	capcmt.SetMessage(c.message)
	capcmt.SetParent(c.parent)
	capcmt.SetRoot(c.root)
	node.SetCommit(capcmt)

	// TODO: Set person, without DRY.
	return msg, nil
}

func (c *Commit) FromCapnp(msg *capnp.Message) error {
	capnode, err := capnp_model.ReadRootNode(msg)
	if err != nil {
		return err
	}

	if err := c.parseBaseAttrsFromNode(capnode); err != nil {
		return err
	}

	capcmt, err := capnode.Commit()
	if err != nil {
		return err
	}

	c.message, err = capcmt.Message()
	if err != nil {
		return err
	}

	c.root, err = capcmt.Root()
	if err != nil {
		return err
	}

	c.parent, err = capcmt.Parent()
	if err != nil {
		return err
	}

	// TODO: Parse author...
	return nil
}

func (c *Commit) IsBoxed() bool {
	return c.hash != nil
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
func (c *Commit) BoxCommit(message string) error {
	if c.root == nil {
		return fmt.Errorf("Cannot box commit: root directory is empty")
	}

	buf := &bytes.Buffer{}

	// If parent == nil, this will be EmptyHash.
	buf.Write(padHash(c.parent))

	// Write the root hash.
	buf.Write(padHash(c.root))

	// Write the author hash. Different author -> different content.
	buf.Write(padHash(c.author.Hash))

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
	c.hash = h.Hash(mh)
	return nil
}

func (c *Commit) String() string {
	return fmt.Sprintf(
		"commit %s/%s <%s>",
		c.hash.B58String(),
		c.root.B58String(),
		c.message,
	)
}

// /////////////////// METADATA INTERFACE ///////////////////

func (c *Commit) Name() string {
	return c.hash.B58String()
}

func (c *Commit) Path() string {
	return prefixSlash(path.Join(".snapshots", c.Name()))
}

func (c *Commit) Size() uint64 {
	return 0
}

/////////////// HIERARCHY INTERFACE ///////////////

func (c *Commit) NChildren(lkr Linker) int {
	return 1
}

func (c *Commit) Child(lkr Linker, _ string) (Node, error) {
	// Return the root directory, no matter what name was passed.
	return lkr.NodeByHash(c.root)
}

func (c *Commit) Parent(lkr Linker) (Node, error) {
	if c.parent == nil {
		return nil, nil
	}

	return lkr.NodeByHash(c.parent)
}

func (c *Commit) SetParent(lkr Linker, nd Node) error {
	c.parent = nd.Hash().Clone()
	return nil
}

// Assert that Commit follows the Node interface:
var _ Node = &Commit{}
