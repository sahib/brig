package fuse

import (
	"io"
	"os"
	"sync"
	"time"

	"bazil.org/fuse"
	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/store"
	"golang.org/x/net/context"
)

type Handle struct {
	*Entry

	// Protect access of `layer`
	sync.Mutex

	// Write in-memory layer
	layer *store.Layer
}

func (h *Handle) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
	return h.flush()
}

// TODO: Honour ctx.Done() and return fuse.EINTR in that case...

func (h *Handle) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	path := h.File.Path()
	stream, err := h.File.Stream()
	if err != nil {
		log.Errorf("fuse: Read: `%s` failed: %v", path, err)
		return fuse.ENODATA
	}

	log.WithFields(log.Fields{
		"path":   path,
		"offset": req.Offset,
		"size":   req.Size,
	}).Debugf("fuse read")

	pos, err := stream.Seek(req.Offset, os.SEEK_SET)
	if err != nil {
		log.Errorf("fuse: Read: seek failed on `%s`: %v", path, err)
		return fuse.ENODATA
	}

	if pos != req.Offset {
		log.Warningf("fuse: Read: warning: seek_off (%d) != req_off (%d)", pos, req.Offset)
	}

	resp.Data = make([]byte, req.Size)
	n, err := io.ReadAtLeast(stream, resp.Data, req.Size)
	if err != nil && err != io.ErrUnexpectedEOF {
		log.Errorf("fuse: Read: streaming `%s` failed: %v", path, err)
		return fuse.ENODATA
	}

	resp.Data = resp.Data[:n]
	return nil
}

func (h *Handle) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	h.Lock()
	defer h.Unlock()

	log.Infof("Oh, a write request!")

	if h.layer == nil {
		stream, err := h.File.Stream()
		if err != nil {
			return fuse.ENODATA
		}

		h.layer = store.NewLayer(stream)
		log.Debugf("fuse: truncating %s to %d %p", h.File.Path(), h.File.Size, h.File)
		h.File.Lock()
		{
			h.layer.Truncate(int64(h.File.Size))
		}
		h.File.Unlock()
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
	h.File.Lock()
	{
		minSize := store.FileSize(h.layer.MinSize())
		if h.File.Size < minSize {
			log.Debugf("fuse: extending file from %d to %d bytes", h.File.Size, minSize)
			h.File.Size = minSize
		}
	}
	h.File.Unlock()
	return nil
}

func (h *Handle) Flush(ctx context.Context, req *fuse.FlushRequest) error {
	return h.flush()
}

func (h *Handle) flush() error {
	h.Lock()
	defer h.Unlock()

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

	// Update the mtime:
	h.File.Metadata.ModTime = time.Now()

	// TODO: Remove debug tee.
	tee := io.TeeReader(h.layer, os.Stdout)

	if err := h.fs.Store.AddFromReader(h.File.Path(), tee, h.File.Metadata); err != nil {
		log.Warningf("Add failed: %v", err)
		return fuse.ENODATA
	}

	if err := h.layer.Close(); err != nil {
		log.Warningf("Close failed: %v", err)
		return fuse.ENODATA
	}

	return nil
}
