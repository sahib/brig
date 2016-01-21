package fuse

import (
	"bytes"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	log "github.com/Sirupsen/logrus"
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
	return f, nil
}

func (f *File) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
	// TODO: Close the file and sync/flush data.
	return nil
}

// TODO: Implement Read, but that needs seekable compression first.
// func (f *File) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
// 	// TODO: Read file at req.Offset for req.Size bytes and set resp.Data.
// 	resp.Data = make([]byte, req.Size)
// 	for i := 0; i < req.Size; i++ {
// 		resp.Data[i] = byte("Na"[i%2])
// 	}
//
// 	return nil
// }

func (f *File) ReadAll(ctx context.Context) ([]byte, error) {
	buf := &bytes.Buffer{}

	path := f.Path()
	if err := f.fs.Store.Cat(path, buf); err != nil {
		log.Errorf("fuse: ReadAll: `%s` failed: %v", path, err)
		return nil, fuse.ENODATA
	}

	return buf.Bytes(), nil
}

// TODO: Implement. Needs scratchpad implementation.
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
