package server

import (
	"fmt"
	"io"
	"net"
	"os"

	"github.com/sahib/brig/catfs"
	ie "github.com/sahib/brig/catfs/errors"
	"github.com/sahib/brig/server/capnp"
	log "github.com/sirupsen/logrus"
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

	if err := capInfo.SetTreeHash(info.TreeHash.Bytes()); err != nil {
		return nil, err
	}

	if err := capInfo.SetContentHash(info.ContentHash.Bytes()); err != nil {
		return nil, err
	}

	if err := capInfo.SetBackendHash(info.BackendHash.Bytes()); err != nil {
		return nil, err
	}

	modTime, err := info.ModTime.MarshalText()
	if err != nil {
		return nil, err
	}

	if err := capInfo.SetModTime(string(modTime)); err != nil {
		return nil, err
	}

	capInfo.SetSize(info.Size)
	capInfo.SetCachedSize(info.CachedSize)
	capInfo.SetInode(info.Inode)
	capInfo.SetIsDir(info.IsDir)
	capInfo.SetDepth(int32(info.Depth))
	capInfo.SetIsPinned(info.IsPinned)
	capInfo.SetIsExplicit(info.IsExplicit)
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

	return fh.base.withFsFromPath(root, func(url *URL, fs *catfs.FS) error {
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

	return fh.base.withFsFromPath(repoPath, func(url *URL, fs *catfs.FS) error {
		fd, err := os.Open(localPath) // #nosec
		if err != nil {
			return err
		}

		defer fd.Close()

		if err := fs.Stage(url.Path, fd); err != nil {
			return err
		}

		fh.base.notifyFsChangeEvent()
		return nil
	})
}

func (fh *fsHandler) Cat(call capnp.FS_cat) error {
	server.Ack(call.Options)

	path, err := call.Params.Path()
	if err != nil {
		return err
	}

	return fh.base.withFsFromPath(path, func(url *URL, fs *catfs.FS) error {
		if call.Params.Offline() {
			isCached, err := fs.IsCached(url.Path)
			if err != nil {
				return err
			}

			if !isCached {
				return fmt.Errorf("file is not in local cache")
			}
		}

		stream, err := fs.Cat(url.Path)
		if err != nil {
			return err
		}

		port, err := bootTransferServer(fs, fh.base.bindHost, func(conn net.Conn) {
			defer stream.Close()
			localAddr := conn.LocalAddr().String()

			n, err := io.Copy(conn, stream)
			if err != nil {
				log.Warningf("IO failed for path %s on %s: %v", path, localAddr, err)
				return
			}

			log.Infof("Wrote %d bytes of `%s` over %s", n, path, localAddr)
		})

		if err != nil {
			// Close the stream, since the copy callback was likely not called.
			stream.Close()
			return err
		}

		call.Results.SetPort(int32(port))
		return nil
	})
}

func (fh *fsHandler) Tar(call capnp.FS_tar) error {
	server.Ack(call.Options)

	path, err := call.Params.Path()
	if err != nil {
		return err
	}

	return fh.base.withFsFromPath(path, func(url *URL, fs *catfs.FS) error {
		if call.Params.Offline() {
			isCached, err := fs.IsCached(url.Path)
			if err != nil {
				return err
			}

			if !isCached {
				return fmt.Errorf("data is not in local cache")
			}
		}

		if _, err := fs.Stat(path); err != nil {
			return err
		}

		port, err := bootTransferServer(fs, fh.base.bindHost, func(conn net.Conn) {
			localAddr := conn.LocalAddr().String()
			if err := fs.Tar(path, conn, nil); err != nil {
				log.Warningf("tar failed for path %s on %s: %v", path, localAddr, err)
			}
		})

		call.Results.SetPort(int32(port))
		return err
	})
}

func (fh *fsHandler) Mkdir(call capnp.FS_mkdir) error {
	server.Ack(call.Options)

	path, err := call.Params.Path()
	if err != nil {
		return err
	}

	createParents := call.Params.CreateParents()
	return fh.base.withFsFromPath(path, func(url *URL, fs *catfs.FS) error {
		if err := fs.Mkdir(url.Path, createParents); err != nil {
			return err
		}

		fh.base.notifyFsChangeEvent()
		return nil
	})
}

func (fh *fsHandler) Remove(call capnp.FS_remove) error {
	server.Ack(call.Options)

	path, err := call.Params.Path()
	if err != nil {
		return err
	}

	return fh.base.withFsFromPath(path, func(url *URL, fs *catfs.FS) error {
		if err := fs.Remove(url.Path); err != nil {
			return err
		}

		fh.base.notifyFsChangeEvent()
		return nil
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

	dstURL, err := parsePath(dstPath)
	if err != nil {
		return err
	}

	return fh.base.withFsFromPath(srcPath, func(srcUrl *URL, fs *catfs.FS) error {
		if srcUrl.User != dstURL.User {
			return fmt.Errorf("cannot move between users: %s <-> %s", srcUrl.User, dstURL.User)
		}

		if err := fs.Move(srcUrl.Path, dstURL.Path); err != nil {
			return err
		}

		fh.base.notifyFsChangeEvent()
		return nil
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

	dstURL, err := parsePath(dstPath)
	if err != nil {
		return err
	}

	return fh.base.withFsFromPath(srcPath, func(srcUrl *URL, fs *catfs.FS) error {
		if srcUrl.User != dstURL.User {
			return fmt.Errorf("cannot copy between users: %s <-> %s", srcUrl.User, dstURL.User)
		}

		if err := fs.Copy(srcUrl.Path, dstURL.Path); err != nil {
			return err
		}

		fh.base.notifyFsChangeEvent()
		return nil
	})
}

func (fh *fsHandler) Pin(call capnp.FS_pin) error {
	server.Ack(call.Options)

	path, err := call.Params.Path()
	if err != nil {
		return err
	}

	return fh.base.withFsFromPath(path, func(url *URL, fs *catfs.FS) error {
		return fs.Pin(url.Path, "curr", true)
	})
}

func (fh *fsHandler) Unpin(call capnp.FS_unpin) error {
	server.Ack(call.Options)

	path, err := call.Params.Path()
	if err != nil {
		return err
	}

	return fh.base.withFsFromPath(path, func(url *URL, fs *catfs.FS) error {
		return fs.Unpin(url.Path, "curr", true)
	})
}

func (fh *fsHandler) Repin(call capnp.FS_repin) error {
	server.Ack(call.Options)

	path, err := call.Params.Path()
	if err != nil {
		return err
	}

	return fh.base.withCurrFs(func(fs *catfs.FS) error {
		return fs.Repin(path)
	})
}

func (fh *fsHandler) Stat(call capnp.FS_stat) error {
	server.Ack(call.Options)

	path, err := call.Params.Path()
	if err != nil {
		return err
	}

	return fh.base.withFsFromPath(path, func(url *URL, fs *catfs.FS) error {
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

	rp := fh.base.repo
	bk := fh.base.backend

	aggressive := call.Params.Aggressive()
	stats, err := rp.GC(bk, aggressive)
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

	return fh.base.withFsFromPath(path, func(url *URL, fs *catfs.FS) error {
		if err := fs.Touch(url.Path); err != nil {
			return err
		}

		fh.base.notifyFsChangeEvent()
		return nil
	})
}

func (fh *fsHandler) Exists(call capnp.FS_exists) error {
	path, err := call.Params.Path()
	if err != nil {
		return err
	}

	return fh.base.withFsFromPath(path, func(url *URL, fs *catfs.FS) error {
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

func (fh *fsHandler) Undelete(call capnp.FS_undelete) error {
	path, err := call.Params.Path()
	if err != nil {
		return err
	}

	return fh.base.withCurrFs(func(fs *catfs.FS) error {
		return fs.Undelete(path)
	})
}

func (fh *fsHandler) DeletedNodes(call capnp.FS_deletedNodes) error {
	root, err := call.Params.Root()
	if err != nil {
		return err
	}

	return fh.base.withCurrFs(func(fs *catfs.FS) error {
		nodes, err := fs.DeletedNodes(root)
		if err != nil {
			return err
		}

		lst, err := capnp.NewStatInfo_List(
			call.Results.Segment(),
			int32(len(nodes)),
		)

		if err != nil {
			return err
		}

		for idx, node := range nodes {
			capEntry, err := statToCapnp(node, call.Results.Segment())
			if err != nil {
				return err
			}

			if err := lst.Set(idx, *capEntry); err != nil {
				return err
			}
		}

		return call.Results.SetNodes(lst)
	})
}

func (fh *fsHandler) IsCached(call capnp.FS_isCached) error {
	path, err := call.Params.Path()
	if err != nil {
		return err
	}

	return fh.base.withFsFromPath(path, func(url *URL, fs *catfs.FS) error {
		isCached, err := fs.IsCached(url.Path)
		if err != nil {
			return err
		}

		call.Results.SetIsCached(isCached)
		return nil
	})
}
