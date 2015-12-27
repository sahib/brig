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
	//       - Lookup d.Children[name]
	//		 - Check if it's a leaf:
	//		   - If yes, create a File and return it.
	//		   - If no, create a Dir and return it.
	return nil, nil
}

func (d *Dir) Mkdir(ctx context.Context, req *fuse.MkdirRequest) (fs.Node, error) {
	// TODO: Actually create Dir and return it.
	//		 - Create dir c.
	//       - Insert it to to d.Root() at join(d.Path(), req.Name)
	//       - Return dir c.
	return nil, nil
}

func (d *Dir) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	// TODO: Create File/Dir and return Node + open Handle.
	//       - Honour req.Name, req.Mode, req.Umask
	return nil, nil, nil
}

func (d *Dir) Remove(ctx context.Context, req *fuse.RemoveRequest) error {
	// TODO: Remove File/Dir.
	return nil
}
