package server

import (
	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/brigd/capnp"
	"github.com/disorganizer/brig/catfs"
	p2pnet "github.com/disorganizer/brig/net"
	cplib "zombiezen.com/go/capnproto2"
	"zombiezen.com/go/capnproto2/server"
)

type vcsHandler struct {
	base *base
}

func (vcs *vcsHandler) Log(call capnp.VCS_log) error {
	server.Ack(call.Options)
	seg := call.Results.Segment()

	return vcs.base.withOwnFs(func(fs *catfs.FS) error {
		entries, err := fs.Log()
		if err != nil {
			return err
		}

		lst, err := capnp.NewLogEntry_List(seg, int32(len(entries)))
		if err != nil {
			return err
		}

		for idx, entry := range entries {
			capEntry, err := capnp.NewLogEntry(seg)
			if err != nil {
				return err
			}

			if err := capEntry.SetHash(entry.Hash); err != nil {
				return err
			}

			modTime, err := entry.Date.MarshalText()
			if err != nil {
				return err
			}

			log.Errorf("ENTRY %v %s", entry, modTime)

			if err := capEntry.SetDate(string(modTime)); err != nil {
				return err
			}

			tagList, err := cplib.NewTextList(seg, int32(len(entry.Tags)))
			if err != nil {
				return err
			}

			for idx, tag := range entry.Tags {
				if err := tagList.Set(idx, tag); err != nil {
					return err
				}
			}

			if err := capEntry.SetTags(tagList); err != nil {
				return err
			}

			if err := capEntry.SetMsg(entry.Msg); err != nil {
				return err
			}

			lst.Set(idx, capEntry)
		}

		return call.Results.SetEntries(lst)
	})
}

func (vcs *vcsHandler) Commit(call capnp.VCS_commit) error {
	server.Ack(call.Options)

	msg, err := call.Params.Msg()
	if err != nil {
		return err
	}

	return vcs.base.withOwnFs(func(fs *catfs.FS) error {
		return fs.MakeCommit(msg)
	})
}

func (vcs *vcsHandler) Tag(call capnp.VCS_tag) error {
	server.Ack(call.Options)

	rev, err := call.Params.Rev()
	if err != nil {
		return err
	}

	tagName, err := call.Params.TagName()
	if err != nil {
		return err
	}

	return vcs.base.withOwnFs(func(fs *catfs.FS) error {
		return fs.Tag(rev, tagName)
	})
}

func (vcs *vcsHandler) Untag(call capnp.VCS_untag) error {
	server.Ack(call.Options)

	tagName, err := call.Params.TagName()
	if err != nil {
		return err
	}

	return vcs.base.withOwnFs(func(fs *catfs.FS) error {
		return fs.RemoveTag(tagName)
	})
}

func (vcs *vcsHandler) Reset(call capnp.VCS_reset) error {
	server.Ack(call.Options)

	path, err := call.Params.Path()
	if err != nil {
		return err
	}

	rev, err := call.Params.Rev()
	if err != nil {
		return err
	}

	return vcs.base.withOwnFs(func(fs *catfs.FS) error {
		return fs.Reset(path, rev)
	})
}

func (vcs *vcsHandler) Checkout(call capnp.VCS_checkout) error {
	server.Ack(call.Options)

	rev, err := call.Params.Rev()
	if err != nil {
		return err
	}

	return vcs.base.withOwnFs(func(fs *catfs.FS) error {
		return fs.Checkout(rev, call.Params.Force())
	})
}

func (vcs *vcsHandler) History(call capnp.VCS_history) error {
	server.Ack(call.Options)

	path, err := call.Params.Path()
	if err != nil {
		return err
	}

	seg := call.Results.Segment()

	return vcs.base.withOwnFs(func(fs *catfs.FS) error {
		history, err := fs.History(path)
		if err != nil {
			return err
		}

		lst, err := capnp.NewHistoryEntry_List(seg, int32(len(history)))
		if err != nil {
			return err
		}

		for idx := 0; idx < len(history); idx++ {
			entry, err := capnp.NewHistoryEntry(seg)
			if err != nil {
				return err
			}

			if err := entry.SetPath(history[idx].Path); err != nil {
				return err
			}

			if err := entry.SetChange(history[idx].Change); err != nil {
				return err
			}

			if err := entry.SetRef(history[idx].Ref); err != nil {
				return err
			}

			if err := lst.Set(idx, entry); err != nil {
				return err
			}
		}

		return call.Results.SetHistory(lst)
	})
}

