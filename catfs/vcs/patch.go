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
	log "github.com/sirupsen/logrus"
	capnp "zombiezen.com/go/capnproto2"
)

// Patch is a set of changes that changed since a certain
// version of a graph.
type Patch struct {
	FromIndex int64
	CurrIndex int64
	Changes   []*Change
}

// Len returns the number of changes in the patch.
func (p *Patch) Len() int {
	return len(p.Changes)
}

func (p *Patch) Swap(i, j int) {
	p.Changes[i], p.Changes[j] = p.Changes[j], p.Changes[i]
}

func (p *Patch) Less(i, j int) bool {
	na, nb := p.Changes[i].Curr, p.Changes[j].Curr

	naIsGhost := na.Type() == n.NodeTypeGhost
	nbIsGhost := nb.Type() == n.NodeTypeGhost
	if naIsGhost != nbIsGhost {
		// Make sure ghosts are first added
		return naIsGhost
	}

	naIsDir := na.Type() == n.NodeTypeDirectory
	nbIsDir := nb.Type() == n.NodeTypeDirectory
	if naIsDir != nbIsDir {
		// Make sure that we first apply directory creation
		// and possible directory moves.
		return naIsDir
	}

	naIsRemove := p.Changes[i].Mask&ChangeTypeRemove != 0
	nbIsRemove := p.Changes[j].Mask&ChangeTypeRemove != 0
	if naIsRemove != nbIsRemove {
		// Make sure that everything is removed before
		// doing any other changes.
		return naIsRemove
	}

	naIsMove := p.Changes[i].Mask&ChangeTypeMove != 0
	nbIsMove := p.Changes[j].Mask&ChangeTypeMove != 0
	if naIsMove != nbIsMove {
		// Make sure that everything is moved before
		// doing any adds / modifcations.
		return naIsMove
	}

	return na.ModTime().Before(nb.ModTime())
}

// ToCapnp serializes a patch to capnproto message.
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

// FromCapnp deserializes `msg` into `p`.
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

// MakePatch creates a patch with all changes starting from `from`. 
// Patch will be created betweed `from` and `status` (current state)
// It will only include nodes that are located under one of the prefixes in `prefixes`.
func MakePatch(lkr *c.Linker, from *n.Commit, prefixes []string) (*Patch, error) {
	to, err := lkr.Status()
	if err != nil {
		return nil, err
	}

	return MakePatchFromTo(lkr, from, to, prefixes)
}

// Creates a patch between two commits `from` (older one)  and `to` (newer one)
func MakePatchFromTo(lkr *c.Linker, from, to *n.Commit, prefixes []string) (*Patch, error) {
	root, err := to.Child(lkr, "does not matter") // child actually means Root for commits
	if err != nil {
		return nil, err
	}

	if from == nil {
		return nil, e.New("The `from` commit is nil")
	}

	if to == nil {
		return nil, e.New("The `to` commit is nil")
	}

	patch := &Patch{
		FromIndex: from.Index(),
		CurrIndex: to.Index(),
	}

	// Shortcut: The patch CURR..CURR would be empty.
	// No need for further computations.
	if from.TreeHash().Equal(to.TreeHash()) {
		return patch, nil
	}

	// Build a prefix trie to quickly check invalid paths.
	// This is not necessarily much faster, but runs in constant time.
	if prefixes == nil {
		prefixes = []string{"/"}
	}
	prefixTrie := buildPrefixTrie(prefixes)

	err = n.Walk(lkr, root, false, func(child n.Node) error {
		childParentPath := path.Dir(child.Path())
		if len(prefixes) != 0 && !hasValidPrefix(prefixTrie, childParentPath) {
			log.Debugf("Ignoring invalid prefix: %s", childParentPath)
			return nil
		}

		// Get all changes between `to` and `from`.
		childModNode, ok := child.(n.ModNode)
		if !ok {
			return e.Wrapf(ie.ErrBadNode, "make-patch: walk")
		}

		changes, err := History(lkr, childModNode, to, from)
		if err != nil {
			return err
		}

		// No need to export empty history, abort early.
		if len(changes) == 0 {
			return nil
		}

		combCh := CombineChanges(changes)

		// Directories are a bit of a special case. We're only interested in them
		// when creating new, empty directories (n_children == 0) or if whole trees
		// were moved. In the latter case we need to also send a notice about that,
		// but we can leave out any other change.
		if child.Type() == n.NodeTypeDirectory {
			dir, ok := child.(*n.Directory)
			if !ok {
				return e.Wrapf(ie.ErrBadNode, "make-patch: dir")
			}

			if combCh.Mask&ChangeTypeMove == 0 {
				if dir.NChildren() > 0 {
					return nil
				}
			} else {
				combCh.Mask = ChangeTypeMove
			}
		}

		// Some special filtering needs to be done here. If it'a "move" ghost
		// we don't want to export it if the move goes outside our prefixes
		// (which would count as "remove").  or if we already reported a top
		// level directory that contains this move.
		isValid, err := filterInvalidMoveGhost(lkr, child, combCh, prefixTrie)
		if err != nil {
			return err
		}

		log.Debugf("combine: %v <= %v (valid %v)", combCh, changes, isValid)
		if isValid && combCh.Mask != 0 {
			patch.Changes = append(patch.Changes, combCh)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Make sure the patch is applied in the right order.
	// The receiving site will sort it again, but it's better
	// to have it in the right order already.
	sort.Sort(patch)

	for _, ch := range patch.Changes {
		log.Debugf("  change: %s", ch)
	}

	return patch, nil
}

// ApplyPatch applies the patch `p` to the linker `lkr`.
func ApplyPatch(lkr *c.Linker, p *Patch) error {
	sort.Sort(p)

	for _, change := range p.Changes {
		log.Debugf("apply %s %v", change, change.Curr.Type())
		if err := change.Replay(lkr); err != nil {
			return err
		}
	}

	return nil
}
