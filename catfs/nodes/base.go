package nodes

import (
	"fmt"
	"strings"
	"time"

	e "github.com/pkg/errors"
	ie "github.com/sahib/brig/catfs/errors"
	capnp_model "github.com/sahib/brig/catfs/nodes/capnp"
	h "github.com/sahib/brig/util/hashlib"
	capnp "zombiezen.com/go/capnproto2"
)

// Base is a place that holds all common attributes of all Nodes.
// It also defines some utility function that will be mixed into real nodes.
type Base struct {
	// Basename of this node
	name string

	// name of the user that last modified this node
	user string

	// Hash of this node (might be empty)
	tree h.Hash

	// Pointer hash to the content in the backend
	backend h.Hash

	// Content hash of this node
	content h.Hash

	// Last modification time of this node.
	modTime time.Time

	// Type of this node
	nodeType NodeType

	// Unique identifier for this node
	inode uint64
}

// copyBase will copy all attributes from the base.
func (b *Base) copyBase(inode uint64) Base {
	return Base{
		name:     b.name,
		user:     b.user,
		tree:     b.tree.Clone(),
		content:  b.content.Clone(),
		backend:  b.backend.Clone(),
		modTime:  b.modTime,
		nodeType: b.nodeType,
		inode:    inode,
	}
}

// User returns the user that last modified this node.
func (b *Base) User() string {
	return b.user
}

// Name returns the name of this node (e.g. /a/b/c -> c)
// The root directory will have the name empty string.
func (b *Base) Name() string {
	return b.name
}

// TreeHash returns the hash of this node.
func (b *Base) TreeHash() h.Hash {
	return b.tree
}

// ContentHash returns the content hash of this node.
func (b *Base) ContentHash() h.Hash {
	return b.content
}

// BackendHash returns the backend hash of this node.
func (b *Base) BackendHash() h.Hash {
	return b.backend
}

// Type returns the type of this node.
func (b *Base) Type() NodeType {
	return b.nodeType
}

// ModTime will return the last time this node's content
// was modified. Metadata changes are not recorded.
func (b *Base) ModTime() time.Time {
	return b.modTime
}

// Inode will return a unique ID that is different for each node.
func (b *Base) Inode() uint64 {
	return b.inode
}

/////// UTILS /////////

func (b *Base) setBaseAttrsToNode(capnode capnp_model.Node) error {
	modTimeBin, err := b.modTime.MarshalBinary()
	if err != nil {
		return err
	}

	if err := capnode.SetModTime(string(modTimeBin)); err != nil {
		return err
	}
	if err := capnode.SetTreeHash(b.tree); err != nil {
		return err
	}
	if err := capnode.SetContentHash(b.content); err != nil {
		return err
	}
	if err := capnode.SetBackendHash(b.backend); err != nil {
		return err
	}
	if err := capnode.SetName(b.name); err != nil {
		return err
	}
	if err := capnode.SetUser(b.user); err != nil {
		return err
	}

	capnode.SetInode(b.inode)
	return nil
}

func (b *Base) parseBaseAttrsFromNode(capnode capnp_model.Node) error {
	var err error
	b.name, err = capnode.Name()
	if err != nil {
		return err
	}

	b.user, err = capnode.User()
	if err != nil {
		return err
	}

	b.tree, err = capnode.TreeHash()
	if err != nil {
		return err
	}

	b.content, err = capnode.ContentHash()
	if err != nil {
		return err
	}

	b.backend, err = capnode.BackendHash()
	if err != nil {
		return err
	}

	unparsedModTime, err := capnode.ModTime()
	if err != nil {
		return err
	}

	if err := b.modTime.UnmarshalBinary([]byte(unparsedModTime)); err != nil {
		return err
	}

	switch typ := capnode.Which(); typ {
	case capnp_model.Node_Which_file:
		b.nodeType = NodeTypeFile
	case capnp_model.Node_Which_directory:
		b.nodeType = NodeTypeDirectory
	case capnp_model.Node_Which_commit:
		b.nodeType = NodeTypeCommit
	case capnp_model.Node_Which_ghost:
		// Ghost set the nodeType themselves.
		// Ignore them here.
	default:
		return fmt.Errorf("Bad capnp node type `%d`", typ)
	}

	b.inode = capnode.Inode()
	return nil
}

