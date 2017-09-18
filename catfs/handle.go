package catfs

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/disorganizer/brig/catfs/mio"
	"github.com/disorganizer/brig/catfs/mio/overlay"
	n "github.com/disorganizer/brig/catfs/nodes"
)

// TODO: Implement, using the implemenation found in fuse.
type Handle struct {
	fs     *FS
	file   *n.File
	lock   sync.Mutex
	layer  *overlay.Layer
	stream mio.Stream
}

func newHandle(file *n.File) *Handle {
	return &Handle{file: file}
}

func (hdl *Handle) initStreamIfNeeded() error {
	if hdl.stream != nil {
		return nil
	}

	// Initialize the stream lazily to avoid i/o on open()
	rawStream, err := hdl.fs.bk.Cat(hdl.file.Content())
	if err != nil {
		return err
	}

	hdl.stream, err = mio.NewOutStream(rawStream, hdl.file.Key())
	if err != nil {
		return err
	}

	return nil
}

func (hdl *Handle) Read(buf []byte) (int, error) {
	var err error

	hdl.lock.Lock()
	defer hdl.lock.Unlock()

	if err := hdl.initStreamIfNeeded(); err != nil {
		return 0, err
	}

	n, err := io.ReadFull(hdl.stream, buf)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return 0, err
	}

	return n, nil
}

func (hdl *Handle) Write(buf []byte) (int, error) {
	hdl.lock.Lock()
	defer hdl.lock.Unlock()

	// TODO: Is this a race-condition?
	size := hdl.file.Size()

	if hdl.layer == nil {
		if err := hdl.initStreamIfNeeded(); err != nil {
			return 0, err
		}

		hdl.layer = overlay.NewLayer(hdl.stream)
		hdl.layer.Truncate(int64(size))
	}

	n, err := hdl.layer.Write(buf)
	if err != nil {
		return n, err
	}

	minSize := uint64(hdl.layer.MinSize())
	if size < minSize {
		hdl.file.SetSize(minSize)
	}

	return n, nil
}

func (hdl *Handle) Seek(offset int64, whence int) (int64, error) {
	hdl.lock.Lock()
	defer hdl.lock.Unlock()

	n1, err := hdl.layer.Seek(offset, whence)
	if err != nil {
		return 0, err
	}

	n2, err := hdl.stream.Seek(offset, whence)
	if err != nil {
		return 0, err
	}

	if n1 != n2 {
		return 0, fmt.Errorf("memory and stream seek pos diverged")
	}

	return n1, nil
}

func (hdl *Handle) Truncate(size uint64) error {
	hdl.lock.Lock()
	defer hdl.lock.Unlock()

	// TODO: Race condition?
	hdl.file.SetSize(size)
	hdl.layer.Truncate(int64(size))
	return nil
}

func (hdl *Handle) Flush() error {
	hdl.lock.Lock()
	defer hdl.lock.Unlock()

	// flush unsets the layer, so we don't flush twice.
	if hdl.layer == nil {
		return nil
	}

	defer func() {
		hdl.layer = nil
	}()

	n, err := hdl.layer.Seek(0, os.SEEK_SET)
	if err != nil {
		return err
	}

	if n != 0 {
		return fmt.Errorf("seek offset is not 0")
	}

	path := hdl.file.Path()
	if err := hdl.fs.Stage(path, hdl.layer); err != nil {
		return err
	}

	return hdl.layer.Close()
}

func (hdl *Handle) Close() error {
	return hdl.Flush()
}
