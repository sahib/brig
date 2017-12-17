package server

import (
	"os"
	"strings"

	"github.com/sahib/brig/catfs"
	"github.com/sahib/brig/server/capnp"
	capnplib "zombiezen.com/go/capnproto2"
	"zombiezen.com/go/capnproto2/server"
)

func prefixSlash(s string) string {
	if !strings.HasPrefix(s, "/") {
		return "/" + s
	}

	return s
}

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

	return fh.base.withCurrFs(func(fs *catfs.FS) error {
		// Collect list params:
		root, err := call.Params.Root()
		if err != nil {
			return err
		}

		root = prefixSlash(root)

		maxDepth := call.Params.MaxDepth()

		// Call List()
		entries, err := fs.List(root, int(maxDepth))
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

	return fh.base.withCurrFs(func(fs *catfs.FS) error {
		repoPath, err := call.Params.RepoPath()
		if err != nil {
			return err
		}

		repoPath = prefixSlash(repoPath)

		localPath, err := call.Params.LocalPath()
		if err != nil {
			return err
		}

		fd, err := os.Open(localPath)
		if err != nil {
			return err
		}

		defer fd.Close()

		return fs.Stage(repoPath, fd)
	})
}

func (fh *fsHandler) Cat(call capnp.FS_cat) error {
	server.Ack(call.Options)

	path, err := call.Params.Path()
	if err != nil {
		return err
	}

	path = prefixSlash(path)
	return fh.base.withCurrFs(func(fs *catfs.FS) error {
		port, err := bootTransferServer(fs, path)
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

	path = prefixSlash(path)
	createParents := call.Params.CreateParents()
	return fh.base.withCurrFs(func(fs *catfs.FS) error {
		return fs.Mkdir(path, createParents)
	})
}

func (fh *fsHandler) Remove(call capnp.FS_remove) error {
	server.Ack(call.Options)

	path, err := call.Params.Path()
	if err != nil {
		return err
	}

	path = prefixSlash(path)
	return fh.base.withCurrFs(func(fs *catfs.FS) error {
		return fs.Remove(path)
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

	srcPath = prefixSlash(srcPath)
	dstPath = prefixSlash(dstPath)
	return fh.base.withCurrFs(func(fs *catfs.FS) error {
		return fs.Move(srcPath, dstPath)
	})
}

func (fh *fsHandler) Pin(call capnp.FS_pin) error {
	server.Ack(call.Options)

	path, err := call.Params.Path()
	if err != nil {
		return err
	}

	path = prefixSlash(path)
	return fh.base.withCurrFs(func(fs *catfs.FS) error {
		return fs.Pin(path)
	})
}

func (fh *fsHandler) Unpin(call capnp.FS_unpin) error {
	server.Ack(call.Options)

	path, err := call.Params.Path()
	if err != nil {
		return err
	}

	path = prefixSlash(path)
	return fh.base.withCurrFs(func(fs *catfs.FS) error {
		return fs.Unpin(path)
	})
}

func (fh *fsHandler) Stat(call capnp.FS_stat) error {
	server.Ack(call.Options)

	path, err := call.Params.Path()
	if err != nil {
		return err
	}

	path = prefixSlash(path)
	return fh.base.withCurrFs(func(fs *catfs.FS) error {
		info, err := fs.Stat(path)
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
