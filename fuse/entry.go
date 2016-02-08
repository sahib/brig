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
	*store.File
	fs *FS
}

// Attr is called to get the stat(2) attributes of a file.
func (e *Entry) Attr(ctx context.Context, a *fuse.Attr) error {
	// TODO: Store special permissions? Is this allowed?
	a.Mode = 0755
	a.Size = uint64(e.Size())
	a.Inode = *(*uint64)(unsafe.Pointer(&e))
	return nil
}

// Open is called to get an opened handle of a file, suitable for reading and writing.
func (e *Entry) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	log.Debugf("fuse-open: %s", e.Path())
	return &Handle{Entry: e}, nil
}

// Setattr is called once an attribute of a file changes.
// Most importantly, size changes are reported here, e.g. after truncating a
// file, the size change is noticed here before Open() is called.
func (e *Entry) Setattr(ctx context.Context, req *fuse.SetattrRequest, resp *fuse.SetattrResponse) error {
	// This is called when any attribute of the file changes,
	// most importantly the file size. For example it is called when truncating
	// the file to zero bytes with a size change of `0`.
	switch {
	case req.Valid&fuse.SetattrSize != 0:
		log.Warningf("SIZE CHANGED OF %s: %d %p", e.Path(), req.Size, e)
		e.UpdateSize(int64(req.Size))
	}

	return nil
}

// Fsync is called when any open buffers need to be written to disk.
// Currently, fsync is completely ignored.
func (e *Entry) Fsync(ctx context.Context, req *fuse.FsyncRequest) error {
	return nil
}

// Getxattr is called to get a single xattr (extended attribute) of a file.
func (e *Entry) Getxattr(ctx context.Context, req *fuse.GetxattrRequest, resp *fuse.GetxattrResponse) error {
	e.Lock()
	defer e.Unlock()

	switch req.Name {
	case "brig.hash":
		resp.Xattr = []byte(e.Hash().B58String())[:req.Size]
	default:
		return fuse.ErrNoXattr
	}

	return nil
}

// Listxattr is called to list all xattrs of this file.
func (e *Entry) Listxattr(ctx context.Context, req *fuse.ListxattrRequest, resp *fuse.ListxattrResponse) error {
	resp.Append("brig.hash")
	return nil
}

// Compile time checks to see which interfaces we implement:
// Please update this list when modifying code here.
var _ = fs.Node(&Entry{})
var _ = fs.NodeFsyncer(&Entry{})
var _ = fs.NodeGetxattrer(&Entry{})
var _ = fs.NodeListxattrer(&Entry{})
var _ = fs.NodeOpener(&Entry{})
var _ = fs.NodeSetattrer(&Entry{})

//var _ = fs.NodeReadlinker(&Entry{})
//var _ = fs.NodeRemover(&Entry{})
//var _ = fs.NodeRemovexattrer(&Entry{})
// var _ = fs.NodeRenamer(&Entry{})
// var _ = fs.NodeRequestLookuper(&Entry{})
// var _ = fs.NodeAccesser(&Entry{})
// var _ = fs.NodeForgetter(&Entry{})

//var _ = fs.NodeGetattrer(&Entry{})

//var _ = fs.NodeLinker(&Entry{})

//var _ = fs.NodeMkdirer(&Entry{})
//var _ = fs.NodeMknoder(&Entry{})

// var _ = fs.NodeSetxattrer(&Entry{})
// var _ = fs.NodeStringLookuper(&Entry{})
// var _ = fs.NodeSymlinker(&Entry{})
