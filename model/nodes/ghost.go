package nodes

import (
	capnp_model "github.com/disorganizer/brig/model/nodes/capnp"
	capnp "zombiezen.com/go/capnproto2"
)

type Ghost struct {
	Base

	// oldType is the type of the file when the ghost still was alive.
	oldType NodeType
}

// MakeGhost takes an existing node and converts it to a ghost.
// In the ghost form no metadata is lost, but the node should
// not show up.
func MakeGhost(nd Node) (*Ghost, error) {
	return &Ghost{
		Base: Base{
			name:     nd.Name(),
			hash:     nd.Hash(),
			modTime:  nd.ModTime(),
			uid:      nd.Inode(),
			nodeType: NodeTypeGhost,
		},
		oldType: nd.Type(),
	}, nil
}

func (g *Ghost) OldType() NodeType {
	return g.oldType
}

func (g *Ghost) GetOldNode(lkr Linker) (Node, error) {
	return lkr.NodeByHash(g.hash)
}

func (g *Ghost) ToCapnp() (*capnp.Message, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, err
	}

	capnode, err := capnp_model.NewRootNode(seg)
	if err != nil {
		return nil, err
	}

	if err := g.setBaseAttrsToNode(capnode); err != nil {
		return nil, err
	}

	capghost, err := capnp_model.NewGhost(seg)
	if err != nil {
		return nil, err
	}

	capghost.SetOldType(uint8(g.oldType))
	capnode.SetGhost(capghost)
	return msg, nil
}

func (g *Ghost) FromCapnp(msg *capnp.Message) error {
	capnode, err := capnp_model.ReadRootNode(msg)
	if err != nil {
		return err
	}

	if err := g.parseBaseAttrsFromNode(capnode); err != nil {
		return err
	}

	capghost, err := capnode.Ghost()
	if err != nil {
		return err
	}

	g.oldType = NodeType(capghost.OldType())
	return nil
}
