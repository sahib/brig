package vcs

import (
	"path"
	"sort"

	e "github.com/pkg/errors"
	c "github.com/sahib/brig/catfs/core"
	ie "github.com/sahib/brig/catfs/errors"
	n "github.com/sahib/brig/catfs/nodes"
	capnp_patch "github.com/sahib/brig/catfs/vcs/capnp"
	"github.com/sahib/brig/util/trie"
	capnp "zombiezen.com/go/capnproto2"
)

type Patch struct {
	FromIndex int64
	CurrIndex int64
	Changes   []*Change
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

	capPatch.SetFromIndex(p.FromIndex)
	capPatch.SetCurrIndex(p.CurrIndex)

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

	p.FromIndex = capPatch.FromIndex()
	p.CurrIndex = capPatch.CurrIndex()

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

// buildPrefixTrie builds a trie of prefixes that can be passed
func buildPrefixTrie(prefixes []string) *trie.Node {
	root := trie.NewNode()
	for _, prefix := range prefixes {
		if prefix == "/" {
			root.Data = true
		} else {
			root.Insert(prefix).Data = true
		}
	}

	return root
}

func hasValidPrefix(root *trie.Node, path string) bool {
	if root.Data != nil && root.Data.(bool) == true {
		return true
	}

	curr := root
	for _, elem := range trie.SplitPath(path) {
		curr = curr.Lookup(elem)

		// No such children, not an allowed prefix.
		if curr == nil {
			return false
		}

		// If it's a prefix node it's over.
		if curr.Data != nil && curr.Data.(bool) == true {
			return true
		}
	}

	return false
}

func filterInvalidMoveGhost(lkr *c.Linker, child n.Node, combCh *Change, prefixTrie *trie.Node) (bool, error) {
	if child.Type() != n.NodeTypeGhost || combCh.Mask&ChangeTypeMove == 0 {
		return true, nil
	}

	moveNd, _, err := lkr.MoveEntryPoint(child)
	if err != nil {
		return false, err
	}

	if moveNd == nil {
		return false, nil
	}

	if !hasValidPrefix(prefixTrie, moveNd.Path()) {
		// The node was moved to the outside. Count it as removed.
		combCh.Mask &= ^ChangeTypeMove
		combCh.Mask |= ChangeTypeRemove
		return true, nil
	}

	return true, nil
}

func MakePatch(lkr *c.Linker, from *n.Commit, prefixes []string) (*Patch, error) {
	root, err := lkr.Root()
	if err != nil {
		return nil, err
	}

	status, err := lkr.Status()
	if err != nil {
		return nil, err
	}

	patch := &Patch{
		FromIndex: from.Index(),
		CurrIndex: status.Index(),
	}

	// Shortcut: The patch CURR..CURR would be empty.
	// No need for further computations.
	if from.TreeHash().Equal(status.TreeHash()) {
		return patch, nil
	}

	// Build a prefix trie to quickly check invalid paths.
	// This is not necessarily much faster, but runs in constant time.
	if prefixes == nil {
		prefixes = []string{"/"}
	}
	prefixTrie := buildPrefixTrie(prefixes)

	err = n.Walk(lkr, root, true, func(child n.Node) error {
		childParentPath := path.Dir(child.Path())
		if len(prefixes) != 0 && !hasValidPrefix(prefixTrie, childParentPath) {
			return nil
		}

		// We're only interested in directories if they're leaf nodes,
		// i.e. empty directories. Directories in between will be shaped
		// by the changes done to them and we do/can not recreate the
		// changes for intermediate directories easily.
		//
		// TODO: What if we move all children of a dir?
		//       We should remove the old "hull" directory.
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

		combCh := CombineChanges(changes)

		// Some special filtering needs to be done here.
		// If it'a "move" ghost we don't want to export it unless
		// the move goes outside our prefixes (which would count as "remove").
		isValid, err := filterInvalidMoveGhost(lkr, child, combCh, prefixTrie)
		if err != nil {
			return err
		}

		if isValid && combCh.Mask != 0 {
			patch.Changes = append(patch.Changes, combCh)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Make sure to apply the modifications *first* that are older.
	sort.Slice(patch.Changes, func(i, j int) bool {
		na, nb := patch.Changes[i].Curr, patch.Changes[j].Curr

		// Ghosts should sort after normal nodes.
		naIsGhost := na.Type() == n.NodeTypeGhost
		nbIsGhost := nb.Type() == n.NodeTypeGhost
		if naIsGhost != nbIsGhost {
			// sort non ghosts to the beginning
			return nbIsGhost
		}

		return na.ModTime().Before(nb.ModTime())
	})

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
