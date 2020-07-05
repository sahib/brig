package catfs

import (
	"sort"

	"github.com/dustin/go-humanize"
	e "github.com/pkg/errors"
	ie "github.com/sahib/brig/catfs/errors"
	n "github.com/sahib/brig/catfs/nodes"
	"github.com/sahib/brig/catfs/vcs"
	"github.com/sahib/brig/util"
	log "github.com/sirupsen/logrus"
)

type partition struct {
	PinSize uint64

	// nodes that are within min_depth and should stay pinned
	// (or are even re-pinned if needed)
	ShouldPin []n.ModNode

	// nodes that are between min_depth and max_depth.
	// they might be unpinned if they exceed the quota.
	QuotaCandidates []n.ModNode

	// nodes that are behind max_depth.
	// all of the are unpinned for sure.
	DepthCandidates []n.ModNode
}

// partitionNodeHashes takes all hashes of a node and sorts them into the
// buckets described in the partition docs.
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

		if curr.Type() == n.NodeTypeGhost {
			// ghosts nodes are always unpinned
			continue
		}

		if seen[curr.BackendHash().B58String()] {
			// We only want to have the first $n distinct versions.
			// Sometimes the versions is duplicated though (removed, readded, moved)
			// so we don't want to include them since the docs say "first 10 versions".
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

func (fs *FS) ensurePin(entries []n.ModNode) (uint64, error) {
	newlyPinned := uint64(0)
	isPinUnpinned := fs.cfg.Bool("repin.pin_unpinned")

	for _, nd := range entries {
		isPinned, _, err := fs.pinner.IsNodePinned(nd)
		if err != nil {
			return newlyPinned, err
		}
		if nd.Type() == n.NodeTypeFile {
			// let's make sure that this file node is pinned at backend as well
			isCached, err := fs.bk.IsCached(nd.BackendHash())
			if err != nil {
				return newlyPinned, err
			}
			if !isCached {
				log.Warningf("The %+v should be cached, but it is not. Recaching", nd)
				err := fs.bk.Pin(nd.BackendHash())
				if err != nil {
					return newlyPinned, err
				}
			}
		}

		if !isPinned && isPinUnpinned {
			if nd.Type() == n.NodeTypeGhost {
				// ghosts cannot be pinned
				continue
			}
			if err := fs.pinner.PinNode(nd, false); err != nil {
				return newlyPinned, err
			}

			newlyPinned += nd.Size()
		}
	}

	return newlyPinned, nil
}

func (fs *FS) ensureUnpin(entries []n.ModNode) (uint64, error) {
	savedStorage := uint64(0)

	for _, nd := range entries {
		isPinned, _, err := fs.pinner.IsNodePinned(nd)
		if err != nil {
			return 0, err
		}

		if isPinned {
			if err := fs.pinner.UnpinNode(nd, false); err != nil {
				return 0, err
			}

			savedStorage += nd.Size()
		}

	}

	return savedStorage, nil
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

func (fs *FS) balanceQuota(ps []*partition, totalStorage, quota uint64) (uint64, error) {
	sort.Slice(ps, func(i, j int) bool {
		return ps[i].PinSize < ps[j].PinSize
	})

	idx, empties := 0, 0
	savedStorage := uint64(0)

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
			return 0, err
		}

		if lastPinIdx < 0 {
			empties++
			ps[idx%len(ps)].QuotaCandidates = cnds[:0]
			continue
		}

		cnd := cnds[lastPinIdx]
		totalStorage -= cnd.Size()
		savedStorage += cnd.Size()

		if err := fs.pinner.UnpinNode(cnd, false); err != nil {
			return 0, err
		}

		ps[idx%len(ps)].QuotaCandidates = cnds[:lastPinIdx]
	}

	log.Infof("quota collector unpinned %d bytes", savedStorage)
	return savedStorage, nil
}

func (fs *FS) repin(root string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// repinning doesn't modify any metadata,
	// but still affects the filesystem.
	if fs.readOnly {
		return nil
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

	rootNd, err := fs.lkr.LookupDirectory(root)
	if err != nil {
		return err
	}

	totalStorage := uint64(0)
	addedToStorage := uint64(0)
	savedStorage := uint64(0)
	parts := []*partition{}

	log.Infof("repin started (min=%d max=%d quota=%s)", minDepth, maxDepth, quotaSrc)

	err = n.Walk(fs.lkr, rootNd, true, func(child n.Node) error {
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

		pinBytes, err := fs.ensurePin(part.ShouldPin)
		if err != nil {
			return err
		}

		unpinBytes, err := fs.ensureUnpin(part.DepthCandidates)
		if err != nil {
			return err
		}

		totalStorage += part.PinSize
		addedToStorage += pinBytes
		savedStorage += unpinBytes

		parts = append(parts, part)
		return nil
	})

	if err != nil {
		return e.Wrapf(err, "repin: walk")
	}

	quotaUnpins, err := fs.balanceQuota(parts, totalStorage, quota)
	if err != nil {
		return e.Wrapf(err, "repin: quota balance")
	}

	savedStorage += quotaUnpins
	totalStorage -= quotaUnpins

	if savedStorage >= addedToStorage{
		log.Infof("repin finished; freed %s, total storage is %s", humanize.Bytes(savedStorage-addedToStorage), humanize.Bytes(totalStorage))
	} else {
		log.Infof("repin finished; used extra %s, total storage is %s", humanize.Bytes(addedToStorage-savedStorage), humanize.Bytes(totalStorage))
	}
	return nil
}

// Repin goes over all files in the filesystem and identifies files that need to be unpinned.
// Only files that are not explicitly pinned, are touched. If a file is explicitly pinned, it will
// survive the repinning process in any case. The repinning is steered by two config variables:
//
// - fs.repin.quota: Maximum amount of pinned storage (excluding explicit pins)
// - fs.repin.depth: How many versions of a file to keep at least. This trumps quota.
//
func (fs *FS) Repin(root string) error {
	fs.repinControl <- prefixSlash(root)
	return nil
}
