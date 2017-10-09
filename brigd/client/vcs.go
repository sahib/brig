package client

import (
	"time"

	"github.com/disorganizer/brig/brigd/capnp"
	h "github.com/disorganizer/brig/util/hashlib"
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
