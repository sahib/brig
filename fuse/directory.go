package fuse

import (
	"os"
	"path"

	"context"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/catfs"
)

// Directory represents a directory node.
type Directory struct {
	path string
	cfs  *catfs.FS
}

// Attr is called to retrieve stat-metadata about the directory.
func (dir *Directory) Attr(ctx context.Context, attrs *fuse.Attr) error {
	log.Debugf("Exec dir attr: %v", dir.path)
	info, err := dir.cfs.Stat(dir.path)
	if err != nil {
		return errorize("dir-attr", err)
	}

	attrs.Mode = os.ModeDir | 0755
	attrs.Size = info.Size
	attrs.Mtime = info.ModTime
	attrs.Inode = info.Inode
	return nil
}

// Lookup is called to lookup a direct child of the directory.
func (dir *Directory) Lookup(ctx context.Context, name string) (fs.Node, error) {
	log.Debugf("Exec lookup: %v", name)
	if name == "." {
		return dir, nil
	}

	if name == ".." && dir.path != "/" {
		return &Directory{path: path.Dir(dir.path), cfs: dir.cfs}, nil
	}

	var result fs.Node
	childPath := path.Join(dir.path, name)

	log.Debugf("   doing stat: %v %v", dir.path, childPath)
	info, err := dir.cfs.Stat(childPath)
	if err != nil {
		return nil, errorize("dir-lookup", err)
	}

	if info.IsDir {
		result = &Directory{path: childPath, cfs: dir.cfs}
	} else {
		result = &File{path: childPath, cfs: dir.cfs}
	}

	return result, nil
}

// Mkdir is called to create a new directory node inside the receiver.
func (dir *Directory) Mkdir(ctx context.Context, req *fuse.MkdirRequest) (fs.Node, error) {
	childPath := path.Join(dir.path, req.Name)
	if err := dir.cfs.Mkdir(childPath, false); err != nil {
		log.WithFields(log.Fields{
			"path":  childPath,
			"error": err,
		}).Warning("fuse-mkdir failed")

		return nil, fuse.ENODATA
	}

	return &Directory{path: childPath, cfs: dir.cfs}, nil
}

// Create is called to create an opened file or directory  as child of the receiver.
func (dir *Directory) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	var err error
	log.Debugf("fuse-create: %v", req.Name)

	switch {
	case req.Mode&os.ModeDir != 0:
		err = dir.cfs.Mkdir(req.Name, false)
	default:
		err = dir.cfs.Touch(req.Name)
	}

	childPath := path.Join(dir.path, req.Name)
	if err != nil {
		log.WithFields(log.Fields{
			"path":  childPath,
			"error": err,
		}).Warning("fuse-create failed")
		return nil, nil, fuse.ENODATA
	}

	fd, err := dir.cfs.Open(childPath)
	if err != nil {
		return nil, nil, errorize("fuse-dir-create", err)
	}

	file := &File{
		path: childPath,
		cfs:  dir.cfs,
	}

	return file, &Handle{fd: fd, cfs: dir.cfs}, nil
}

// Remove is called when a direct child in the directory needs to be removed.
func (dir *Directory) Remove(ctx context.Context, req *fuse.RemoveRequest) error {
	path := path.Join(dir.path, req.Name)
	if err := dir.cfs.Remove(path); err != nil {
		log.Errorf("fuse: dir-remove: `%s` failed: %v", path, err)
		return fuse.ENOENT
	}

	return nil
}

// ReadDirAll is called to get a directory listing of the receiver.
func (dir *Directory) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	log.Debugf("Exec read dir all")
	selfInfo, err := dir.cfs.Stat(dir.path)
	if err != nil {
		log.Debugf("Failed to stat: %v", dir.path)
		return nil, errorize("fuse-dir-ls-stat", err)
	}

	parentDir := path.Dir(dir.path)
	parInfo, err := dir.cfs.Stat(parentDir)
	if err != nil {
		log.Debugf("Failed to stat parent: %v", parentDir)
		return nil, errorize("fuse-dir-ls-stat-par", err)
	}

	fuseEnts := []fuse.Dirent{
		fuse.Dirent{
			Inode: selfInfo.Inode,
			Type:  fuse.DT_Dir,
			Name:  ".",
		},
		fuse.Dirent{
			Inode: parInfo.Inode,
			Type:  fuse.DT_Dir,
			Name:  "..",
		},
	}

	entries, err := dir.cfs.List(dir.path, 1)
	if err != nil {
		log.Debugf("Failed to list entries: %v", dir.path)
		return nil, errorize("fuse-dir-readall", err)
	}

	for _, entry := range entries {
		childType := fuse.DT_File
		if entry.IsDir {
			childType = fuse.DT_Dir
		}

		// If we return the same path (or just "/") to fuse
		// it will return a EIO to userland. Weird.
		if entry.Path == "/" || entry.Path == dir.path {
			continue
		}

		fuseEnts = append(fuseEnts, fuse.Dirent{
			Inode: entry.Inode,
			Type:  childType,
			Name:  path.Base(entry.Path),
		})
	}

	return fuseEnts, nil
}
