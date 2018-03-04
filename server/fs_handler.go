package server

import (
	"bytes"
	"fmt"
	"net"
	"os"

	"github.com/sahib/brig/catfs"
	ie "github.com/sahib/brig/catfs/errors"
	"github.com/sahib/brig/server/capnp"
	capnplib "zombiezen.com/go/capnproto2"
	"zombiezen.com/go/capnproto2/server"
)

type fsHandler struct {
	base *base
}

func statToCapnp(info *catfs.StatInfo, seg *capnplib.Segment) (*capnp.StatInfo, error) {
	capInfo, err := capnp.NewStatInfo(seg)
	if err != nil {
		return nil, err
	}

	if err := capInfo.SetPath(info.Path); err != nil {
		return nil, err
	}

	if err := capInfo.SetUser(info.User); err != nil {
		return nil, err
	}

	if err := capInfo.SetHash(info.Hash.Bytes()); err != nil {
		return nil, err
	}

	modTime, err := info.ModTime.MarshalText()
	if err != nil {
		return nil, err
	}

	if err := capInfo.SetModTime(string(modTime)); err != nil {
		return nil, err
	}

	if err := capInfo.SetContent(info.Content.Bytes()); err != nil {
		return nil, err
	}

	capInfo.SetSize(info.Size)
	capInfo.SetInode(info.Inode)
	capInfo.SetIsDir(info.IsDir)
	capInfo.SetDepth(int32(info.Depth))
	capInfo.SetIsPinned(info.IsPinned)
	return &capInfo, nil
}

////////////////////////////////////
// ACTUAL HANDLER IMPLEMENTATIONS //
////////////////////////////////////

func (fh *fsHandler) List(call capnp.FS_list) error {
	server.Ack(call.Options)

	// Collect list params:
	root, err := call.Params.Root()
	if err != nil {
		return err
	}

	maxDepth := call.Params.MaxDepth()

	return fh.base.withFsFromPath(root, func(url *Url, fs *catfs.FS) error {
		entries, err := fs.List(url.Path, int(maxDepth))
		if err != nil {
			return err
		}

		// ...and convert results for the wire:
		lst, err := capnp.NewStatInfo_List(
			call.Results.Segment(),
			int32(len(entries)),
		)
		if err != nil {
			return err
		}

		for idx, entry := range entries {
			capEntry, err := statToCapnp(entry, call.Results.Segment())
			if err != nil {
				return err
			}

			if err := lst.Set(idx, *capEntry); err != nil {
				return err
			}
		}

		return call.Results.SetEntries(lst)
	})
}

func (fh *fsHandler) Stage(call capnp.FS_stage) error {
	server.Ack(call.Options)

	repoPath, err := call.Params.RepoPath()
	if err != nil {
		return err
	}

	localPath, err := call.Params.LocalPath()
	if err != nil {
		return err
	}

	return fh.base.withFsFromPath(repoPath, func(url *Url, fs *catfs.FS) error {
		fd, err := os.Open(localPath)
		if err != nil {
			return err
		}

		defer fd.Close()

		return fs.Stage(url.Path, fd)
	})
}

func (fh *fsHandler) StageFromData(call capnp.FS_stageFromData) error {
	server.Ack(call.Options)

	repoPath, err := call.Params.RepoPath()
	if err != nil {
		return err
	}

	port, err := bootReceiveServer(fh.base.bindHost, func(conn net.Conn) error {
		return fh.base.withFsFromPath(repoPath, func(url *Url, fs *catfs.FS) error {
			b := &bytes.Buffer{}
			if _, err := b.ReadFrom(conn); err != nil {
				return err
			}

			return fs.Stage(url.Path, bytes.NewReader(b.Bytes()))
		})
	})

	if err != nil {
		return err
	}

	call.Results.SetPort(int32(port))
	return nil
}

func (fh *fsHandler) Cat(call capnp.FS_cat) error {
	server.Ack(call.Options)

	path, err := call.Params.Path()
	if err != nil {
		return err
	}

	return fh.base.withFsFromPath(path, func(url *Url, fs *catfs.FS) error {
		port, err := bootTransferServer(fs, fh.base.bindHost, url.Path)
		if err != nil {
			return err
		}

		call.Results.SetPort(int32(port))
		return nil
	})
}

func (fh *fsHandler) Mkdir(call capnp.FS_mkdir) error {
	server.Ack(call.Options)

	path, err := call.Params.Path()
	if err != nil {
		return err
	}

	createParents := call.Params.CreateParents()
	return fh.base.withFsFromPath(path, func(url *Url, fs *catfs.FS) error {
		return fs.Mkdir(url.Path, createParents)
	})
}

func (fh *fsHandler) Remove(call capnp.FS_remove) error {
	server.Ack(call.Options)

	path, err := call.Params.Path()
	if err != nil {
		return err
	}

	return fh.base.withFsFromPath(path, func(url *Url, fs *catfs.FS) error {
		return fs.Remove(url.Path)
	})
}

