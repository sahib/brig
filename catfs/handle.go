package catfs

import (
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/sahib/brig/catfs/mio"
	"github.com/sahib/brig/catfs/mio/overlay"
	n "github.com/sahib/brig/catfs/nodes"
)

var (
	// ErrIsClosed is returned when an operation is performed on an already
	// closed file.
	ErrIsClosed = errors.New("File handle is closed")
)

// Handle is a emulation of a os.File handle, as returned by os.Open()
// It supports the usual operations like open, read, write and seek.
// Take note that the flushing operation currently is quite expensive.
type Handle struct {
	fs          *FS
	file        *n.File
	lock        sync.Mutex
	layer       *overlay.Layer
	stream      mio.Stream
	wasModified bool
	isClosed    bool
	readOnly    bool
}

func newHandle(fs *FS, file *n.File, readOnly bool) *Handle {
	return &Handle{
		fs:       fs,
		file:     file,
		readOnly: readOnly,
	}
}

func (hdl *Handle) initStreamIfNeeded() error {
	if hdl.stream != nil {
		return nil
	}

	// Initialize the stream lazily to avoid I/O on open()
	rawStream, err := hdl.fs.bk.Cat(hdl.file.BackendHash())
	if err != nil {
		return err
	}

	// Stack the mio stack on top:
	hdl.stream, err = mio.NewOutStream(
		rawStream,
		hdl.file.IsRaw(),
		hdl.file.Key(),
	)
	if err != nil {
		return err
	}

	hdl.layer = overlay.NewLayer(hdl.stream)
	hdl.layer.Truncate(int64(hdl.file.Size()))
	hdl.layer.SetSize(int64(hdl.file.Size()))
	return nil
}

// Read will try to fill `buf` as much as possible.
// The seek pointer will be advanced by the number of bytes written.
// Take care, `buf` might still have contents, even if io.EOF was returned.
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

	// TODO: not sure if that makes sense...
	// we should just read whatever the underlying stream thinks it has.
	n, err := io.ReadFull(hdl.layer, buf)
	isEOF := (err == io.ErrUnexpectedEOF || err == io.EOF)
	if err != nil && !isEOF {
		return 0, err
	}

	if isEOF {
		return n, io.EOF
	}

	return n, nil
}

// Write will write the contents of `buf` to the current position.
// It will return the number of currently written bytes.
func (hdl *Handle) Write(buf []byte) (int, error) {
	hdl.lock.Lock()
	defer hdl.lock.Unlock()

	if hdl.readOnly {
		return 0, ErrReadOnly
	}

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
	if _, err := hdl.stream.Seek(int64(n), io.SeekCurrent); err != nil && err != io.EOF {
		return n, err
	}

	minSize := uint64(hdl.layer.MinSize())
	if hdl.file.Size() < minSize {
		hdl.fs.mu.Lock()
		hdl.file.SetSize(minSize)

		// Make sure to save the size change:
		if err := hdl.fs.lkr.StageNode(hdl.file); err != nil {
			hdl.fs.mu.Unlock()
			return 0, err
		}

		hdl.fs.mu.Unlock()

		// Also auto-truncate on every write.
		hdl.layer.Truncate(int64(minSize))
	}

	return n, nil
}

// Writes data from `buf` at offset `off` counted from the start (0 offset).
// Mimics `WriteAt` from `io` package https://golang.org/pkg/io/#WriterAt
func (hdl *Handle) WriteAt(buf []byte, off int64) (n int, err error) {
	hdl.lock.Lock()
	defer hdl.lock.Unlock()

	if hdl.readOnly {
		return 0, ErrReadOnly
	}

	if hdl.isClosed {
		return 0, ErrIsClosed
	}

	if err := hdl.initStreamIfNeeded(); err != nil {
		return 0, err
	}

	hdl.wasModified = true
	n, err = hdl.layer.WriteAt(buf, off)
	return n, err
}

// Seek will jump to the `offset` relative to `whence`.
// There next read and write operation will start from this point.
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

	if hdl.readOnly {
		return ErrReadOnly
	}

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

	// No need to flush anything if no write calls happened.
	if !hdl.wasModified {
		return nil
	}

	// Make sure that hdl.layer is unset in any case.
	// but only do that in case of real writes.
	defer func() {
		hdl.layer = nil
		hdl.stream = nil
		hdl.wasModified = false
	}()

	// Jump back to the beginning of the file, since fs.Stage()
	// should read all content starting from there.
	n, err := hdl.layer.Seek(0, io.SeekStart)
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

// Flush makes sure to write the current state to the backend.
// Please remember that this method is currently pretty expensive.
func (hdl *Handle) Flush() error {
	hdl.lock.Lock()
	defer hdl.lock.Unlock()

	if hdl.readOnly {
		return ErrReadOnly
	}

	if hdl.isClosed {
		return ErrIsClosed
	}

	return hdl.flush()
}

// Close will finalize the file. It should not be used after.
// This will call flush if it did not happen yet.
func (hdl *Handle) Close() error {
	hdl.lock.Lock()
	defer hdl.lock.Unlock()

	if hdl.isClosed {
		return ErrIsClosed
	}

	hdl.isClosed = true
	return hdl.flush()
}

// Path returns the absolute path of the file.
func (hdl *Handle) Path() string {
	hdl.lock.Lock()
	defer hdl.lock.Unlock()

	return hdl.file.Path()
}
