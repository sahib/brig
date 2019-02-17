package catfs

import (
	"fmt"
	"sort"

	"github.com/dustin/go-humanize"
	e "github.com/pkg/errors"
	ie "github.com/sahib/brig/catfs/errors"
	n "github.com/sahib/brig/catfs/nodes"
	"github.com/sahib/brig/catfs/vcs"
	"github.com/sahib/brig/util"
)

type partition struct {
	PinSize         uint64
	ShouldPin       []n.ModNode
	QuotaCandidates []n.ModNode
	DepthCandidates []n.ModNode
}

func (fs *FS) partitionNodeHashes(nd n.ModNode, minDepth, maxDepth int64) (*partition, error) {
	currDepth := int64(0)
	part := &partition{}

	curr, err := fs.lkr.Status()
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	walker := vcs.NewHistoryWalker(fs.lkr, curr, nd)

	for walker.Next() {
		state := walker.State()
		curr := state.Curr

		if seen[curr.BackendHash().B58String()] {
			// We only want to have the first $n distinct versions.
			// Sometimes the versions is duplicated though (removed, readded, moved)
			continue
		}

		// Sort the entry into the right bucket:
		if currDepth < minDepth {
			part.ShouldPin = append(part.ShouldPin, curr)
			part.PinSize += nd.Size()
		} else if currDepth >= minDepth && currDepth < maxDepth {
			part.QuotaCandidates = append(part.QuotaCandidates, curr)

			isPinned, isExplicit, err := fs.pinner.IsNodePinned(nd)
			if err != nil {
				return nil, err
			}

			if isPinned && !isExplicit {
				part.PinSize += nd.Size()
			}
		} else {
			part.DepthCandidates = append(part.DepthCandidates, curr)
		}

		seen[curr.BackendHash().B58String()] = true
		currDepth++

		// TODO: Optimization: Save depth of last run and abort early if we know
		//       that we unpinned everything at this level already.
	}

	if err := walker.Err(); err != nil {
		return nil, err
	}

	return part, nil
}

func (fs *FS) ensurePin(entries []n.ModNode) error {
	for _, nd := range entries {
		if err := fs.pinner.PinNode(nd, false); err != nil {
			return err
		}
	}

	return nil
}

func (fs *FS) ensureUnpin(entries []n.ModNode) error {
	for _, nd := range entries {
		if err := fs.pinner.UnpinNode(nd, false); err != nil {
			return err
		}
	}

	return nil
}

func findLastPinnedIdx(pinner *Pinner, nds []n.ModNode) (int, error) {
	for idx := len(nds) - 1; idx >= 0; idx-- {
		isPinned, isExplicit, err := pinner.IsNodePinned(nds[idx])
		if err != nil {
			return -1, err
		}

		if isPinned && !isExplicit {
			return idx, nil
		}
	}

	return -1, nil
}

func (fs *FS) balanceQuota(ps []*partition, totalStorage, quota uint64) error {
	sort.Slice(ps, func(i, j int) bool {
		return ps[i].PinSize < ps[j].PinSize
	})

	idx, empties := 0, 0

	// Try to reduce the pinned storage amount until
	// we stay below the determined quota.
	for totalStorage >= quota && empties < len(ps) {
		cnds := ps[idx%len(ps)].QuotaCandidates
		if len(cnds) == 0 {
			empties++
			continue
		}

		// Find the last index (i.e. earliest version) that is pinned.
		lastPinIdx, err := findLastPinnedIdx(fs.pinner, cnds)
		if err != nil {
			return err
		}

		fmt.Println("LAST PIN", lastPinIdx, len(cnds))
		if lastPinIdx < 0 {
			empties++
			ps[idx%len(ps)].QuotaCandidates = cnds[:0]
			continue
		}

		cnd := cnds[lastPinIdx]
		totalStorage -= cnd.Size()

		fmt.Println("UNPIN", cnd.Path())
		if err := fs.pinner.UnpinNode(cnd, false); err != nil {
			return err
		}

		ps[idx%len(ps)].QuotaCandidates = cnds[:lastPinIdx]
	}

	return nil
}

func (fs *FS) repin() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if fs.readOnly {
		return ErrReadOnly
	}

	if !fs.cfg.Bool("repin.enabled") {
		return nil
	}

	minDepth := util.Max64(0, fs.cfg.Int("repin.min_depth"))
	maxDepth := util.Max64(1, fs.cfg.Int("repin.max_depth"))
	quotaSrc := fs.cfg.String("repin.quota")

	quota, err := humanize.ParseBytes(quotaSrc)
	if err != nil {
		return err
	}

	root, err := fs.lkr.Root()
	if err != nil {
		return err
	}

	totalStorage := uint64(0)
	parts := []*partition{}

	err = n.Walk(fs.lkr, root, true, func(child n.Node) error {
		if child.Type() == n.NodeTypeDirectory {
			return nil
		}

		modChild, ok := child.(n.ModNode)
		if !ok {
			return e.Wrapf(ie.ErrBadNode, "repin")
		}

		part, err := fs.partitionNodeHashes(modChild, minDepth, maxDepth)
		if err != nil {
			return err
		}

		fmt.Println(child.Path())
		fmt.Println(" -", len(part.ShouldPin))
		fmt.Println(" -", len(part.QuotaCandidates))
		fmt.Println(" -", len(part.DepthCandidates))

		if err := fs.ensurePin(part.ShouldPin); err != nil {
			return err
		}

		if err := fs.ensureUnpin(part.DepthCandidates); err != nil {
			return err
		}

		totalStorage += part.PinSize
		parts = append(parts, part)
		return nil
	})

	fmt.Println("total", totalStorage)

	if err != nil {
		return err
	}

	return fs.balanceQuota(parts, totalStorage, quota)
}

// Repin goes over all files in the filesystem and identifies files that need to be unpinned.
// Only files that are not explicitly pinned, are touched. If a file is explicitly pinned, it will
// survive the repinning process in any case. The repinning is steered by two config variables:
//
// - fs.repin.quota: Maximum amount of pinned storage (excluding explicit pins)
// - fs.repin.depth: How many versions of a file to keep at least. This trumps quota.
//
func (fs *FS) Repin() error {
	fs.repinControl <- true
	return nil
}
