package catfs

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/disorganizer/brig/catfs/mio"
	"github.com/disorganizer/brig/catfs/mio/overlay"
	n "github.com/disorganizer/brig/catfs/nodes"
)

var (
	ErrIsClosed = errors.New("File handle is closed")
)

type Handle struct {
	fs          *FS
	file        *n.File
	lock        sync.Mutex
	layer       *overlay.Layer
	stream      mio.Stream
	wasModified bool
	isClosed    bool
}

func newHandle(fs *FS, file *n.File) *Handle {
	return &Handle{
		fs:   fs,
		file: file,
	}
}

func (hdl *Handle) initStreamIfNeeded() error {
	if hdl.stream != nil {
		return nil
	}

	// Initialize the stream lazily to avoid I/O on open()
	rawStream, err := hdl.fs.bk.Cat(hdl.file.Content())
	if err != nil {
		return err
	}

	// Stack the mio stack on top:
	hdl.stream, err = mio.NewOutStream(rawStream, hdl.file.Key())
	if err != nil {
		return err
	}

	hdl.layer = overlay.NewLayer(hdl.stream)
	hdl.layer.Truncate(int64(hdl.file.Size()))
	return nil
}

func (hdl *Handle) Read(buf []byte) (int, error) {
	var err error

	hdl.lock.Lock()
	defer hdl.lock.Unlock()

	if hdl.isClosed {
		return 0, ErrIsClosed
	}

	if err := hdl.initStreamIfNeeded(); err != nil {
		return 0, err
	}

	n, err := io.ReadFull(hdl.layer, buf)
	isEOF := err != io.ErrUnexpectedEOF || err != io.EOF
	if err != nil && !isEOF {
		return 0, err
	}

	if isEOF {
		return n, io.EOF
	}

	return n, nil
}

func (hdl *Handle) Write(buf []byte) (int, error) {
	hdl.lock.Lock()
	defer hdl.lock.Unlock()

	if hdl.isClosed {
		return 0, ErrIsClosed
	}

	if err := hdl.initStreamIfNeeded(); err != nil {
		return 0, err
	}

	// Currently, we do not check if the file was actually modified
	// (i.e. data changed compared to before)
	hdl.wasModified = true

	n, err := hdl.layer.Write(buf)
	if err != nil {
		return n, err
	}

	// Advance the write pointer when writing things to the buffer.
	if _, err := hdl.stream.Seek(int64(n), os.SEEK_CUR); err != nil && err != io.EOF {
		return n, err
	}

	minSize := uint64(hdl.layer.MinSize())
	if hdl.file.Size() < minSize {
		hdl.fs.mu.Lock()
		hdl.file.SetSize(minSize)
		hdl.fs.mu.Unlock()

		// Also auto-truncate on every write.
		hdl.layer.Truncate(int64(minSize))
	}

	return n, nil
}

func (hdl *Handle) Seek(offset int64, whence int) (int64, error) {
	hdl.lock.Lock()
	defer hdl.lock.Unlock()

	if hdl.isClosed {
		return 0, ErrIsClosed
	}

	if err := hdl.initStreamIfNeeded(); err != nil {
		return 0, err
	}

	n, err := hdl.layer.Seek(offset, whence)
	if err != nil {
		return 0, err
	}

	return n, nil
}

// Truncate truncates the file to a specific length.
func (hdl *Handle) Truncate(size uint64) error {
	hdl.lock.Lock()
	defer hdl.lock.Unlock()

	if hdl.isClosed {
		return ErrIsClosed
	}

	if err := hdl.initStreamIfNeeded(); err != nil {
		return err
	}

	hdl.fs.mu.Lock()
	hdl.file.SetSize(size)
	hdl.fs.mu.Unlock()

	hdl.layer.Truncate(int64(size))
	return nil
}

// unlocked version of Flush()
func (hdl *Handle) flush() error {
	// flush unsets the layer, so we don't flush twice.
	if hdl.layer == nil {
		return nil
	}

	// Make sure that hdl.layer is unset in any case.
	defer func() {
		hdl.layer = nil
	}()

	// No need to flush anything if no write calles happened.
	if !hdl.wasModified {
		return nil
	}

	// Jump back to the beginning of the file, since fs.Stage()
	// should read all content starting from there.
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

func (hdl *Handle) Flush() error {
	hdl.lock.Lock()
	defer hdl.lock.Unlock()

	if hdl.isClosed {
		return ErrIsClosed
	}

	return hdl.flush()
}

func (hdl *Handle) Close() error {
	hdl.lock.Lock()
	defer hdl.lock.Unlock()

	if hdl.isClosed {
		return ErrIsClosed
	}

	hdl.isClosed = true
	return hdl.flush()
}

func (hdl *Handle) Path() string {
	hdl.lock.Lock()
	defer hdl.lock.Unlock()

	return hdl.file.Path()
}
