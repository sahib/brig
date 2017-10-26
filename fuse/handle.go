package fuse

import (
	"os"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/catfs"
	"golang.org/x/net/context"
)

// Handle is an open Entry.
type Handle struct {
	fd  *catfs.Handle
	cfs *catfs.FS
}

// Release is called to close this handle.
func (hd *Handle) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
	return hd.flush()
}

// TODO: Honour ctx.Done() and return fuse.EINTR in that case...

// Read is called to read a block of data at a certain offset.
func (hd *Handle) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	log.WithFields(log.Fields{
		"path":   hd.fd.Path(),
		"offset": req.Offset,
		"size":   req.Size,
	}).Debugf("fuse: handle: read")

	newOff, err := hd.fd.Seek(req.Offset, os.SEEK_SET)
	if err != nil {
		return errorize("handle-read-seek", err)
	}

	if newOff != req.Offset {
		log.Warningf("read seek offset differs (want %d, got %d)", req.Offset, newOff)
	}

	n, err := hd.fd.Read(resp.Data[:req.Size])
	if err != nil {
		return errorize("handle-read-io", err)
	}

	resp.Data = resp.Data[:n]
	return nil
}

// Write is called to write a block of data at a certain offset.
func (hd *Handle) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	log.Debugf(
		"fuse-write: %s (off: %d size: %d)",
		hd.fd.Path(),
		req.Offset,
		len(req.Data),
	)

	newOff, err := hd.fd.Seek(req.Offset, os.SEEK_SET)
	if err != nil {
		return errorize("handle-write-seek", err)
	}

	if newOff != req.Offset {
		log.Warningf("write seek offset differs (want %d, got %d)", req.Offset, newOff)
	}

	n, err := hd.fd.Write(req.Data)
	if err != nil {
		return errorize("handle-write-io", err)
	}

	// Report back to fuse how many bytes we wrote.
	resp.Size = n
	return nil
}

// Flush is called to make sure all written contents get synced to disk.
func (hd *Handle) Flush(ctx context.Context, req *fuse.FlushRequest) error {
	return hd.flush()
}

// flush does the actual adding to brig.
func (hd *Handle) flush() error {
	log.Debugf("fuse-flush: %v", hd.fd.Path())
	return errorize("handle-flush", hd.fd.Flush())
}

// Compiler checks to see if we got all the interfaces right:
var _ = fs.HandleFlusher(&Handle{})
var _ = fs.HandleReader(&Handle{})
var _ = fs.HandleReleaser(&Handle{})
var _ = fs.HandleWriter(&Handle{})
