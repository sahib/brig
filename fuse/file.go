package fuse

import (
	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/disorganizer/brig/util/trie"
	"golang.org/x/net/context"
)

type File struct {
	*trie.Node
}

func (f *File) Attr(ctx context.Context, a *fuse.Attr) error {
	// TODO: Store special permissions? Is this allowed?
	a.Mode = 0755
	return nil
}

func (f *File) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	// TODO: Open the file and return a fs.Handle.
	//       actual data will be read by Read()
	return nil, nil
}

func (f *File) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
	// TODO: Close the file and sync/flush data.
	return nil
}

func (f *File) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	// TODO: Read file at req.Offset for req.Size bytes and set resp.Data.
	return nil
}

func (f *File) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	// TODO: Write req.Data at req.Offset to file.
	//       Expand file if necessary and update Size.
	//       Return the number of written bytes in resp.Size
	return nil
}

func (f *File) Flush(ctx context.Context, req *fuse.FlushRequest) error {
	// TODO: Flush any pending data. Maybe a No-Op?
	return nil
}

func (f *File) Setattr(ctx context.Context, req *fuse.SetattrRequest, resp *fuse.SetattrResponse) error {
	// TODO: Update {m,c,a}time? Maybe not needed/Unsure when this is called.
	return nil
}
