package fuse

import (
	"os"
	"unsafe"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/store"
	"golang.org/x/net/context"
)

type Dir struct {
	File *store.File
	fs   *FS
}

func (d *Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Mode = os.ModeDir | 0755
	a.Size = uint64(d.File.Size)
	a.Mtime = d.File.ModTime
	return nil
}

func (d *Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	child := d.File.Lookup(name)
	if child == nil {
		return nil, fuse.ENOENT
	}

	if !child.IsLeaf() {
		return &Dir{
			File: child,
			fs:   d.fs,
		}, nil
	}

	return &Entry{
		File: child,
		fs:   d.fs,
	}, nil
}

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

func (d *Dir) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	child, err := d.File.Insert(req.Name, req.Mode&os.ModeDir == 0)
	if err != nil {
		log.WithFields(log.Fields{
			"child": child,
			"error": err,
		}).Warning("fuse-create failed")
		return nil, nil, fuse.ENODATA
	}

	entry := &Entry{File: child, fs: d.fs}
	return entry, Handle{Entry: entry}, nil
}

func (d *Dir) Remove(ctx context.Context, req *fuse.RemoveRequest) error {
	child := d.File.Lookup(req.Name)
	if child == nil {
		return fuse.ENOENT
	}

	child.Remove()
	return nil
}

func (d *Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	children := d.File.Children()
	fuseEnts := make([]fuse.Dirent, 0, len(children))

	for _, child := range children {
		childType := fuse.DT_File
		if !child.IsLeaf() {
			childType = fuse.DT_Dir
		}

		fuseEnts = append(fuseEnts, fuse.Dirent{
			Inode: *(*uint64)(unsafe.Pointer(&d.File)),
			Type:  childType,
			Name:  child.Name(),
		})
	}

	return fuseEnts, nil
}
