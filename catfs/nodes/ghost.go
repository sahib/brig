package nodes

import (
	"fmt"

	ie "github.com/sahib/brig/catfs/errors"
	capnp_model "github.com/sahib/brig/catfs/nodes/capnp"
	h "github.com/sahib/brig/util/hashlib"
	capnp "zombiezen.com/go/capnproto2"
)

// Ghost is a special kind of Node that marks a moved node.
// If a file was moved, a ghost will be created for the old place.
// If another file is moved to the new place, the ghost will be "resurrected"
// with the new content.
type Ghost struct {
	ModNode

	ghostPath  string
	ghostInode uint64
	oldType    NodeType
}

// MakeGhost takes an existing node and converts it to a ghost.
// In the ghost form no metadata is lost, but the node should
// not show up. `inode` will be the new inode of the ghost.
// It should differ to the previous node.
func MakeGhost(nd ModNode, inode uint64) (*Ghost, error) {
	if nd.Type() == NodeTypeGhost {
		panic("cannot put a ghost in a ghost")
	}

	return &Ghost{
		ModNode:    nd.Copy(nd.Inode()),
		oldType:    nd.Type(),
		ghostInode: inode,
		ghostPath:  nd.Path(),
	}, nil
}

// Type always returns NodeTypeGhost
func (g *Ghost) Type() NodeType {
	return NodeTypeGhost
}

// OldNode returns the node the ghost was when it still was alive.
func (g *Ghost) OldNode() ModNode {
	return g.ModNode
}

// OldFile returns the file the ghost was when it still was alive.
// Returns ErrBadNode when it wasn't a file.
func (g *Ghost) OldFile() (*File, error) {
	file, ok := g.ModNode.(*File)
	if !ok {
		return nil, ie.ErrBadNode
	}

	return file, nil
}

// OldDirectory returns the old directory that the node was in lifetime
// If the ghost was not a directory, ErrBadNode is returned.
func (g *Ghost) OldDirectory() (*Directory, error) {
	directory, ok := g.ModNode.(*Directory)
	if !ok {
		return nil, ie.ErrBadNode
	}

	return directory, nil
}

func (g *Ghost) String() string {
	return fmt.Sprintf("<ghost: %s %v>", g.TreeHash(), g.ModNode)
}

// Path returns the path of the node.
func (g *Ghost) Path() string {
	return g.ghostPath
}

// TreeHash returns the hash of the node.
func (g *Ghost) TreeHash() h.Hash {
	return h.Sum([]byte(fmt.Sprintf("ghost:%s", g.ModNode.TreeHash())))
}

// Inode returns the inode
func (g *Ghost) Inode() uint64 {
	return g.ghostInode
}

// SetGhostPath sets the path of the ghost.
func (g *Ghost) SetGhostPath(newPath string) {
	g.ghostPath = newPath
}

// ToCapnp serializes the underlying node
func (g *Ghost) ToCapnp() (*capnp.Message, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, err
	}

	capNd, err := capnp_model.NewRootNode(seg)
	if err != nil {
		return nil, err
	}

	return msg, g.ToCapnpNode(seg, capNd)
}

// ToCapnpNode converts this node to a serializable capnp proto node.
func (g *Ghost) ToCapnpNode(seg *capnp.Segment, capNd capnp_model.Node) error {
	var base *Base
	capghost, err := capNd.NewGhost()
	if err != nil {
		return err
	}

	capghost.SetGhostInode(g.ghostInode)
	if err = capghost.SetGhostPath(g.ghostPath); err != nil {
		return err
	}

	switch g.oldType {
	case NodeTypeFile:
		file, ok := g.ModNode.(*File)
		if !ok {
			return ie.ErrBadNode
		}

		capfile, err := file.setFileAttrs(seg)
		if err != nil {
			return err
		}

		base = &file.Base
		if err = capghost.SetFile(*capfile); err != nil {
			return err
		}
	case NodeTypeDirectory:
		dir, ok := g.ModNode.(*Directory)
		if !ok {
			return ie.ErrBadNode
		}

		capdir, err := dir.setDirectoryAttrs(seg)
		if err != nil {
			return err
		}

		base = &dir.Base
		if err = capghost.SetDirectory(*capdir); err != nil {
			return err
		}
	case NodeTypeGhost:
		panic("Recursive ghosts are not possible")
	default:
		panic(fmt.Sprintf("Unknown node type: %d", g.oldType))
	}

	if err != nil {
		return err
	}

	if err := base.setBaseAttrsToNode(capNd); err != nil {
		return err
	}

	return capNd.SetGhost(capghost)
}

// FromCapnp reads all attributes from a previously marshaled ghost.
func (g *Ghost) FromCapnp(msg *capnp.Message) error {
	capNd, err := capnp_model.ReadRootNode(msg)
	if err != nil {
		return err
	}

	return g.FromCapnpNode(capNd)
}

// FromCapnpNode converts a serialized node to a normal node.
func (g *Ghost) FromCapnpNode(capNd capnp_model.Node) error {
	if typ := capNd.Which(); typ != capnp_model.Node_Which_ghost {
		return fmt.Errorf("BUG: ghost unmarshal with non ghost type: %d", typ)
	}

	capghost, err := capNd.Ghost()
	if err != nil {
		return err
	}

	g.ghostInode = capghost.GhostInode()
	g.ghostPath, err = capghost.GhostPath()
	if err != nil {
		return err
	}

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
		return ie.ErrBadNode
	}

	return base.parseBaseAttrsFromNode(capNd)
}
