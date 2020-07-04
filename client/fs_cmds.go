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

// StatInfo gives information about a file or directory
// similar to the normal stat(2) call on POSIX.
type StatInfo struct {
	Path        string
	User        string
	Size        uint64
	CachedSize  uint64
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
	info.CachedSize = capInfo.CachedSize()
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

// List will list all nodes beneath and including `root` up to `maxDepth`.
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
	if err != nil {
		return nil, err
	}

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

// Stage will add a new node at `repoPath` with the contents of `localPath`.
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

// StageFromReader will create a new node at `repoPath` from the contents of `r`.
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

// Cat outputs the contents of the node at `path`.
// The node must be a file.
func (cl *Client) Cat(path string, offline bool) (io.ReadCloser, error) {
	call := cl.api.Cat(cl.ctx, func(p capnp.FS_cat_Params) error {
		p.SetOffline(offline)
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

// Tar outputs a tar archive with the contents of `path`.
// `path` can be either a file or directory.
func (cl *Client) Tar(path string, offline bool) (io.ReadCloser, error) {
	call := cl.api.Tar(cl.ctx, func(p capnp.FS_tar_Params) error {
		p.SetOffline(offline)
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

// Mkdir creates a new empty directory at `path`, possibly creating
// intermediate directories if `createParents` is set.
func (cl *Client) Mkdir(path string, createParents bool) error {
	call := cl.api.Mkdir(cl.ctx, func(p capnp.FS_mkdir_Params) error {
		p.SetCreateParents(createParents)
		return p.SetPath(path)
	})

	_, err := call.Struct()
	return err
}

// Remove removes the node at `path`.
// Directories are removed recursively.
func (cl *Client) Remove(path string) error {
	call := cl.api.Remove(cl.ctx, func(p capnp.FS_remove_Params) error {
		return p.SetPath(path)
	})

	_, err := call.Struct()
	return err
}

// Move moves the node at `srcPath` to `dstPath`.
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

// Copy copies the node at `srcPath` to `dstPath`.
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

// Pin sets an explicit pin on the node at `path`.
func (cl *Client) Pin(path string) error {
	call := cl.api.Pin(cl.ctx, func(p capnp.FS_pin_Params) error {
		return p.SetPath(path)
	})

	_, err := call.Struct()
	return err
}

// Unpin removes an explicit pin at the node at `path`.
func (cl *Client) Unpin(path string) error {
	call := cl.api.Unpin(cl.ctx, func(p capnp.FS_unpin_Params) error {
		return p.SetPath(path)
	})

	_, err := call.Struct()
	return err
}

// Repin schedules a repinning operation
func (cl *Client) Repin(root string) error {
	call := cl.api.Repin(cl.ctx, func(p capnp.FS_repin_Params) error {
		return p.SetPath(root)
	})

	_, err := call.Struct()
	return err
}

// Stat gives detailed information about the node at `path`.
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

// Touch creates a new empty file at `path`.
func (cl *Client) Touch(path string) error {
	call := cl.api.Touch(cl.ctx, func(p capnp.FS_touch_Params) error {
		return p.SetPath(path)
	})

	_, err := call.Struct()
	return err
}

// Exists tells us if a file at `path` exists.
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

// Undelete restores the deleted file at `path`.
func (cl *Client) Undelete(path string) error {
	call := cl.api.Undelete(cl.ctx, func(p capnp.FS_undelete_Params) error {
		return p.SetPath(path)
	})

	_, err := call.Struct()
	return err
}

// DeletedNodes returns a list of deleted nodes under `root`.
func (cl *Client) DeletedNodes(root string) ([]StatInfo, error) {
	call := cl.api.DeletedNodes(cl.ctx, func(p capnp.FS_deletedNodes_Params) error {
		return p.SetRoot(root)
	})

	result, err := call.Struct()
	if err != nil {
		return nil, err
	}

	capNodes, err := result.Nodes()
	if err != nil {
		return nil, err
	}

	results := []StatInfo{}
	for idx := 0; idx < capNodes.Len(); idx++ {
		capInfo := capNodes.At(idx)
		info, err := convertCapStatInfo(&capInfo)
		if err != nil {
			return nil, err
		}

		results = append(results, *info)
	}

	return results, err
}

func (cl *Client) IsCached(path string) (bool, error) {
	call := cl.api.IsCached(cl.ctx, func(p capnp.FS_isCached_Params) error {
		return p.SetPath(path)
	})

	result, err := call.Struct()
	if err != nil {
		return false, err
	}

	return result.IsCached(), nil
}
