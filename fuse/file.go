package fuse

import (
	"path"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	log "github.com/Sirupsen/logrus"
	"github.com/sahib/brig/catfs"
	"golang.org/x/net/context"
)

// File is a file inside a directory.
type File struct {
	path string
	cfs  *catfs.FS
}

// Attr is called to get the stat(2) attributes of a file.
func (fi *File) Attr(ctx context.Context, attr *fuse.Attr) error {
	log.Debugf("exec file attr: %v", fi.path)
	info, err := fi.cfs.Stat(fi.path)
	if err != nil {
		return err
	}

	attr.Mode = 0755
	attr.Size = info.Size
	attr.Mtime = info.ModTime
	attr.Inode = info.Inode
	return nil
}

// Open is called to get an opened handle of a file, suitable for reading and writing.
func (fi *File) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	log.Debugf("fuse-open: %s", fi.path)
	fd, err := fi.cfs.Open(fi.path)
	if err != nil {
		return nil, errorize("file-open", err)
	}

	return &Handle{fd: fd, cfs: fi.cfs}, nil
}

// Setattr is called once an attribute of a file changes.
// Most importantly, size changes are reported here, e.g. after truncating a
// file, the size change is noticed here before Open() is called.
func (fi *File) Setattr(ctx context.Context, req *fuse.SetattrRequest, resp *fuse.SetattrResponse) error {
	// This is called when any attribute of the file changes,
	// most importantly the file size. For example it is called when truncating
	// the file to zero bytes with a size change of `0`.
	log.Debugf("exec file setattr")
	switch {
	case req.Valid&fuse.SetattrSize != 0:
		log.Warningf("SIZE CHANGED OF %s: %d", fi.path, req.Size)
		if err := fi.cfs.Truncate(fi.path, req.Size); err != nil {
			return errorize("file-setattr-size", err)
		}
	case req.Valid&fuse.SetattrMtime != 0:
		if err := fi.cfs.Touch(fi.path); err != nil {
			return errorize("file-setattr-mtime", err)
		}
	}

	return nil
}

// Fsync is called when any open buffers need to be written to disk.
// Currently, fsync is completely ignored.
func (fi *File) Fsync(ctx context.Context, req *fuse.FsyncRequest) error {
	log.Debugf("exec file fsync")
	return nil
}

// Getxattr is called to get a single xattr (extended attribute) of a file.
func (fi *File) Getxattr(ctx context.Context, req *fuse.GetxattrRequest, resp *fuse.GetxattrResponse) error {
	log.Debugf("exec file getxattr: %v", fi.path)
	switch req.Name {
	case "brig.hash":
		info, err := fi.cfs.Stat(fi.path)
		if err != nil {
			return errorize("file-getxattr", err)
		}

		// Truncate if less bytes were requested for some reason:
		hash := info.Hash.B58String()
		if uint32(len(hash)) > req.Size {
			hash = hash[:req.Size]
		}

		resp.Xattr = []byte(hash)
	default:
		return fuse.ErrNoXattr
	}

	return nil
}

// Listxattr is called to list all xattrs of this file.
func (fi *File) Listxattr(ctx context.Context, req *fuse.ListxattrRequest, resp *fuse.ListxattrResponse) error {
	log.Debugf("exec file listxattr")
	resp.Append("brig.hash")
	return nil
}

func (fi *File) Rename(ctx context.Context, req *fuse.RenameRequest, newDir fs.Node) error {
	log.Debugf("exec file rename")
	newParent, ok := newDir.(*Directory)
	if !ok {
		return fuse.EIO
	}

	newPath := path.Join(newParent.path, req.NewName)
	if err := fi.cfs.Move(fi.path, newPath); err != nil {
		log.Warningf("fuse: File: mv: %v", err)
		return err
	}

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
