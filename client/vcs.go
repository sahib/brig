package client

import (
	"time"

	"github.com/sahib/brig/server/capnp"
	h "github.com/sahib/brig/util/hashlib"
)

func (cl *Client) MakeCommit(msg string) error {
	call := cl.api.Commit(cl.ctx, func(p capnp.VCS_commit_Params) error {
		return p.SetMsg(msg)
	})

	_, err := call.Struct()
	return err
}

type LogEntry struct {
	Hash h.Hash
	Msg  string
	Tags []string
	Date time.Time
}

func convertCapLogEntry(capEntry *capnp.LogEntry) (*LogEntry, error) {
	result := LogEntry{}
	modTimeStr, err := capEntry.Date()
	if err != nil {
		return nil, err
	}

	if err := result.Date.UnmarshalText([]byte(modTimeStr)); err != nil {
		return nil, err
	}

	result.Hash, err = capEntry.Hash()
	if err != nil {
		return nil, err
	}

	result.Msg, err = capEntry.Msg()
	if err != nil {
		return nil, err
	}

	tagList, err := capEntry.Tags()
	if err != nil {
		return nil, err
	}

	tags := []string{}
	for idx := 0; idx < tagList.Len(); idx++ {
		tag, err := tagList.At(idx)
		if err != nil {
			return nil, err
		}

		tags = append(tags, tag)
	}

	result.Tags = tags
	return &result, nil
}

func (cl *Client) Log() ([]LogEntry, error) {
	call := cl.api.Log(cl.ctx, func(p capnp.VCS_log_Params) error {
		return nil
	})

	results := []LogEntry{}
	result, err := call.Struct()
	if err != nil {
		return nil, err
	}

	entries, err := result.Entries()
	if err != nil {
		return nil, err
	}

	for idx := 0; idx < entries.Len(); idx++ {
		capEntry := entries.At(idx)
		result, err := convertCapLogEntry(&capEntry)
		if err != nil {
			return nil, err
		}

		results = append(results, *result)
	}

	return results, nil
}

func (cl *Client) Tag(rev, name string) error {
	call := cl.api.Tag(cl.ctx, func(p capnp.VCS_tag_Params) error {
		if err := p.SetTagName(name); err != nil {
			return err
		}

		return p.SetRev(rev)
	})

	_, err := call.Struct()
	return err
}

func (cl *Client) Untag(name string) error {
	call := cl.api.Untag(cl.ctx, func(p capnp.VCS_untag_Params) error {
		return p.SetTagName(name)
	})

	_, err := call.Struct()
	return err
}

func (cl *Client) Reset(path, rev string, force bool) error {
	call := cl.api.Reset(cl.ctx, func(p capnp.VCS_reset_Params) error {
		if err := p.SetPath(path); err != nil {
			return err
		}

		p.SetForce(force)
		return p.SetRev(rev)
	})

	_, err := call.Struct()
	return err
}

type HistoryEntry struct {
	Path   string
	Change string
	Commit LogEntry
}

func (cl *Client) History(path string) ([]*HistoryEntry, error) {
	call := cl.api.History(cl.ctx, func(p capnp.VCS_history_Params) error {
		return p.SetPath(path)
	})

	result, err := call.Struct()
	if err != nil {
		return nil, err
	}

	histList, err := result.History()
	if err != nil {
		return nil, err
	}

	results := []*HistoryEntry{}
	for idx := 0; idx < histList.Len(); idx++ {
		entry := histList.At(idx)
		path, err := entry.Path()
		if err != nil {
			return nil, err
		}

		change, err := entry.Change()
		if err != nil {
			return nil, err
		}

		capLogEntry, err := entry.Commit()
		if err != nil {
			return nil, err
		}

		logEntry, err := convertCapLogEntry(&capLogEntry)
		if err != nil {
			return nil, err
		}

		results = append(results, &HistoryEntry{
			Path:   path,
			Change: change,
			Commit: *logEntry,
		})
	}

	return results, nil
}

// Again: duplicated from catfs/fs.go
type DiffPair struct {
	Src StatInfo
	Dst StatInfo
}

type Diff struct {
	Added   []StatInfo
	Removed []StatInfo
	Ignored []StatInfo

	Merged   []DiffPair
	Conflict []DiffPair
}

func convertDiffList(lst capnp.StatInfo_List) ([]StatInfo, error) {
	infos := []StatInfo{}

	for idx := 0; idx < lst.Len(); idx++ {
		capInfo := lst.At(idx)
		info, err := convertCapStatInfo(&capInfo)
		if err != nil {
			return nil, err
		}

		infos = append(infos, *info)
	}

	return infos, nil
}

func convertDiffPairList(lst capnp.DiffPair_List) ([]DiffPair, error) {
	pairs := []DiffPair{}
	for idx := 0; idx < lst.Len(); idx++ {
		capPair := lst.At(idx)
		capSrc, err := capPair.Src()
		if err != nil {
			return nil, err
		}

		capDst, err := capPair.Dst()
		if err != nil {
			return nil, err
		}

		srcInfo, err := convertCapStatInfo(&capSrc)
		if err != nil {
			return nil, err
		}

		dstInfo, err := convertCapStatInfo(&capDst)
		if err != nil {
			return nil, err
		}

		pairs = append(pairs, DiffPair{
			Src: *srcInfo,
			Dst: *dstInfo,
		})
	}

	return pairs, nil
}

func convertCapDiffToDiff(capDiff capnp.Diff) (*Diff, error) {
	diff := &Diff{}

	lst, err := capDiff.Added()
	if err != nil {
		return nil, err
	}

	diff.Added, err = convertDiffList(lst)
	if err != nil {
		return nil, err
	}

	lst, err = capDiff.Removed()
	if err != nil {
		return nil, err
	}

	diff.Removed, err = convertDiffList(lst)
	if err != nil {
		return nil, err
	}

	lst, err = capDiff.Ignored()
	if err != nil {
		return nil, err
	}

	diff.Ignored, err = convertDiffList(lst)
	if err != nil {
		return nil, err
	}

	pairs, err := capDiff.Merged()
	if err != nil {
		return nil, err
	}

	diff.Merged, err = convertDiffPairList(pairs)
	if err != nil {
		return nil, err
	}

	pairs, err = capDiff.Conflict()
	if err != nil {
		return nil, err
	}

	diff.Conflict, err = convertDiffPairList(pairs)
	if err != nil {
		return nil, err
	}

	return diff, nil
}

func (cl *Client) MakeDiff(remote, headRevOwn, headRevRemote string) (*Diff, error) {
	call := cl.api.MakeDiff(cl.ctx, func(p capnp.VCS_makeDiff_Params) error {
		if err := p.SetRemoteOwner(remote); err != nil {
			return err
		}

		if err := p.SetHeadRevOwn(headRevOwn); err != nil {
			return err
		}

		return p.SetHeadRevRemote(headRevRemote)
	})

	result, err := call.Struct()
	if err != nil {
		return nil, err
	}

	capDiff, err := result.Diff()
	if err != nil {
		return nil, err
	}

	return convertCapDiffToDiff(capDiff)
}

func (ctl *Client) Fetch(remote string) error {
	call := ctl.api.Fetch(ctl.ctx, func(p capnp.VCS_fetch_Params) error {
		return p.SetWho(remote)
	})

	_, err := call.Struct()
	return err
}

func (ctl *Client) Sync(remote string, needFetch bool) error {
	call := ctl.api.Sync(ctl.ctx, func(p capnp.VCS_sync_Params) error {
		p.SetNeedFetch(needFetch)
		return p.SetWithWhom(remote)
	})

	_, err := call.Struct()
	return err
}
