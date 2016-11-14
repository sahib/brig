package fuse

import (
	"fmt"
	"os"
	"path"
	"unsafe"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/store"
	"golang.org/x/net/context"
)

// Dir represents a directory node.
type Dir struct {
	path string
	fsys *Filesystem
}

func Errorize(name string, err error) error {
	if store.IsNoSuchFileError(err) {
		return fuse.ENOENT
	}

	if err != nil {
		log.Warningf("fuse: %s: %v", name, err)
		return fuse.EIO
	}

	return nil
}

// Attr is called to retrieve stat-metadata about the directory.
func (d *Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	return Errorize("dir-attr", d.fsys.Store.ViewNode(d.path, func(nd store.Node) error {
		a.Mode = os.ModeDir | 0755
		a.Size = nd.Size()
		a.Mtime = nd.ModTime()
		return nil
	}))
}

// Lookup is called to lookup a direct child of the directory.
func (d *Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	if name == "." {
		return d, nil
	}

	if name == ".." && d.path != "/" {
		return &Dir{path: path.Dir(d.path), fsys: d.fsys}, nil
	}

	var result fs.Node
	childPath := path.Join(d.path, name)

	err := Errorize("dir-lookup", d.fsys.Store.ViewNode(childPath, func(nd store.Node) error {
		switch knd := nd.GetType(); knd {
		case store.NodeTypeFile:
			fmt.Println("is a entry", d.path)
			result = &Entry{path: childPath, fsys: d.fsys}
		case store.NodeTypeDirectory:
			fmt.Println("is a dir", d.path)
			result = &Dir{path: childPath, fsys: d.fsys}
		case store.NodeTypeCommit:
			// NOTE: Might be useful in the future.
			fallthrough
		default:
			log.Errorf("Bad/unsupported file type: %d", knd)
			return fuse.ENOENT
		}

		return nil
	}))

	if err != nil {
		return nil, err
	}

	return result, nil
}

// Mkdir is called to create a new directory node inside the receiver.
func (d *Dir) Mkdir(ctx context.Context, req *fuse.MkdirRequest) (fs.Node, error) {
	childPath := path.Join(d.path, req.Name)
	if err := d.fsys.Store.Mkdir(childPath); err != nil {
		log.WithFields(log.Fields{
			"path":  childPath,
			"error": err,
		}).Warning("fuse-mkdir failed")

		return nil, fuse.ENODATA
	}

	return &Dir{path: childPath, fsys: d.fsys}, nil
}

// Create is called to create an opened file or directory  as child of the receiver.
func (d *Dir) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	var err error
	log.Debugf("fuse-create: %v", req.Name)

	switch {
	case req.Mode&os.ModeDir != 0:
		err = d.fsys.Store.Mkdir(req.Name)
	default:
		err = d.fsys.Store.Touch(req.Name)
	}

	if err != nil {
		log.WithFields(log.Fields{
			"path":  d.path,
			"child": req.Name,
			"error": err,
		}).Warning("fuse-create failed")
		return nil, nil, fuse.ENODATA
	}

	entry := &Entry{
		path: path.Join(d.path, req.Name),
		fsys: d.fsys,
	}

	return entry, &Handle{Entry: entry}, nil
}

// Remove is called when a direct child in the directory needs to be removed.
func (d *Dir) Remove(ctx context.Context, req *fuse.RemoveRequest) error {
	path := path.Join(d.path, req.Name)
	if err := d.fsys.Store.Remove(path, false); err != nil {
		log.Errorf("fuse: dir-remove: `%s` failed: %v", path, err)
		return fuse.ENOENT
	}

	return nil
}

// ReadDirAll is called to get a directory listing of the receiver.
func (d *Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	fuseEnts := make([]fuse.Dirent, 2)

	// TODO: use node.ID() as inode.
	fuseEnts[0] = fuse.Dirent{
		Inode: *(*uint64)(unsafe.Pointer(&d)),
		Type:  fuse.DT_Dir,
		Name:  ".",
	}

	fuseEnts[1] = fuse.Dirent{
		Inode: *(*uint64)(unsafe.Pointer(&d)) + 1,
		Type:  fuse.DT_Dir,
		Name:  "..",
	}

	err := d.fsys.Store.ViewDir(d.path, func(par *store.Directory) error {
		return par.VisitChildren(func(child store.Node) error {
			childType := fuse.DT_File
			switch kind := child.GetType(); kind {
			case store.NodeTypeDirectory:
				childType = fuse.DT_Dir
			case store.NodeTypeFile:
				childType = fuse.DT_File
			default:
				log.Errorf("Warning: Bad/unsupported file type: %v", kind)
				return fuse.EIO
			}

			fuseEnts = append(fuseEnts, fuse.Dirent{
				Inode: *(*uint64)(unsafe.Pointer(&child)),
				Type:  childType,
				Name:  child.Name(),
			})
			return nil
		})
	})

	if err != nil {
		log.Warningf("fuse: dir: readall: %v", err)
		return nil, fuse.EIO
	}

	return fuseEnts, nil
}
