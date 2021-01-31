package server

import (
	"fmt"

	e "github.com/pkg/errors"
	"github.com/sahib/brig/catfs"
	"github.com/sahib/brig/server/capnp"
	cplib "zombiezen.com/go/capnproto2"
	"zombiezen.com/go/capnproto2/server"
)

type vcsHandler struct {
	base *base
}

func commitToCap(entry *catfs.Commit, seg *cplib.Segment) (*capnp.Commit, error) {
	capEntry, err := capnp.NewCommit(seg)
	if err != nil {
		return nil, err
	}

	if err := capEntry.SetHash(entry.Hash); err != nil {
		return nil, err
	}

	modTime, err := entry.Date.MarshalText()
	if err != nil {
		return nil, err
	}

	if err := capEntry.SetDate(string(modTime)); err != nil {
		return nil, err
	}

	tagList, err := cplib.NewTextList(seg, int32(len(entry.Tags)))
	if err != nil {
		return nil, err
	}

	for idx, tag := range entry.Tags {
		if err := tagList.Set(idx, tag); err != nil {
			return nil, err
		}
	}

	if err := capEntry.SetTags(tagList); err != nil {
		return nil, err
	}

	if err := capEntry.SetMsg(entry.Msg); err != nil {
		return nil, err
	}

	return &capEntry, nil
}

