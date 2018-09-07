package client

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"time"

	"github.com/sahib/brig/server/capnp"
	h "github.com/sahib/brig/util/hashlib"
)

type StatInfo struct {
	Path        string
	User        string
	Size        uint64
	Inode       uint64
	IsDir       bool
	Depth       int
	ModTime     time.Time
	IsPinned    bool
	IsExplicit  bool
	TreeHash    h.Hash
	ContentHash h.Hash
	BackendHash h.Hash
}

func convertHash(hashBytes []byte, err error) (h.Hash, error) {
	if err != nil {
		return nil, err
	}

	return h.Cast(hashBytes)
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

	treeHash, err := convertHash(capInfo.TreeHash())
	if err != nil {
		return nil, err
	}

	contentHash, err := convertHash(capInfo.ContentHash())
	if err != nil {
		return nil, err
	}

	backendHash, err := convertHash(capInfo.BackendHash())
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
	info.Size = capInfo.Size()
	info.Inode = capInfo.Inode()
	info.IsDir = capInfo.IsDir()
	info.IsPinned = capInfo.IsPinned()
	info.IsExplicit = capInfo.IsExplicit()
	info.Depth = int(capInfo.Depth())

	info.TreeHash = treeHash
	info.ContentHash = contentHash
	info.BackendHash = backendHash
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

func (cl *Client) StageFromReader(repoPath string, r io.Reader) error {
	fd, err := ioutil.TempFile("", "brig-stage-temp")
	if err != nil {
		return err
	}

	defer os.Remove(fd.Name())

	if _, err := io.Copy(fd, r); err != nil {
		return err
	}

	return cl.Stage(fd.Name(), repoPath)
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

type ExplicitPin struct {
	Path   string
	Commit string
}

func (cl *Client) ListExplicitPins(prefix, from, to string) ([]ExplicitPin, error) {
	call := cl.api.ListExplicitPins(cl.ctx, func(p capnp.FS_listExplicitPins_Params) error {
		if err := p.SetPrefix(prefix); err != nil {
			return err
		}

		if err := p.SetFrom(from); err != nil {
			return err
		}

		return p.SetTo(to)
	})

	result, err := call.Struct()
	if err != nil {
		return nil, err
	}

	capPins, err := result.Pins()
	if err != nil {
		return nil, err
	}

	pins := []ExplicitPin{}
	for idx := 0; idx < capPins.Len(); idx++ {
		capPin := capPins.At(idx)
		path, err := capPin.Path()
		if err != nil {
			return nil, err
		}

		commit, err := capPin.Commit()
		if err != nil {
			return nil, err
		}

		pins = append(pins, ExplicitPin{
			Path:   path,
			Commit: commit,
		})
	}

	return pins, nil
}

func (cl *Client) ClearExplicitPins(prefix, from, to string) (int, error) {
	call := cl.api.ClearExplicitPins(cl.ctx, func(p capnp.FS_clearExplicitPins_Params) error {
		if err := p.SetPrefix(prefix); err != nil {
			return err
		}

		if err := p.SetFrom(from); err != nil {
			return err
		}

		return p.SetTo(to)
	})

	result, err := call.Struct()
	if err != nil {
		return 0, err
	}

	return int(result.Count()), nil
}

func (cl *Client) SetExplicitPins(prefix, from, to string) (int, error) {
	call := cl.api.SetExplicitPins(cl.ctx, func(p capnp.FS_setExplicitPins_Params) error {
		if err := p.SetPrefix(prefix); err != nil {
			return err
		}

		if err := p.SetFrom(from); err != nil {
			return err
		}

		return p.SetTo(to)
	})

	result, err := call.Struct()
	if err != nil {
		return 0, err
	}

	return int(result.Count()), nil
}
