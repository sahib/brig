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

	movedTo h.Hash
	oldType NodeType
}

// MakeGhost takes an existing node and converts it to a ghost.
// In the ghost form no metadata is lost, but the node should
// not show up.
func MakeGhost(nd ModNode, movedTo h.Hash) (*Ghost, error) {
	return &Ghost{
		ModNode: nd.Copy(),
		oldType: nd.Type(),
		movedTo: movedTo,
	}, nil
}

// Type always returns NodeTypeGhost
func (g *Ghost) Type() NodeType {
	return NodeTypeGhost
}

func (g *Ghost) OldNode() Node {
	return g.ModNode
}

func (g *Ghost) MovedTo(lkr Linker) (Node, error) {
	return lkr.NodeByHash(g.movedTo)
}

func (g *Ghost) OldFile() (*File, error) {
	file, ok := g.ModNode.(*File)
	if !ok {
		return nil, ErrBadNode
	}

	return file, nil
}

func (g *Ghost) String() string {
	return fmt.Sprintf("<ghost: %v>", g.ModNode)
}

func (g *Ghost) Hash() h.Hash {
	return h.Sum([]byte(fmt.Sprintf("ghost:%s", g.ModNode.Hash())))
}

// ToCapnp serializes the underlying node
func (g *Ghost) ToCapnp() (*capnp.Message, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, err
	}
	fmt.Println("Converting to ghost..")

	capnode, err := capnp_model.NewRootNode(seg)
	if err != nil {
		return nil, err
	}

	var base *Base

	capghost, err := capnode.NewGhost()
	if err != nil {
		return nil, err
	}

	if err := capghost.SetMovedTo(g.movedTo); err != nil {
		return nil, err
	}

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

	g.movedTo, err = capghost.MovedTo()
	if err != nil {
		return err
	}

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
	default:
		return ErrBadNode
	}

	return nil
}
