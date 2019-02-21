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

// File is a file inside a directory.
type File struct {
	path string
	m    *Mount
}

// Attr is called to get the stat(2) attributes of a file.
func (fi *File) Attr(ctx context.Context, attr *fuse.Attr) error {
	defer logPanic("file: attr")

	info, err := fi.m.fs.Stat(fi.path)
	if err != nil {
		return err
	}
	debugLog("exec file attr: %v", fi.path)

	attr.Mode = 0755
	attr.Size = info.Size
	attr.Mtime = info.ModTime
	attr.Inode = info.Inode

	// Act like the file is owned by the user of the brig process.
	attr.Uid = uint32(os.Getuid())
	attr.Gid = uint32(os.Getgid())

	// tools like `du` rely on this for size calculation
	// (assuming every fs block takes actual storage, but we only emulate this
	// here for compatibility; see man 2 stat for the why for "512")
	attr.BlockSize = 4096
	attr.Blocks = info.Size / 512
	if info.Size%uint64(512) > 0 {
		attr.Blocks++
	}

	return nil
}

// Open is called to get an opened handle of a file, suitable for reading and writing.
func (fi *File) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	defer logPanic("file: open")

	debugLog("fuse-open: %s", fi.path)
	fd, err := fi.m.fs.Open(fi.path)
	if err != nil {
		return nil, errorize("file-open", err)
	}

	return &Handle{fd: fd, m: fi.m}, nil
}

// Setattr is called once an attribute of a file changes.
// Most importantly, size changes are reported here, e.g. after truncating a
// file, the size change is noticed here before Open() is called.
func (fi *File) Setattr(ctx context.Context, req *fuse.SetattrRequest, resp *fuse.SetattrResponse) error {
	defer logPanic("file: setattr")

	// This is called when any attribute of the file changes,
	// most importantly the file size. For example it is called when truncating
	// the file to zero bytes with a size change of `0`.
	debugLog("exec file setattr")
	switch {
	case req.Valid&fuse.SetattrSize != 0:
		if err := fi.m.fs.Truncate(fi.path, req.Size); err != nil {
			return errorize("file-setattr-size", err)
		}
	case req.Valid&fuse.SetattrMtime != 0:
		if err := fi.m.fs.Touch(fi.path); err != nil {
			return errorize("file-setattr-mtime", err)
		}
	}

	return nil
}

// Fsync is called when any open buffers need to be written to disk.
// Currently, fsync is completely ignored.
func (fi *File) Fsync(ctx context.Context, req *fuse.FsyncRequest) error {
	defer logPanic("file: fsync")

	debugLog("exec file fsync")
	return nil
}

// Getxattr is called to get a single xattr (extended attribute) of a file.
func (fi *File) Getxattr(ctx context.Context, req *fuse.GetxattrRequest, resp *fuse.GetxattrResponse) error {
	defer logPanic("file: getxattr")

	debugLog("exec file getxattr: %v: %v", fi.path, req.Name)
	xattrs, err := getXattr(fi.m.fs, req.Name, fi.path, req.Size)
	if err != nil {
		return err
	}

	resp.Xattr = xattrs
	return nil
}

// Listxattr is called to list all xattrs of this file.
func (fi *File) Listxattr(ctx context.Context, req *fuse.ListxattrRequest, resp *fuse.ListxattrResponse) error {
	defer logPanic("file: listxattr")

	debugLog("exec file listxattr")
	resp.Xattr = listXattr(req.Size)
	return nil
}

// Rename is called when the node changed its path.
func (fi *File) Rename(ctx context.Context, req *fuse.RenameRequest, newDir fs.Node) error {
	defer logPanic("file: rename")

	debugLog("exec file rename")
	newParent, ok := newDir.(*Directory)
	if !ok {
		return fuse.EIO
	}

	newPath := path.Join(newParent.path, req.NewName)
	if err := fi.m.fs.Move(fi.path, newPath); err != nil {
		log.Warningf("fuse: file: mv: %v", err)
		return err
	}

	notifyChange(fi.m, 100*time.Millisecond)
	return nil
}

// Compile time checks to see which interfaces we implement:
// Please update this list when modifying code here.
var _ = fs.Node(&File{})
var _ = fs.NodeFsyncer(&File{})
var _ = fs.NodeGetxattrer(&File{})
var _ = fs.NodeListxattrer(&File{})
var _ = fs.NodeOpener(&File{})
var _ = fs.NodeSetattrer(&File{})

// Other interfaces are available, but currently not needed or make sense:
// var _ = fs.NodeRenamer(&File{})
// var _ = fs.NodeReadlinker(&File{})
// var _ = fs.NodeRemover(&File{})
// var _ = fs.NodeRemovexattrer(&File{})
// var _ = fs.NodeRequestLookuper(&File{})
// var _ = fs.NodeAccesser(&File{})
// var _ = fs.NodeForgetter(&File{})
// var _ = fs.NodeGetattrer(&File{})
// var _ = fs.NodeLinker(&File{})
// var _ = fs.NodeMkdirer(&File{})
// var _ = fs.NodeMknoder(&File{})
// var _ = fs.NodeSetxattrer(&File{})
// var _ = fs.NodeStringLookuper(&File{})
// var _ = fs.NodeSymlinker(&File{})