func (vcs *vcsHandler) Log(call capnp.VCS_log) error {
	server.Ack(call.Options)
	seg := call.Results.Segment()

	return vcs.base.withCurrFs(func(fs *catfs.FS) error {
		// TODO: Support partial logs at some point.
		// (like in gateway. currently everything is dumped.)
		entries := []*catfs.Commit{}
		err := fs.Log("", func(cmt *catfs.Commit) error {
			entries = append(entries, cmt)
			return nil
		})

		if err != nil {
			return err
		}

		lst, err := capnp.NewCommit_List(seg, int32(len(entries)))
		if err != nil {
			return err
		}

		for idx, entry := range entries {
			capEntry, err := commitToCap(entry, seg)
			if err != nil {
				return err
			}

			lst.Set(idx, *capEntry)
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

	return vcs.base.withCurrFs(func(fs *catfs.FS) error {
		msg = "user: " + msg
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

	return vcs.base.withCurrFs(func(fs *catfs.FS) error {
		return fs.Tag(rev, tagName)
	})
}

func (vcs *vcsHandler) Untag(call capnp.VCS_untag) error {
	server.Ack(call.Options)

	tagName, err := call.Params.TagName()
	if err != nil {
		return err
	}

	return vcs.base.withCurrFs(func(fs *catfs.FS) error {
		return fs.RemoveTag(tagName)
	})
}

func (vcs *vcsHandler) Reset(call capnp.VCS_reset) error {
	server.Ack(call.Options)

	rev, err := call.Params.Rev()
	if err != nil {
		return err
	}

	path, err := call.Params.Path()
	if err != nil {
		return err
	}

	// If there was no path, it means we should checkout
	// the whole commit.
	if path == "" {
		return vcs.base.withCurrFs(func(fs *catfs.FS) error {
			return fs.Checkout(rev, call.Params.Force())
		})
	}

	// Reset a specific file or directory otherwise:
	return vcs.base.withFsFromPath(path, func(url *URL, fs *catfs.FS) error {
		if err := fs.Reset(url.Path, rev); err != nil {
			return err
		}

		vcs.base.notifyFsChangeEvent()
		return nil
	})
}

func (vcs *vcsHandler) History(call capnp.VCS_history) error {
	server.Ack(call.Options)

	path, err := call.Params.Path()
	if err != nil {
		return err
	}

	seg := call.Results.Segment()

	return vcs.base.withFsFromPath(path, func(url *URL, fs *catfs.FS) error {
		history, err := fs.History(url.Path)
		if err != nil {
			return err
		}

		lst, err := capnp.NewChange_List(seg, int32(len(history)))
		if err != nil {
			return err
		}

		for idx := 0; idx < len(history); idx++ {
			entry, err := capnp.NewChange(seg)
			if err != nil {
				return err
			}

			change := history[idx]
			if err := entry.SetPath(change.Path); err != nil {
				return err
			}

			entry.SetIsPinned(change.IsPinned)
			entry.SetIsExplicit(change.IsExplicit)

			if err := entry.SetChange(change.Change); err != nil {
				return err
			}

			capHead, err := commitToCap(change.Head, seg)
			if err != nil {
				return err
			}

			if err := entry.SetHead(*capHead); err != nil {
				return err
			}

			if change.Next != nil {
				capNext, err := commitToCap(change.Next, seg)
				if err != nil {
					return err
				}

				if err := entry.SetNext(*capNext); err != nil {
					return err
				}
			}

			if err := entry.SetMovedTo(change.MovedTo); err != nil {
				return err
			}

			if err := entry.SetWasPreviouslyAt(change.WasPreviouslyAt); err != nil {
				return err
			}

			if err := lst.Set(idx, entry); err != nil {
				return err
			}
		}

		return call.Results.SetHistory(lst)
	})
}

func fillInfoLst(fs *catfs.FS, seg *cplib.Segment, infos []catfs.StatInfo) (*capnp.StatInfo_List, error) {
	lst, err := capnp.NewStatInfo_List(seg, int32(len(infos)))
	if err != nil {
		return nil, err
	}

	for idx, info := range infos {
		capInfo, err := statToCapnp(fs, &info, seg)
		if err != nil {
			return nil, err
		}

		if err := lst.Set(idx, *capInfo); err != nil {
			return nil, err
		}
	}

	return &lst, nil
}

func fillDiffPairLst(fs *catfs.FS, seg *cplib.Segment, pairs []catfs.DiffPair) (*capnp.DiffPair_List, error) {
	capLst, err := capnp.NewDiffPair_List(seg, int32(len(pairs)))
	if err != nil {
		return nil, err
	}

	for idx, pair := range pairs {
		capSrcInfo, err := statToCapnp(fs, &pair.Src, seg)
		if err != nil {
			return nil, err
		}

		capDstInfo, err := statToCapnp(fs, &pair.Dst, seg)
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

func diffToCapnpDiff(fs *catfs.FS, seg *cplib.Segment, diff *catfs.Diff) (*capnp.Diff, error) {
	capDiff, err := capnp.NewDiff(seg)
	if err != nil {
		return nil, err
	}

	addedLst, err := fillInfoLst(fs, seg, diff.Added)
	if err != nil {
		return nil, err
	}

	if err := capDiff.SetAdded(*addedLst); err != nil {
		return nil, err
	}

	removedLst, err := fillInfoLst(fs, seg, diff.Removed)
	if err != nil {
		return nil, err
	}

	if err := capDiff.SetRemoved(*removedLst); err != nil {
		return nil, err
	}

	missingLst, err := fillInfoLst(fs, seg, diff.Missing)
	if err != nil {
		return nil, err
	}

	if err := capDiff.SetMissing(*missingLst); err != nil {
		return nil, err
	}

	ignoredLst, err := fillInfoLst(fs, seg, diff.Ignored)
	if err != nil {
		return nil, err
	}

	if err := capDiff.SetIgnored(*ignoredLst); err != nil {
		return nil, err
	}

	mergedLst, err := fillDiffPairLst(fs, seg, diff.Merged)
	if err != nil {
		return nil, err
	}

	if err := capDiff.SetMerged(*mergedLst); err != nil {
		return nil, err
	}

	movedLst, err := fillDiffPairLst(fs, seg, diff.Moved)
	if err != nil {
		return nil, err
	}

	if err := capDiff.SetMoved(*movedLst); err != nil {
		return nil, err
	}

	conflictLst, err := fillDiffPairLst(fs, seg, diff.Conflict)
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

	localOwner, err := call.Params.LocalOwner()
	if err != nil {
		return err
	}

	remoteOwner, err := call.Params.RemoteOwner()
	if err != nil {
		return err
	}

	rp := vcs.base.repo
	if call.Params.NeedFetch() {
		if err := vcs.base.doFetch(remoteOwner); err != nil {
			return e.Wrapf(err, "fetch-remote")
		}

		if err := vcs.base.doFetch(localOwner); err != nil {
			return e.Wrapf(err, "fetch-local")
		}
	}

	// Check if the stores are valid:
	for _, owner := range []string{localOwner, remoteOwner} {
		if !rp.HaveFS(owner) {
			return fmt.Errorf("We do not have data for `%s`", owner)
		}
	}

	localRev, err := call.Params.LocalRev()
	if err != nil {
		return err
	}

	remoteRev, err := call.Params.RemoteRev()
	if err != nil {
		return err
	}

	return vcs.base.withRemoteFs(localOwner, func(localFs *catfs.FS) error {
		return vcs.base.withRemoteFs(remoteOwner, func(remoteFs *catfs.FS) error {
			diff, err := localFs.MakeDiff(remoteFs, localRev, remoteRev)
			if err != nil {
				return err
			}

			capDiff, err := diffToCapnpDiff(localFs, call.Results.Segment(), diff)
			if err != nil {
				return err
			}

			return call.Results.SetDiff(*capDiff)
		})
	})
}

func (vcs *vcsHandler) Fetch(call capnp.VCS_fetch) error {
	server.Ack(call.Options)

	who, err := call.Params.Who()
	if err != nil {
		return err
	}

	return vcs.base.doFetch(who)
}

func (vcs *vcsHandler) Sync(call capnp.VCS_sync) error {
	server.Ack(call.Options)

	withWhom, err := call.Params.WithWhom()
	if err != nil {
		return err
	}

	diff, err := vcs.base.doSync(withWhom, call.Params.NeedFetch(), "")
	if err != nil {
		return err
	}

	return vcs.base.withCurrFs(func(localFs *catfs.FS) error {
		capDiff, err := diffToCapnpDiff(localFs, call.Results.Segment(), diff)
		if err != nil {
			return err
		}

		return call.Results.SetDiff(*capDiff)
	})
}

func (vcs *vcsHandler) CommitInfo(call capnp.VCS_commitInfo) error {
	server.Ack(call.Options)

	rev, err := call.Params.Rev()
	if err != nil {
		return err
	}

	return vcs.base.withCurrFs(func(fs *catfs.FS) error {
		cmt, err := fs.CommitInfo(rev)
		if err != nil {
			return err
		}

		call.Results.SetIsValidRef(cmt != nil)
		if cmt != nil {
			capCmt, err := commitToCap(cmt, call.Results.Segment())
			if err != nil {
				return err
			}

			if err := call.Results.SetCommit(*capCmt); err != nil {
				return err
			}
		}

		return nil
	})
}
