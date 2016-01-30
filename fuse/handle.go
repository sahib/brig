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

func (h *Handle) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	path := h.File.Path()
	stream, err := h.File.Stream()
	if err != nil {
		log.Errorf("fuse: Read: `%s` failed: %v", path, err)
		return fuse.ENODATA
	}

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
