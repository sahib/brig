package nodes

import (
	"fmt"

	capnp_model "github.com/disorganizer/brig/cafs/nodes/capnp"
	h "github.com/disorganizer/brig/util/hashlib"
	capnp "zombiezen.com/go/capnproto2"
)

// Ghost is a special kind of Node that marks a moved node.
// If a file was moved, a ghost will be created for the old place.
// If another file is moved to the new place, the ghost will be "resurrected"
// with the new content.
type Ghost struct {
	ModNode

	ghostInode uint64
	oldType    NodeType
}

// MakeGhost takes an existing node and converts it to a ghost.
// In the ghost form no metadata is lost, but the node should
// not show up. `inode` will be the new inode of the ghost.
// It should differ to the previous node.
func MakeGhost(nd ModNode, inode uint64) (*Ghost, error) {
	return &Ghost{
		ModNode:    nd.Copy(),
		oldType:    nd.Type(),
		ghostInode: inode,
	}, nil
}

// Type always returns NodeTypeGhost
func (g *Ghost) Type() NodeType {
	return NodeTypeGhost
}

func (g *Ghost) OldNode() Node {
	return g.ModNode
}

func (g *Ghost) OldFile() (*File, error) {
	file, ok := g.ModNode.(*File)
	if !ok {
		return nil, ErrBadNode
	}

	return file, nil
}

func (g *Ghost) OldDirectory() (*Directory, error) {
	directory, ok := g.ModNode.(*Directory)
	if !ok {
		return nil, ErrBadNode
	}

	return directory, nil
}

func (g *Ghost) String() string {
	return fmt.Sprintf("<ghost: %s %v>", g.Hash(), g.ModNode)
}

func (g *Ghost) Hash() h.Hash {
	return h.Sum([]byte(fmt.Sprintf("ghost:%s", g.ModNode.Hash())))
}

func (g *Ghost) Inode() uint64 {
	return g.ghostInode
}

// ToCapnp serializes the underlying node
func (g *Ghost) ToCapnp() (*capnp.Message, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, err
	}

	capnode, err := capnp_model.NewRootNode(seg)
	if err != nil {
		return nil, err
	}

	var base *Base

	capghost, err := capnode.NewGhost()
	if err != nil {
		return nil, err
	}

	capghost.SetGhostInode(g.ghostInode)

	switch g.oldType {
	case NodeTypeFile:
		file, ok := g.ModNode.(*File)
		if !ok {
			return nil, ErrBadNode
		}

		capfile, err := file.setFileAttrs(seg)
		if err != nil {
			return nil, err
		}

		base = &file.Base
		err = capghost.SetFile(*capfile)
	case NodeTypeDirectory:
		dir, ok := g.ModNode.(*Directory)
		if !ok {
			return nil, ErrBadNode
		}

		capdir, err := dir.setDirectoryAttrs(seg)
		if err != nil {
			return nil, err
		}

		base = &dir.Base
		err = capghost.SetDirectory(*capdir)
	case NodeTypeGhost:
		panic("Recursive ghosts are not possible")
	default:
		panic(fmt.Sprintf("Unknown node type: %d", g.oldType))
	}

	if err != nil {
		return nil, err
	}

	if err := base.setBaseAttrsToNode(capnode); err != nil {
		return nil, err
	}

	if err := capnode.SetGhost(capghost); err != nil {
		return nil, err
	}

	return msg, nil
}

// FromCapnp reads all attributes from a previously marshaled ghost.
func (g *Ghost) FromCapnp(msg *capnp.Message) error {
	capnode, err := capnp_model.ReadRootNode(msg)
	if err != nil {
		return err
	}

	if typ := capnode.Which(); typ != capnp_model.Node_Which_ghost {
		return fmt.Errorf("BUG: ghost unmarshal with non ghost type: %d", typ)
	}

	capghost, err := capnode.Ghost()
	if err != nil {
		return err
	}

	g.ghostInode = capghost.GhostInode()

	var base *Base

	switch typ := capghost.Which(); typ {
	case capnp_model.Ghost_Which_directory:
		capdir, err := capghost.Directory()
		if err != nil {
			return err
		}

		dir := &Directory{}
		if err := dir.readDirectoryAttr(capdir); err != nil {
			return err
		}

		g.ModNode = dir
		g.oldType = NodeTypeDirectory
		base = &dir.Base
	case capnp_model.Ghost_Which_file:
		capfile, err := capghost.File()
		if err != nil {
			return err
		}

		file := &File{}
		if err := file.readFileAttrs(capfile); err != nil {
			return err
		}

		g.ModNode = file
		g.oldType = NodeTypeFile
		base = &file.Base
	default:
		return ErrBadNode
	}

	if err := base.parseBaseAttrsFromNode(capnode); err != nil {
		return err
	}

	return nil
}
