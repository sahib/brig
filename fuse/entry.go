package fuse

import (
	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/store"
	"golang.org/x/net/context"

	// Don't panic.
	// This is just to convert a pointer to an inode.
	"unsafe"
)

// Entry is a file inside a directory.
type Entry struct {
	File *store.File
	fs   *FS
}

func (e *Entry) Attr(ctx context.Context, a *fuse.Attr) error {
	// TODO: Store special permissions? Is this allowed?
	a.Mode = 0755
	a.Size = uint64(e.File.Size)
	a.Inode = *(*uint64)(unsafe.Pointer(&e.File))
	return nil
}

func (e *Entry) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	log.Debugf("fuse-open: %s", e.File.Path())
	return &Handle{Entry: e}, nil
}

func (e *Entry) Setattr(ctx context.Context, req *fuse.SetattrRequest, resp *fuse.SetattrResponse) error {
	// TODO: Update {m,c,a}time? Maybe not needed/Unsure when this is called.
	switch {
	case req.Valid&fuse.SetattrSize != 0:
		log.Warningf("SIZE CHANGED OF %s: %d %p", e.File.Path(), req.Size, e.File)
		e.File.Lock()
		e.File.Size = store.FileSize(req.Size)
		e.File.Unlock()
	}

	return nil
}

func (e *Entry) Fsync(ctx context.Context, req *fuse.FsyncRequest) error {
	// TODO: fsync is simply ignored for now.
	return nil
}

func (e *Entry) Getxattr(ctx context.Context, req *fuse.GetxattrRequest, resp *fuse.GetxattrResponse) error {
	e.File.Lock()
	defer e.File.Unlock()

	switch req.Name {
	case "brig.hash":
		resp.Xattr = []byte(e.File.Hash.B58String())[:req.Size]
	default:
		return fuse.ErrNoXattr
	}

	return nil
}

func (e *Entry) Listxattr(ctx context.Context, req *fuse.ListxattrRequest, resp *fuse.ListxattrResponse) error {
	resp.Append("brig.hash")
	return nil
}
