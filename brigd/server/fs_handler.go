package server

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/containerd/fifo"
	"github.com/disorganizer/brig/brigd/capnp"
	"github.com/disorganizer/brig/catfs"
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

////////////////////////////////////
// ACTUAL HANDLER IMPLEMENTATIONS //
////////////////////////////////////

func (fh *fsHandler) List(call capnp.FS_list) error {
	server.Ack(call.Options)

	return fh.base.withOwnFs(func(fs *catfs.FS) error {
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
	server.Ack(call.Options)

	return fh.base.withOwnFs(func(fs *catfs.FS) error {
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
	server.Ack(call.Options)

	path, err := call.Params.Path()
	if err != nil {
		return err
	}

	return fh.base.withOwnFs(func(fs *catfs.FS) error {
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

		ready := make(chan error)

		go func() {
			readyMarker := [1]byte{}
			if _, err := stream.Read(readyMarker[:]); err != nil {
				if err == io.EOF {
					// This is a special cases, that will probably
					// only happen for empty files.
					ready <- nil
				}

				msg := fmt.Sprintf("Failed to start stream of `%s`", path)
				log.Warning(msg)
				ready <- errors.New(msg)
				return
			}

			// Do not forget to write back the start marker:
			if _, err := fifoFd.Write(readyMarker[:]); err != nil {
				msg := fmt.Sprintf("Write start marker failed: %v", err)
				log.Warning(msg)
				ready <- errors.New(msg)
				return
			}

			// Assume that the rest of the file is still transfered o.k.
			ready <- nil

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

		// Wait with returning until the go func gets the first bytes.
		// We do this to stop a race condition between client & server,
		// where the client could read the stil-empty fifo first and
		// exit too fast afterwards.
		return <-ready
	})
}

func (fh *fsHandler) Mkdir(call capnp.FS_mkdir) error {
	server.Ack(call.Options)

	path, err := call.Params.Path()
	if err != nil {
		return err
	}

	createParents := call.Params.CreateParents()
	return fh.base.withOwnFs(func(fs *catfs.FS) error {
		return fs.Mkdir(path, createParents)
	})
}

func (fh *fsHandler) Remove(call capnp.FS_remove) error {
	server.Ack(call.Options)

	path, err := call.Params.Path()
	if err != nil {
		return err
	}

	return fh.base.withOwnFs(func(fs *catfs.FS) error {
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

	return fh.base.withOwnFs(func(fs *catfs.FS) error {
		return fs.Move(srcPath, dstPath)
	})
}

func (fh *fsHandler) Pin(call capnp.FS_pin) error {
	server.Ack(call.Options)

	path, err := call.Params.Path()
	if err != nil {
		return err
	}

	return fh.base.withOwnFs(func(fs *catfs.FS) error {
		return fs.Pin(path)
	})
}

func (fh *fsHandler) Unpin(call capnp.FS_unpin) error {
	server.Ack(call.Options)

	path, err := call.Params.Path()
	if err != nil {
		return err
	}

	return fh.base.withOwnFs(func(fs *catfs.FS) error {
		return fs.Unpin(path)
	})
}

func (fh *fsHandler) IsPinned(call capnp.FS_isPinned) error {
	server.Ack(call.Options)

	path, err := call.Params.Path()
	if err != nil {
		return err
	}

	return fh.base.withOwnFs(func(fs *catfs.FS) error {
		isPinned, err := fs.IsPinned(path)
		if err != nil {
			return err
		}

		call.Results.SetIsPinned(isPinned)
		return nil
	})
}