func (fh *fsHandler) Move(call capnp.FS_move) error {
	server.Ack(call.Options)

	srcPath, err := call.Params.SrcPath()
	if err != nil {
		return err
	}

	dstPath, err := call.Params.DstPath()
	if err != nil {
		return err
	}

	dstUrl, err := parsePath(dstPath)
	if err != nil {
		return err
	}

	return fh.base.withFsFromPath(srcPath, func(srcUrl *Url, fs *catfs.FS) error {
		if srcUrl.User != dstUrl.User {
			return fmt.Errorf("cannot move between users: %s <-> %s", srcUrl.User, dstUrl.User)
		}

		return fs.Move(srcUrl.Path, dstUrl.Path)
	})
}

func (fh *fsHandler) Copy(call capnp.FS_copy) error {
	server.Ack(call.Options)

	srcPath, err := call.Params.SrcPath()
	if err != nil {
		return err
	}

	dstPath, err := call.Params.DstPath()
	if err != nil {
		return err
	}

	dstUrl, err := parsePath(dstPath)
	if err != nil {
		return err
	}

	return fh.base.withFsFromPath(srcPath, func(srcUrl *Url, fs *catfs.FS) error {
		if srcUrl.User != dstUrl.User {
			return fmt.Errorf("cannot copy between users: %s <-> %s", srcUrl.User, dstUrl.User)
		}

		return fs.Copy(srcUrl.Path, dstUrl.Path)
	})
}

func (fh *fsHandler) Pin(call capnp.FS_pin) error {
	server.Ack(call.Options)

	path, err := call.Params.Path()
	if err != nil {
		return err
	}

	return fh.base.withFsFromPath(path, func(url *Url, fs *catfs.FS) error {
		return fs.Pin(url.Path)
	})
}

func (fh *fsHandler) Unpin(call capnp.FS_unpin) error {
	server.Ack(call.Options)

	path, err := call.Params.Path()
	if err != nil {
		return err
	}

	return fh.base.withFsFromPath(path, func(url *Url, fs *catfs.FS) error {
		return fs.Unpin(url.Path)
	})
}

func (fh *fsHandler) Stat(call capnp.FS_stat) error {
	server.Ack(call.Options)

	path, err := call.Params.Path()
	if err != nil {
		return err
	}

	return fh.base.withFsFromPath(path, func(url *Url, fs *catfs.FS) error {
		info, err := fs.Stat(url.Path)
		if err != nil {
			return err
		}

		capInfo, err := statToCapnp(info, call.Results.Segment())
		if err != nil {
			return err
		}

		return call.Results.SetInfo(*capInfo)
	})
}

func (fh *fsHandler) GarbageCollect(call capnp.FS_garbageCollect) error {
	server.Ack(call.Options)

	repo, err := fh.base.Repo()
	if err != nil {
		return err
	}

	bk, err := fh.base.Backend()
	if err != nil {
		return err
	}

	aggressive := call.Params.Aggressive()
	stats, err := repo.GC(bk, aggressive)
	if err != nil {
		return err
	}

	gcItems := []capnp.GarbageItem{}

	for owner, subStats := range stats {
		for path, content := range subStats {
			gcItem, err := capnp.NewGarbageItem(call.Results.Segment())
			if err != nil {
				return err
			}

			if err := gcItem.SetPath(path); err != nil {
				return err
			}

			if err := gcItem.SetContent(content.Bytes()); err != nil {
				return err
			}

			if err := gcItem.SetOwner(owner); err != nil {
				return err
			}

			gcItems = append(gcItems, gcItem)
		}
	}

	freed, err := capnp.NewGarbageItem_List(
		call.Results.Segment(),
		int32(len(gcItems)),
	)

	if err != nil {
		return err
	}

	for idx := 0; idx < len(gcItems); idx++ {
		if err := freed.Set(idx, gcItems[idx]); err != nil {
			return err
		}
	}

	return call.Results.SetFreed(freed)
}

func (fh *fsHandler) Touch(call capnp.FS_touch) error {
	path, err := call.Params.Path()
	if err != nil {
		return err
	}

	return fh.base.withFsFromPath(path, func(url *Url, fs *catfs.FS) error {
		return fs.Touch(url.Path)
	})
}

func (fh *fsHandler) Exists(call capnp.FS_exists) error {
	path, err := call.Params.Path()
	if err != nil {
		return err
	}

	return fh.base.withFsFromPath(path, func(url *Url, fs *catfs.FS) error {
		_, err := fs.Stat(url.Path)

		exists := true
		if err != nil {
			if ie.IsNoSuchFileError(err) {
				exists = false
			} else {
				return err
			}
		}

		call.Results.SetExists(exists)
		return nil
	})
}
