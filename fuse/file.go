package fuse

import (
	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/disorganizer/brig/util/trie"
	"golang.org/x/net/context"

	// Don't panic.
	// This is just to convert a pointer to an inode.
	"unsafe"
)

type File struct {
	*trie.Node
	fs *FS
}

func (f *File) Attr(ctx context.Context, a *fuse.Attr) error {
	// TODO: Store special permissions? Is this allowed?
	a.Mode = 0755
	a.Size = 200
	a.Inode = *(*uint64)(unsafe.Pointer(&f.Node))
	return nil
}

func (f *File) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	// TODO: Open the file and return a fs.Handle.
	//       actual data will be read by Read()
	return &Handle{File: f}, nil
}

func (f *File) Setattr(ctx context.Context, req *fuse.SetattrRequest, resp *fuse.SetattrResponse) error {
	// TODO: Update {m,c,a}time? Maybe not needed/Unsure when this is called.
	return nil
}
