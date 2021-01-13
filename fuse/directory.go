// +build !windows

package fuse

import (
	"os"
	"path"
	"time"

	"context"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	log "github.com/sirupsen/logrus"
)

// Directory represents a directory node.
type Directory struct {
	path string
	m    *Mount
}

// Attr is called to retrieve stat-metadata about the directory.
func (dir *Directory) Attr(ctx context.Context, attr *fuse.Attr) error {
	defer logPanic("dir: attr")

	debugLog("Exec dir attr: %v", dir.path)
	info, err := dir.m.fs.Stat(dir.path)
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
		return &Directory{path: path.Dir(dir.path), m: dir.m}, nil
	}

	var result fs.Node
	childPath := path.Join(dir.path, name)

	info, err := dir.m.fs.Stat(childPath)
	if err != nil {
		return nil, errorize("dir-lookup", err)
	}

	if info.IsDir {
		result = &Directory{path: childPath, m: dir.m}
	} else {
		result = &File{path: childPath, m: dir.m}
	}

	return result, nil
}

// Mkdir is called to create a new directory node inside the receiver.
func (dir *Directory) Mkdir(ctx context.Context, req *fuse.MkdirRequest) (fs.Node, error) {
	defer logPanic("dir: mkdir")

	debugLog("fuse-mkdir: %v", req.Name)

	childPath := path.Join(dir.path, req.Name)
	if err := dir.m.fs.Mkdir(childPath, false); err != nil {
		log.WithFields(log.Fields{
			"path":  childPath,
			"error": err,
		}).Warning("fuse-mkdir failed")

		return nil, fuse.EIO
	}

	notifyChange(dir.m, 100*time.Millisecond)
	return &Directory{path: childPath, m: dir.m}, nil
}

// Create is called to create an opened file or directory  as child of the receiver.
func (dir *Directory) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	defer logPanic("dir: create")

	var err error
	debugLog("fuse-create: %v", req.Name)

	childPath := path.Join(dir.path, req.Name)
	switch {
	case req.Mode&os.ModeDir != 0:
		err = dir.m.fs.Mkdir(childPath, false)
	default:
		err = dir.m.fs.Touch(childPath)
	}

	if err != nil {
		log.WithFields(log.Fields{
			"path":  childPath,
			"error": err,
		}).Warning("fuse-create failed")
		return nil, nil, fuse.EIO
	}

	fd, err := dir.m.fs.Open(childPath)
	if err != nil {
		return nil, nil, errorize("fuse-dir-create", err)
	}

	notifyChange(dir.m, 100*time.Millisecond)
	file := &File{path: childPath, m: dir.m}
	return file, &Handle{fd: fd, m: dir.m}, nil
}

// Remove is called when a direct child in the directory needs to be removed.
func (dir *Directory) Remove(ctx context.Context, req *fuse.RemoveRequest) error {
	defer logPanic("dir: remove")

	path := path.Join(dir.path, req.Name)
	if err := dir.m.fs.Remove(path); err != nil {
		log.Errorf("fuse: dir-remove: `%s` failed: %v", path, err)
		return fuse.ENOENT
	}

	notifyChange(dir.m, 100*time.Millisecond)
	return nil
}

// ReadDirAll is called to get a directory listing of the receiver.
func (dir *Directory) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	defer logPanic("dir: readdirall")

	debugLog("Exec read dir all")
	selfInfo, err := dir.m.fs.Stat(dir.path)
	if err != nil {
		log.Debugf("Failed to stat: %v", dir.path)
		return nil, errorize("fuse-dir-ls-stat", err)
	}

	parentDir := path.Dir(dir.path)
	parInfo, err := dir.m.fs.Stat(parentDir)
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

	entries, err := dir.m.fs.List(dir.path, 1)
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
	xattrs, err := getXattr(dir.m.fs, req.Name, dir.path, req.Size)
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

// Rename or move files or directories
// TODO: fix info availability,
//       somehow the info about moved item is not visible for a little while after move
//       It usually available after a second or two.
//       How to reproduce
//       mv file1 file2
//       ls -l file2
//       You will see that username, permission, size, date, and so on all in question marks
//       For what I can see. ls cannot access this particular file, even though
//       It will appear as an entry in the call to ReadDirAll done by `ls on_dir`
//       Seems to be cache related issue
func (dir *Directory) Rename(ctx context.Context, req *fuse.RenameRequest, newDir fs.Node) error {
	defer logPanic("dir: rename")

	debugLog("exec dir rename")
	newParent, ok := newDir.(*Directory)
	if !ok {
		return fuse.EIO
	}
	oldPath := path.Join(dir.path, req.OldName)
	newPath := path.Join(newParent.path, req.NewName)
	if err := dir.m.fs.Move(oldPath, newPath); err != nil {
		log.Warningf("fuse: dir: mv: %v", err)
		return err
	}

	notifyChange(dir.m, 100*time.Millisecond)
	return nil
}