func prefixSlash(s string) string {
	if !strings.HasPrefix(s, "/") {
		return "/" + s
	}

	return s
}

/////////////////////////////////////////
// MARSHAL HELPERS FOR ARBITRARY NODES //
/////////////////////////////////////////

// MarshalNode will convert any Node to a byte string
// Use UnmarshalNode to load a Node from it again.
func MarshalNode(nd Node) ([]byte, error) {
	msg, err := nd.ToCapnp()
	if err != nil {
		return nil, err
	}

	return msg.Marshal()
}

// UnmarshalNode will try to interpret data as a Node
func UnmarshalNode(data []byte) (Node, error) {
	msg, err := capnp.Unmarshal(data)
	if err != nil {
		return nil, err
	}

	capNd, err := capnp_model.ReadRootNode(msg)
	if err != nil {
		return nil, err
	}

	return CapNodeToNode(capNd)
}

// CapNodeToNode converts a capnproto `capNd` to a normal `Node`.
func CapNodeToNode(capNd capnp_model.Node) (Node, error) {
	// Find out the correct node struct to initialize.
	var node Node

	switch typ := capNd.Which(); typ {
	case capnp_model.Node_Which_ghost:
		node = &Ghost{}
	case capnp_model.Node_Which_file:
		node = &File{}
	case capnp_model.Node_Which_directory:
		node = &Directory{}
	case capnp_model.Node_Which_commit:
		node = &Commit{}
	default:
		return nil, fmt.Errorf("Bad capnp node type `%d`", typ)
	}

	if err := node.FromCapnpNode(capNd); err != nil {
		return nil, err
	}

	return node, nil
}

//////////////////////////
// GENERAL NODE HELPERS //
//////////////////////////

// Depth returns the depth of the node.
// It does this by looking at the path separators.
// The depth of "/" is defined as 0.
func Depth(nd Node) int {
	path := nd.Path()
	if path == "/" {
		return 0
	}

	depth := 0
	for _, rn := range path {
		if rn == '/' {
			depth++
		}
	}

	return depth
}

// RemoveNode removes `nd` from it's parent directory using `lkr`.
// Removing the root is a no-op.
func RemoveNode(lkr Linker, nd Node) error {
	parDir, err := ParentDirectory(lkr, nd)
	if err != nil {
		return err
	}

	// Cannot remove root:
	if parDir == nil {
		return nil
	}

	return parDir.RemoveChild(lkr, nd)
}

// ParentDirectory returns the parent directory of `nd`.
// For the root it will return nil.
func ParentDirectory(lkr Linker, nd Node) (*Directory, error) {
	par, err := nd.Parent(lkr)
	if err != nil {
		return nil, err
	}

	if par == nil {
		return nil, nil
	}

	parDir, ok := par.(*Directory)
	if !ok {
		return nil, ie.ErrBadNode
	}

	return parDir, nil
}

// ContentHash returns the correct content hash for `nd`.
// This also works for ghosts where the content hash is taken from the
// underlying node (ghosts themselve have no content).
func ContentHash(nd Node) (h.Hash, error) {
	switch nd.Type() {
	case NodeTypeDirectory, NodeTypeCommit, NodeTypeFile:
		return nd.ContentHash(), nil
	case NodeTypeGhost:
		ghost, ok := nd.(*Ghost)
		if !ok {
			return nil, e.Wrapf(ie.ErrBadNode, "cannot convert to ghost")
		}

		switch ghost.OldNode().Type() {
		case NodeTypeFile:
			oldFile, err := ghost.OldFile()
			if err != nil {
				return nil, err
			}

			return oldFile.ContentHash(), nil
		case NodeTypeDirectory:
			oldDirectory, err := ghost.OldDirectory()
			if err != nil {
				return nil, err
			}

			return oldDirectory.ContentHash(), nil
		}
	}

	return nil, ie.ErrBadNode
}
