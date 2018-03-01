package client

import (
	"fmt"
	"io"
	"net"
	"time"

	"github.com/sahib/brig/server/capnp"
	h "github.com/sahib/brig/util/hashlib"
)

type StatInfo struct {
	Path     string
	User     string
	Hash     h.Hash
	Size     uint64
	Inode    uint64
	IsDir    bool
	Depth    int
	ModTime  time.Time
	IsPinned bool
	Content  h.Hash
}

func convertCapStatInfo(capInfo *capnp.StatInfo) (*StatInfo, error) {
	info := &StatInfo{}

	path, err := capInfo.Path()
	if err != nil {
		return nil, err
	}

	user, err := capInfo.User()
	if err != nil {
		return nil, err
	}

	hashBytes, err := capInfo.Hash()
	if err != nil {
		return nil, err
	}

	hash, err := h.Cast(hashBytes)
	if err != nil {
		return nil, err
	}

	contentBytes, err := capInfo.Content()
	if err != nil {
		return nil, err
	}

	content, err := h.Cast(contentBytes)
	if err != nil {
		return nil, err
	}

	modTimeData, err := capInfo.ModTime()
	if err != nil {
		return nil, err
	}

	if err := info.ModTime.UnmarshalText([]byte(modTimeData)); err != nil {
		return nil, err
	}

	info.Path = path
	info.User = user
	info.Hash = hash
	info.Size = capInfo.Size()
	info.Inode = capInfo.Inode()
	info.IsDir = capInfo.IsDir()
	info.IsPinned = capInfo.IsPinned()
	info.Depth = int(capInfo.Depth())
	info.Content = content
	return info, nil
}

func (cl *Client) List(root string, maxDepth int) ([]StatInfo, error) {
	call := cl.api.List(cl.ctx, func(p capnp.FS_list_Params) error {
		p.SetMaxDepth(int32(maxDepth))
		return p.SetRoot(root)
	})

	result, err := call.Struct()
	if err != nil {
		return nil, err
	}

	results := []StatInfo{}
	statList, err := result.Entries()
	for idx := 0; idx < statList.Len(); idx++ {
		capInfo := statList.At(idx)
		info, err := convertCapStatInfo(&capInfo)
		if err != nil {
			return nil, err
		}

		results = append(results, *info)
	}

	return results, err
}

func (cl *Client) Stage(localPath, repoPath string) error {
	call := cl.api.Stage(cl.ctx, func(p capnp.FS_stage_Params) error {
		if err := p.SetRepoPath(repoPath); err != nil {
			return err
		}

		return p.SetLocalPath(localPath)
	})

	_, err := call.Struct()
	return err
}

func (cl *Client) Cat(path string) (io.ReadCloser, error) {
	call := cl.api.Cat(cl.ctx, func(p capnp.FS_cat_Params) error {
		return p.SetPath(path)
	})

	result, err := call.Struct()
	if err != nil {
		return nil, err
	}

	port := result.Port()
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (cl *Client) Mkdir(path string, createParents bool) error {
	call := cl.api.Mkdir(cl.ctx, func(p capnp.FS_mkdir_Params) error {
		p.SetCreateParents(createParents)
		return p.SetPath(path)
	})

	_, err := call.Struct()
	return err
}

func (cl *Client) Remove(path string) error {
	call := cl.api.Remove(cl.ctx, func(p capnp.FS_remove_Params) error {
		return p.SetPath(path)
	})

	_, err := call.Struct()
	return err
}

func (cl *Client) Move(srcPath, dstPath string) error {
	call := cl.api.Move(cl.ctx, func(p capnp.FS_move_Params) error {
		if err := p.SetSrcPath(srcPath); err != nil {
			return err
		}

		return p.SetDstPath(dstPath)
	})

	_, err := call.Struct()
	return err
}

func (cl *Client) Copy(srcPath, dstPath string) error {
	call := cl.api.Copy(cl.ctx, func(p capnp.FS_copy_Params) error {
		if err := p.SetSrcPath(srcPath); err != nil {
			return err
		}

		return p.SetDstPath(dstPath)
	})

	_, err := call.Struct()
	return err
}

func (cl *Client) Pin(path string) error {
	call := cl.api.Pin(cl.ctx, func(p capnp.FS_pin_Params) error {
		return p.SetPath(path)
	})

	_, err := call.Struct()
	return err
}

func (cl *Client) Unpin(path string) error {
	call := cl.api.Unpin(cl.ctx, func(p capnp.FS_unpin_Params) error {
		return p.SetPath(path)
	})

	_, err := call.Struct()
	return err
}

func (cl *Client) Stat(path string) (*StatInfo, error) {
	call := cl.api.Stat(cl.ctx, func(p capnp.FS_stat_Params) error {
		return p.SetPath(path)
	})

	result, err := call.Struct()
	if err != nil {
		return nil, err
	}

	capInfo, err := result.Info()
	if err != nil {
		return nil, err
	}

	return convertCapStatInfo(&capInfo)
}

type GarbageItem struct {
	Path    string
	Owner   string
	Content h.Hash
}

func (cl *Client) GarbageCollect(aggressive bool) ([]*GarbageItem, error) {
	call := cl.api.GarbageCollect(cl.ctx, func(p capnp.FS_garbageCollect_Params) error {
		p.SetAggressive(aggressive)
		return nil
	})

	result, err := call.Struct()
	if err != nil {
		return nil, err
	}

	freed := []*GarbageItem{}

	capFreed, err := result.Freed()
	if err != nil {
		return nil, err
	}

	for idx := 0; idx < capFreed.Len(); idx++ {
		capGcItem := capFreed.At(idx)
		gcItem := &GarbageItem{}

		gcItem.Owner, err = capGcItem.Owner()
		if err != nil {
			return nil, err
		}

		gcItem.Path, err = capGcItem.Path()
		if err != nil {
			return nil, err
		}

		content, err := capGcItem.Content()
		if err != nil {
			return nil, err
		}

		gcItem.Content, err = h.Cast(content)
		if err != nil {
			return nil, err
		}

		freed = append(freed, gcItem)
	}

	return freed, nil
}

func (cl *Client) Touch(path string) error {
	call := cl.api.Touch(cl.ctx, func(p capnp.FS_touch_Params) error {
		return p.SetPath(path)
	})

	_, err := call.Struct()
	return err
}

func (cl *Client) Exists(path string) (bool, error) {
	call := cl.api.Exists(cl.ctx, func(p capnp.FS_exists_Params) error {
		return p.SetPath(path)
	})

	result, err := call.Struct()
	if err != nil {
		return false, err
	}

	return result.Exists(), nil
}
