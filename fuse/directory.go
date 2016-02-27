package fuse

import (
	"os"
	"path/filepath"
	"unsafe"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/store"
	"golang.org/x/net/context"
)

// Dir represents a directory node.
type Dir struct {
	*store.File
	fs *FS
}

// Attr is called to retrieve stat-metadata about the directory.
func (d *Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Mode = os.ModeDir | 0755
	a.Size = uint64(d.Size())
	a.Mtime = d.ModTime()
	return nil
}

// Lookup is called to lookup a direct child of the directory.
func (d *Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	child := d.File.Lookup(name)
	if child == nil {
		return nil, fuse.ENOENT
	}

	if name == "." {
		return d, nil
	}

	if name == ".." {
		return &Dir{File: d.Parent(), fs: d.fs}, nil
	}

	switch knd := child.Kind(); knd {
	case store.FileTypeRegular:
		return &Entry{File: child, fs: d.fs}, nil
	case store.FileTypeDir:
		return &Dir{File: child, fs: d.fs}, nil
	default:
		log.Errorf("Bad/unsupported file type: %d", knd)
		return nil, fuse.EIO
	}
}

// Mkdir is called to create a new directory node inside the receiver.
func (d *Dir) Mkdir(ctx context.Context, req *fuse.MkdirRequest) (fs.Node, error) {
	child, err := d.fs.Store.Mkdir(req.Name)
	if err != nil {
		log.WithFields(log.Fields{
			"child": child,
			"error": err,
		}).Warning("fuse-mkdir failed")

		return nil, fuse.ENODATA
	}

	return &Dir{File: child, fs: d.fs}, nil
}

// Create is called to create an opened file or directory  as child of the receiver.
func (d *Dir) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	var err error

	log.Debugf("fuse-create: %v", req.Name)

	switch {
	case req.Mode&os.ModeDir != 0:
		_, err = d.fs.Store.Mkdir(req.Name)
	default:
		err = d.fs.Store.Touch(req.Name)
	}

	if err != nil {
		log.WithFields(log.Fields{
			"path":  d.Path(),
			"child": req.Name,
			"error": err,
		}).Warning("fuse-create failed")
		return nil, nil, fuse.ENODATA
	}

	child := d.Child(req.Name)
	if child == nil {
		log.Warning("No child %v in %v", req.Name, d)
		return nil, nil, fuse.ENODATA
	}

	entry := &Entry{File: d.Child(req.Name), fs: d.fs}
	return entry, &Handle{Entry: entry}, nil
}

// Remove is called when a direct child in the directory needs to be removed.
func (d *Dir) Remove(ctx context.Context, req *fuse.RemoveRequest) error {
	path := filepath.Join(d.Path(), req.Name)
	if err := d.fs.Store.Rm(path); err != nil {
		log.Errorf("fuse-rm `%s` failed: %v", path, err)
		return fuse.ENOENT
	}

	return nil
}

// ReadDirAll is called to get a directory listing of the receiver.
func (d *Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	children := d.Children()
	fuseEnts := make([]fuse.Dirent, 2, len(children)+2)

	fuseEnts[0] = fuse.Dirent{
		Inode: *(*uint64)(unsafe.Pointer(&d.File)),
		Type:  fuse.DT_Dir,
		Name:  ".",
	}

	fuseEnts[1] = fuse.Dirent{
		Inode: *(*uint64)(unsafe.Pointer(&d.File)) + 1,
		Type:  fuse.DT_Dir,
		Name:  "..",
	}

	for _, child := range children {
		childType := fuse.DT_File
		switch kind := child.Kind(); kind {
		case store.FileTypeDir:
			childType = fuse.DT_Dir
		case store.FileTypeRegular:
			childType = fuse.DT_File
		default:
			log.Errorf("Warning: Bad/unsupported file type: %v", kind)
			return nil, fuse.EIO
		}

		fuseEnts = append(fuseEnts, fuse.Dirent{
			Inode: *(*uint64)(unsafe.Pointer(&child)),
			Type:  childType,
			Name:  child.Name(),
		})
	}

	return fuseEnts, nil
}