func fillInfoLst(seg *cplib.Segment, infos []catfs.StatInfo) (*capnp.StatInfo_List, error) {
	lst, err := capnp.NewStatInfo_List(seg, int32(len(infos)))
	if err != nil {
		return nil, err
	}

	for idx, info := range infos {
		capInfo, err := statToCapnp(&info, seg)
		if err != nil {
			return nil, err
		}

		if err := lst.Set(idx, *capInfo); err != nil {
			return nil, err
		}
	}

	return &lst, nil
}

func fillDiffPairLst(seg *cplib.Segment, pairs []catfs.DiffPair) (*capnp.DiffPair_List, error) {
	capLst, err := capnp.NewDiffPair_List(seg, int32(len(pairs)))
	if err != nil {
		return nil, err
	}

	for idx, pair := range pairs {
		capSrcInfo, err := statToCapnp(&pair.Src, seg)
		if err != nil {
			return nil, err
		}

		capDstInfo, err := statToCapnp(&pair.Dst, seg)
		if err != nil {
			return nil, err
		}

		capPair, err := capnp.NewDiffPair(seg)
		if err != nil {
			return nil, err
		}

		if err := capPair.SetSrc(*capSrcInfo); err != nil {
			return nil, err
		}

		if err := capPair.SetDst(*capDstInfo); err != nil {
			return nil, err
		}

		if err := capLst.Set(idx, capPair); err != nil {
			return nil, err
		}
	}

	return &capLst, nil
}

func diffToCapnpDiff(seg *cplib.Segment, diff *catfs.Diff) (*capnp.Diff, error) {
	capDiff, err := capnp.NewDiff(seg)
	if err != nil {
		return nil, err
	}

	addedLst, err := fillInfoLst(seg, diff.Added)
	if err != nil {
		return nil, err
	}

	if err := capDiff.SetAdded(*addedLst); err != nil {
		return nil, err
	}

	removedLst, err := fillInfoLst(seg, diff.Removed)
	if err != nil {
		return nil, err
	}

	if err := capDiff.SetRemoved(*removedLst); err != nil {
		return nil, err
	}

	ignoredLst, err := fillInfoLst(seg, diff.Ignored)
	if err != nil {
		return nil, err
	}

	if err := capDiff.SetIgnored(*ignoredLst); err != nil {
		return nil, err
	}

	mergedLst, err := fillDiffPairLst(seg, diff.Merged)
	if err != nil {
		return nil, err
	}

	if err := capDiff.SetMerged(*mergedLst); err != nil {
		return nil, err
	}

	conflictLst, err := fillDiffPairLst(seg, diff.Conflict)
	if err != nil {
		return nil, err
	}

	if err := capDiff.SetConflict(*conflictLst); err != nil {
		return nil, err
	}

	return &capDiff, nil
}

func (vcs *vcsHandler) MakeDiff(call capnp.VCS_makeDiff) error {
	server.Ack(call.Options)

	remoteOwner, err := call.Params.RemoteOwner()
	if err != nil {
		return err
	}

	headRevOwn, err := call.Params.HeadRevOwn()
	if err != nil {
		return err
	}

	headRevRemote, err := call.Params.HeadRevRemote()
	if err != nil {
		return err
	}

	return vcs.base.withOwnFs(func(ownFs *catfs.FS) error {
		return vcs.base.withRemoteFs(remoteOwner, func(remoteFs *catfs.FS) error {
			diff, err := ownFs.MakeDiff(remoteFs, headRevOwn, headRevRemote)
			if err != nil {
				return err
			}

			capDiff, err := diffToCapnpDiff(call.Results.Segment(), diff)
			if err != nil {
				return err
			}

			return call.Results.SetDiff(*capDiff)
		})
	})
}

func (vcs *vcsHandler) Sync(call capnp.VCS_sync) error {
	server.Ack(call.Options)
	return nil

	withWhom, err := call.Params.WithWhom()
	if err != nil {
		return err
	}

	return vcs.base.withNetClient(withWhom, func(ctl *p2pnet.Client) error {
		r, err := ctl.GetStore()
		if err != nil {
			return err
		}

		bk, err := vcs.base.Backend()
		if err != nil {
			return err
		}

		// TODO:
		// Those should be somewhat locked, so not more than
		// one sync request can be processed in parallel.

		remoteFS, err := vcs.base.repo.FS(withWhom, bk)
		if err != nil {
			return err
		}

		if err := remoteFS.Import(r); err != nil {
			return err
		}

		ownFS, err := vcs.base.repo.OwnFS(bk)
		if err != nil {
			return err
		}

		return ownFS.Sync(remoteFS)
	})
}
