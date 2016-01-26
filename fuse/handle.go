package fuse

import (
	"bytes"
	"os"
	"sync"

	"bazil.org/fuse"
	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/store"
	"golang.org/x/net/context"
)

type Handle struct {
	*File
	sync.Mutex
	layer *store.Layer
}

func (h *Handle) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
	return h.flush()
}

// TODO: Implement Read, but that needs seekable compression first.
// func (f *File) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
// 	// TODO: Read file at req.Offset for req.Size bytes and set resp.Data.
// 	resp.Data = make([]byte, req.Size)
// 	for i := 0; i < req.Size; i++ {
// 		resp.Data[i] = byte("Na"[i%2])
// 	}
//
// 	return nil
// }

func (h *Handle) ReadAll(ctx context.Context) ([]byte, error) {
	buf := &bytes.Buffer{}

	path := h.Path()
	if err := h.fs.Store.Cat(path, buf); err != nil {
		log.Errorf("fuse: ReadAll: `%s` failed: %v", path, err)
		return nil, fuse.ENODATA
	}

	return buf.Bytes(), nil
}

// TODO: Implement. Needs scratchpad implementation.
func (h *Handle) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	// TODO: Write req.Data at req.Offset to file.
	//       Expand file if necessary and update Size.
	//       Return the number of written bytes in resp.Size
	h.Lock()
	defer h.Unlock()

	log.Infof("Oh, a write request!")

	if h.layer == nil {
		stream, err := h.fs.Store.Stream(h.Path())
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

	if err := h.fs.Store.AddFromReader(h.Path(), h.layer); err != nil {
		log.Warningf("Add failed: %v", err)
	}

	if err := h.layer.Close(); err != nil {
		log.Warningf("Close failed: %v", err)
	}

	return nil
}
