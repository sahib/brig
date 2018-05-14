package vcs

import (
	e "github.com/pkg/errors"
	c "github.com/sahib/brig/catfs/core"
	ie "github.com/sahib/brig/catfs/errors"
	n "github.com/sahib/brig/catfs/nodes"
	capnp_model "github.com/sahib/brig/catfs/nodes/capnp"
	capnp_patch "github.com/sahib/brig/catfs/vcs/capnp"
	capnp "zombiezen.com/go/capnproto2"
)

type Patch struct {
	From    *n.Commit
	Changes []*Change
}

func (p *Patch) ToCapnp() (*capnp.Message, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, err
	}

	capPatch, err := capnp_patch.NewRootPatch(seg)
	if err != nil {
		return nil, err
	}

	capFromNd, err := capnp_model.NewNode(seg)
	if err != nil {
		return nil, err
	}

	if err := p.From.ToCapnpNode(seg, capFromNd); err != nil {
		return nil, err
	}

	if err := capPatch.SetFrom(capFromNd); err != nil {
		return nil, err
	}

	capChangeLst, err := capnp_patch.NewChange_List(seg, int32(len(p.Changes)))
	if err != nil {
		return nil, err
	}

	if err := capPatch.SetChanges(capChangeLst); err != nil {
		return nil, err
	}

	for idx, change := range p.Changes {
		capCh, err := capnp_patch.NewChange(seg)
		if err != nil {
			return nil, err
		}

		if err := change.toCapnpChange(seg, &capCh); err != nil {
			return nil, err
		}

		if err := capChangeLst.Set(idx, capCh); err != nil {
			return nil, err
		}
	}

	return msg, nil
}

func (p *Patch) FromCapnp(msg *capnp.Message) error {
	capPatch, err := capnp_patch.ReadRootPatch(msg)
	if err != nil {
		return err
	}

	capFromNd, err := capPatch.From()
	if err != nil {
		return err
	}

	p.From = &n.Commit{}
	if err := p.From.FromCapnpNode(capFromNd); err != nil {
		return err
	}

	capChs, err := capPatch.Changes()
	if err != nil {
		return err
	}

	for idx := 0; idx < capChs.Len(); idx++ {
		ch := &Change{}
		if err := ch.fromCapnpChange(capChs.At(idx)); err != nil {
			return e.Wrapf(err, "patch: from-capnp: change")
		}

		p.Changes = append(p.Changes, ch)
	}

	return nil
}

func MakePatch(lkr *c.Linker, from *n.Commit) (*Patch, error) {
	root, err := lkr.Root()
	if err != nil {
		return nil, err
	}

	status, err := lkr.Status()
	if err != nil {
		return nil, err
	}

	patch := &Patch{
		From: from,
	}

	// Shortcut: The patch CURR..CURR would be empty.
	// No need for further computations.
	if from.TreeHash().Equal(status.TreeHash()) {
		return patch, nil
	}

	err = n.Walk(lkr, root, true, func(child n.Node) error {
		// The respective ghosts on the other side
		if child.Type() == n.NodeTypeGhost {
			return nil
		}

		// We're only interested in directories if they're leaf nodes,
		// i.e. empty directories. Directories in between will be shaped
		// by the changes done to them and we do/can not recreate the
		// changes for intermediate directories easily.
		if child.Type() == n.NodeTypeDirectory {
			dir, ok := child.(*n.Directory)
			if !ok {
				return e.Wrapf(ie.ErrBadNode, "make-patch: dir")
			}

			if dir.NChildren(lkr) > 0 {
				return nil
			}
		}

		// Get all changes between status and `from`.
		childModNode, ok := child.(n.ModNode)
		if !ok {
			return e.Wrapf(ie.ErrBadNode, "make-patch: walk")
		}

		changes, err := History(lkr, childModNode, status, from)
		if err != nil {
			return err
		}

		// No need to export empty history, abort early.
		if len(changes) == 0 {
			return nil
		}

		if len(changes) > 0 {
			patch.Changes = append(patch.Changes, CombineChanges(changes))
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return patch, nil
}

func ApplyPatch(lkr *c.Linker, p *Patch) error {
	for _, change := range p.Changes {
		if err := change.Replay(lkr); err != nil {
			return err
		}
	}

	return nil
}
