package server

import (
	"io"
	"io/ioutil"
	"os"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/containerd/fifo"
	"github.com/disorganizer/brig/brigd/capnp"
	"github.com/disorganizer/brig/catfs"
	capnplib "zombiezen.com/go/capnproto2"
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

	capInfo.SetSize(info.Size)
	capInfo.SetInode(info.Inode)
	capInfo.SetIsDir(info.IsDir)
	capInfo.SetDepth(int32(info.Depth))
	return &capInfo, nil
}

func (fh *fsHandler) withOwnFs(fn func(fs *catfs.FS) error) error {
	rp, err := fh.base.Repo()
	if err != nil {
		return err
	}

	bk, err := fh.base.Backend()
	if err != nil {
		return err
	}

	fs, err := rp.OwnFS(bk)
	if err != nil {
		return err
	}

	return fn(fs)
}

////////////////////////////////////
// ACTUAL HANDLER IMPLEMENTATIONS //
////////////////////////////////////

func (fh *fsHandler) List(call capnp.FS_list) error {
	return fh.withOwnFs(func(fs *catfs.FS) error {
		// Collect list params:
		root, err := call.Params.Root()
		if err != nil {
			return err
		}

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
	return fh.withOwnFs(func(fs *catfs.FS) error {
		repoPath, err := call.Params.RepoPath()
		if err != nil {
			return err
		}

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
	path, err := call.Params.Path()
	if err != nil {
		return err
	}

	return fh.withOwnFs(func(fs *catfs.FS) error {
		stream, err := fs.Cat(path)
		if err != nil {
			return err
		}

		// TODO: It's kinda pointless to open a file just to get it's name.
		//       Think of a better naming strategy.
		tempFile, err := ioutil.TempFile("", "brig-fifo")
		if err != nil {
			return err
		}

		fifoPath := tempFile.Name()
		if err := tempFile.Close(); err != nil {
			return err
		}

		flags := syscall.O_CREAT | syscall.O_NONBLOCK | syscall.O_WRONLY
		fifoFd, err := fifo.OpenFifo(call.Ctx, fifoPath, flags, 0644)
		if err != nil {
			return err
		}

		if err := call.Results.SetFifoPath(fifoPath); err != nil {
			return err
		}

		go func() {
			if _, err := io.Copy(fifoFd, stream); err != nil {
				log.Warningf(
					"Failed to copy contents of `%s` to fifo (%s): %v",
					path,
					fifoPath,
					err,
				)
			}

			if err := stream.Close(); err != nil {
				log.Warningf("Failed to close stream for `%s`: %v", path, err)
			}

			if err := fifoFd.Close(); err != nil {
				log.Warningf("Failed to close fifo at %s for %s: %v", fifoPath, path, err)
			}
		}()

		return nil
	})
}

func (fh *fsHandler) Mkdir(call capnp.FS_mkdir) error {
	path, err := call.Params.Path()
	if err != nil {
		return err
	}

	createParents := call.Params.CreateParents()
	return fh.withOwnFs(func(fs *catfs.FS) error {
		return fs.Mkdir(path, createParents)
	})
}

func (fh *fsHandler) Remove(call capnp.FS_remove) error {
	path, err := call.Params.Path()
	if err != nil {
		return err
	}

	return fh.withOwnFs(func(fs *catfs.FS) error {
		return fs.Remove(path)
	})
}

func (fh *fsHandler) Move(call capnp.FS_move) error {
	srcPath, err := call.Params.SrcPath()
	if err != nil {
		return err
	}

	dstPath, err := call.Params.DstPath()
	if err != nil {
		return err
	}

	return fh.withOwnFs(func(fs *catfs.FS) error {
		return fs.Move(srcPath, dstPath)
	})
}

// TODO: Move to vcs.

func (fh *fsHandler) Log(call capnp.FS_log) error {
	seg := call.Results.Segment()

	return fh.withOwnFs(func(fs *catfs.FS) error {
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

			tagList, err := capnplib.NewTextList(seg, int32(len(entry.Tags)))
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

func (fh *fsHandler) Commit(call capnp.FS_commit) error {
	msg, err := call.Params.Msg()
	if err != nil {
		return err
	}

	return fh.withOwnFs(func(fs *catfs.FS) error {
		return fs.MakeCommit(msg)
	})
}
