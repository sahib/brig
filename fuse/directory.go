package fuse

import (
	"os"
	"unsafe"

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
	a.String()
	return nil
}

func (d *Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	// TODO: Actually lookup `name` (no path) and create File or Dir.
	//       - Lookup d.Children[name]
	//		 - Check if it's a leaf:
	//		   - If yes, create a File and return it.
	//		   - If no, create a Dir and return it.
	d.Node.RLock()
	defer d.Node.RUnlock()

	child, ok := d.Node.Children[name]
	if !ok {
		return nil, fuse.ENOENT
	}

	if !child.IsLeaf() {
		return &Dir{Node: child, fs: d.fs}, nil
	}

	return &File{Node: child}, nil
}

func (d *Dir) Mkdir(ctx context.Context, req *fuse.MkdirRequest) (fs.Node, error) {
	// TODO: Actually create Dir and return it.
	//		 - Create dir c.
	//       - Insert it to to d.Root() at join(d.Path(), req.Name)
	//       - Return dir c.
	d.Node.Lock()
	defer d.Node.Unlock()

	child := d.Insert(req.Name)
	return &Dir{Node: child, fs: d.fs}, nil
}

func (d *Dir) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	// TODO: Create File/Dir and return Node + open Handle.
	//       - Honour req.Name, req.Mode, req.Umask
	d.Node.Lock()
	defer d.Node.Unlock()

	// TODO: Differentiate between dir/file
	child := d.Insert(req.Name)
	file := &File{Node: child}
	return file, file, nil
}

func (d *Dir) Remove(ctx context.Context, req *fuse.RemoveRequest) error {
	// TODO: Remove File/Dir.
	d.Node.Lock()
	defer d.Node.Unlock()

	return nil
}

// TODO: LOCKING
func (d *Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	d.Node.RLock()
	defer d.Node.RUnlock()

	children := make([]fuse.Dirent, 0, len(d.Children))

	for name, child := range d.Children {
		childType := fuse.DT_File
		if !child.IsLeaf() {
			childType = fuse.DT_Dir
		}

		children = append(children, fuse.Dirent{
			Inode: *(*uint64)(unsafe.Pointer(&d.Node)),
			Type:  childType,
			Name:  name,
		})
	}

	return children, nil
}
