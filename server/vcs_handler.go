package server

import (
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/server/capnp"
	"github.com/disorganizer/brig/catfs"
	fserrs "github.com/disorganizer/brig/catfs/errors"
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

	return vcs.base.withCurrFs(func(fs *catfs.FS) error {
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

	return vcs.base.withCurrFs(func(fs *catfs.FS) error {
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

	path, err := call.Params.Path()
	if err != nil {
		return err
	}

	rev, err := call.Params.Rev()
	if err != nil {
		return err
	}

	return vcs.base.withCurrFs(func(fs *catfs.FS) error {
		return fs.Reset(path, rev)
	})
}

func (vcs *vcsHandler) Checkout(call capnp.VCS_checkout) error {
	server.Ack(call.Options)

	rev, err := call.Params.Rev()
	if err != nil {
		return err
	}

	return vcs.base.withCurrFs(func(fs *catfs.FS) error {
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

	return vcs.base.withCurrFs(func(fs *catfs.FS) error {
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

	return vcs.base.withCurrFs(func(ownFs *catfs.FS) error {
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

func (vcs *vcsHandler) doFetch(who string) error {
	// TODO: Optimize by implementing store diffs.
	// This is currently implemented very stupidly by simply fetching
	// all the store from remote, saving it and using it as sync base.
	return vcs.base.withNetClient(who, func(ctl *p2pnet.Client) error {
		storeBuf, err := ctl.FetchStore()
		if err != nil {
			return err
		}

		bk, err := vcs.base.Backend()
		if err != nil {
			return err
		}

		remoteFS, err := vcs.base.repo.FS(who, bk)
		if err != nil {
			return err
		}

		return remoteFS.Import(storeBuf)
	})
}

func (vcs *vcsHandler) Fetch(call capnp.VCS_fetch) error {
	server.Ack(call.Options)

	who, err := call.Params.Who()
	if err != nil {
		return err
	}

	return vcs.doFetch(who)
}

func (vcs *vcsHandler) Sync(call capnp.VCS_sync) error {
	server.Ack(call.Options)

	withWhom, err := call.Params.WithWhom()
	if err != nil {
		return err
	}

	if call.Params.NeedFetch() {
		if err := vcs.doFetch(withWhom); err != nil {
			return err
		}
	}

	return vcs.base.withCurrFs(func(ownFs *catfs.FS) error {
		return vcs.base.withRemoteFs(withWhom, func(remoteFs *catfs.FS) error {
			// Automatically make a commit before merging with their state:
			// TODO: Check if we can also merge with CURR as starting point
			//       and only commit a merge commit if there were changes.
			timeStamp := time.Now().UTC().Format(time.RFC3339)
			commitMsg := fmt.Sprintf("sync with %s on %s", withWhom, timeStamp)
			if err = ownFs.MakeCommit(commitMsg); err != nil && err != fserrs.ErrNoChange {
				return err
			}

			return ownFs.Sync(remoteFs)
		})
	})
}
