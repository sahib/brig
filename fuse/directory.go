// +build linux

package fuse

import (
	"os"
	"path"

	"golang.org/x/net/context"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	log "github.com/Sirupsen/logrus"
	"github.com/sahib/brig/catfs"
)

// Directory represents a directory node.
type Directory struct {
	path string
	cfs  *catfs.FS
}

// Attr is called to retrieve stat-metadata about the directory.
func (dir *Directory) Attr(ctx context.Context, attr *fuse.Attr) error {
	defer logPanic("dir: attr")

	debugLog("Exec dir attr: %v", dir.path)
	info, err := dir.cfs.Stat(dir.path)
	if err != nil {
		return errorize("dir-attr", err)
	}

	// Act like the file is owned by the user of the brig process.
	attr.Uid = uint32(os.Getuid())
	attr.Gid = uint32(os.Getgid())

	attr.Mode = os.ModeDir | 0755
	attr.Size = info.Size
	attr.Mtime = info.ModTime
	attr.Inode = info.Inode
	return nil
}

// Lookup is called to lookup a direct child of the directory.
func (dir *Directory) Lookup(ctx context.Context, name string) (fs.Node, error) {
	defer logPanic("dir: lookup")

	debugLog("Exec lookup: %v", name)
	if name == "." {
		return dir, nil
	}

	if name == ".." && dir.path != "/" {
		return &Directory{path: path.Dir(dir.path), cfs: dir.cfs}, nil
	}

	var result fs.Node
	childPath := path.Join(dir.path, name)

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
	defer logPanic("dir: mkdir")

	debugLog("fuse-mkdir: %v", req.Name)

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
	defer logPanic("dir: create")

	var err error
	debugLog("fuse-create: %v", req.Name)

	childPath := path.Join(dir.path, req.Name)
	switch {
	case req.Mode&os.ModeDir != 0:
		err = dir.cfs.Mkdir(childPath, false)
	default:
		err = dir.cfs.Touch(childPath)
	}

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
	defer logPanic("dir: remove")

	path := path.Join(dir.path, req.Name)
	if err := dir.cfs.Remove(path); err != nil {
		log.Errorf("fuse: dir-remove: `%s` failed: %v", path, err)
		return fuse.ENOENT
	}

	return nil
}

// ReadDirAll is called to get a directory listing of the receiver.
func (dir *Directory) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	defer logPanic("dir: readdirall")

	debugLog("Exec read dir all")
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
		{
			Inode: selfInfo.Inode,
			Type:  fuse.DT_Dir,
			Name:  ".",
		},
		{
			Inode: parInfo.Inode,
			Type:  fuse.DT_Dir,
			Name:  "..",
		},
	}

	entries, err := dir.cfs.List(dir.path, 1)
	if err != nil {
		log.Warningf("Failed to list entries: %v", dir.path)
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

// Getxattr is called to get a single xattr (extended attribute) of a file.
func (dir *Directory) Getxattr(ctx context.Context, req *fuse.GetxattrRequest, resp *fuse.GetxattrResponse) error {
	defer logPanic("dir: getxattr")

	debugLog("exec dir getxattr: %v: %v", dir.path, req.Name)
	xattrs, err := getXattr(dir.cfs, req.Name, dir.path, req.Size)
	if err != nil {
		return err
	}

	resp.Xattr = xattrs
	return nil
}

// Listxattr is called to list all xattrs of this file.
func (dir *Directory) Listxattr(ctx context.Context, req *fuse.ListxattrRequest, resp *fuse.ListxattrResponse) error {
	defer logPanic("dir: listxattr")

	debugLog("exec dir listxattr")
	resp.Xattr = listXattr(req.Size)
	return nil
}
