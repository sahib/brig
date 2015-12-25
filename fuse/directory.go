package fuse

import (
	"os"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/disorganizer/brig/util/trie"
	"golang.org/x/net/context"
)

type Dir struct {
	*trie.Node

	fs *FS
}

func (d *Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Mode = os.ModeDir | 0755
	return nil
}

func (d *Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	// TODO: Actually lookup `name` (no path) and create File or Dir.
	return nil, nil
}

func (d *Dir) Mkdir(ctx context.Context, req *fuse.MkdirRequest) (fs.Node, error) {
	// TODO: Actually create Dir and return it.
	return nil, nil
}

func (d *Dir) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	// TODO: Create File/Dir and return Node + open Handle.
	return nil, nil, nil
}

func (d *Dir) Remove(ctx context.Context, req *fuse.RemoveRequest) error {
	// TODO: Remove File/Dir.
	return nil
}
