package fuse

import (
	"io"
	"os"
	"sync"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/store"
	"github.com/disorganizer/brig/util/ipfsutil"
	"golang.org/x/net/context"
)

// Handle is an open Entry.
type Handle struct {
	*Entry

	// Protect access of `layer`
	laymu sync.Mutex

	// actual data stream
	stream ipfsutil.Reader

	// Write in-memory layer
	layer *store.Layer
}

// Release is called to close this handle.
func (h *Handle) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
	return h.flush()
}

// TODO: Honour ctx.Done() and return fuse.EINTR in that case...

// Read is called to read a block of data at a certain offset.
func (h *Handle) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	path := h.Path()

	h.laymu.Lock()
	defer h.laymu.Unlock()

	log.WithFields(log.Fields{
		"path":   path,
		"offset": req.Offset,
		"size":   req.Size,
	}).Debugf("fuse read")

	if h.stream == nil {
		stream, err := h.Stream()
		if err != nil {
			log.Errorf("fuse-read: Cannot open stream: %v", err)
			return fuse.ENODATA
		}

		h.stream = stream
	}

	pos, err := h.stream.Seek(req.Offset, os.SEEK_SET)
	if err != nil && err != io.EOF {
		log.Errorf("fuse-read: seek failed on `%s`: %v", path, err)
		return fuse.ENODATA
	}

	if pos != req.Offset {
		log.Warningf("fuse-read: warning: seek_off (%d) != req_off (%d)", pos, req.Offset)
	}

	n, err := io.ReadAtLeast(h.stream, resp.Data[:req.Size], req.Size)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		log.Errorf("fuse-read: streaming `%s` failed: %v", path, err)
		return fuse.ENODATA
	}

	resp.Data = resp.Data[:n]
	return nil
}

// Write is called to write a block of data at a certain offset.
func (h *Handle) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	h.laymu.Lock()
	defer h.laymu.Unlock()

	log.Debugf("fuse-write: %s (off: %d size: %d)", h.Path(), req.Offset, len(req.Data))

	if h.layer == nil {
		if h.stream == nil {
			stream, err := h.Stream()
			if err != nil {
				return fuse.ENODATA
			}
			h.stream = stream
		}

		log.Debugf("fuse-write: truncating %s to %d %p", h.Path(), h.Size(), h)
		h.layer = store.NewLayer(h.stream)
		h.layer.Truncate(h.Size())
	}

	_, err := h.layer.Seek(req.Offset, os.SEEK_SET)
	if err != nil {
		log.Warningf("Seek failure: %v", err)
		return fuse.ENODATA
	}

	n, err := h.layer.Write(req.Data)
	if err != nil {
		log.Warningf("Write failure: %v", err)
		return fuse.ENODATA
	}

	resp.Size = n

	// Update the file size, if it changed; fuse doc demands this:
	// https://godoc.org/bazil.org/fuse/fs#HandleWriter
	minSize := h.layer.MinSize()
	if h.Size() < minSize {
		log.Debugf("fuse-write: extending file from %d to %d bytes", h.Size(), minSize)
		h.UpdateSize(minSize)
	}

	return nil
}

// Flush is called to make sure all written contents get synced to disk.
func (h *Handle) Flush(ctx context.Context, req *fuse.FlushRequest) error {
	return h.flush()
}

// flush does the actual adding to brig.
func (h *Handle) flush() error {
	h.laymu.Lock()
	defer h.laymu.Unlock()

	log.Debugf("fuse-flush: %v (%p)", h.Path(), h.layer)

	if h.layer == nil {
		return nil
	}

	defer func() {
		h.layer = nil
	}()

	n, err := h.layer.Seek(0, os.SEEK_SET)
	if err != nil {
		log.Warningf("Seek failed on flush: %v", err)
		return fuse.ENODATA
	}

	if n != 0 {
		log.Warningf("Seek offset is not 0")
	}

	err = h.fs.Store.AddFromReader(h.Path(), h.layer, h.Size())
	if err != nil && err != store.ErrNoChange {
		log.Warningf("Add failed: %v", err)
		return fuse.ENODATA
	}

	if err := h.layer.Close(); err != nil {
		log.Warningf("Close failed: %v", err)
		return fuse.ENODATA
	}

	return nil
}

// Compiler checks to see if we got all the interfaces right:
var _ = fs.HandleFlusher(&Handle{})
var _ = fs.HandleReader(&Handle{})
var _ = fs.HandleReleaser(&Handle{})
var _ = fs.HandleWriter(&Handle{})
