package nodes

import (
	"fmt"

	capnp_model "github.com/disorganizer/brig/cafs/nodes/capnp"
	capnp "zombiezen.com/go/capnproto2"
)

// Ghost is a special kind of Node that marks a moved node.
// If a file was moved, a ghost will be created for the old place.
// If another file is moved to the new place, the ghost will be "resurrected"
// with the new content.
type Ghost struct {
	Node

	oldType NodeType
}

// MakeGhost takes an existing node and converts it to a ghost.
// In the ghost form no metadata is lost, but the node should
// not show up.
func MakeGhost(nd Node) (*Ghost, error) {
	return &Ghost{
		Node:    nd,
		oldType: nd.Type(),
	}, nil
}

// Type always returns NodeTypeGhost
func (g *Ghost) Type() NodeType {
	return NodeTypeGhost
}

func (g *Ghost) OldNode() Node {
	return g.Node
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

	switch g.oldType {
	case NodeTypeFile:
		file, ok := g.Node.(*File)
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
		dir, ok := g.Node.(*Directory)
		if !ok {
			return nil, ErrBadNode
		}

		capdir, err := dir.setDirectoryAttrs(seg)
		if err != nil {
			return nil, err
		}

		base = &dir.Base
		err = capghost.SetDirectory(*capdir)
	case NodeTypeCommit:
		cmt, ok := g.Node.(*Commit)
		if !ok {
			return nil, ErrBadNode
		}

		capcmt, err := cmt.setCommitAttrs(seg)
		if err != nil {
			return nil, err
		}

		base = &cmt.Base
		err = capghost.SetCommit(*capcmt)
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
		return fmt.Errorf("BUG: ghost unmarhsal with non ghost type: %d", typ)
	}

	capghost, err := capnode.Ghost()
	if err != nil {
		return err
	}

	switch typ := capghost.Which(); typ {
	case capnp_model.Ghost_Which_commit:
		capcmt, err := capghost.Commit()
		if err != nil {
			return err
		}

		cmt := &Commit{}
		if err := cmt.readCommitAttrs(capcmt); err != nil {
			return err
		}

		g.Node = cmt
	case capnp_model.Ghost_Which_directory:
		capdir, err := capghost.Directory()
		if err != nil {
			return err
		}

		dir := &Directory{}
		if err := dir.readDirectoryAttr(capdir); err != nil {
			return err
		}

		g.Node = dir
	case capnp_model.Ghost_Which_file:
		capfile, err := capghost.File()
		if err != nil {
			return err
		}

		file := &File{}
		if err := file.readFileAttrs(capfile); err != nil {
			return err
		}

		g.Node = file
	default:
		return ErrBadNode
	}

	return nil
}
