// +build !windows

package fuse

import (
	"io"
	"sync"
	"time"
	"syscall"
	"bytes"

	"context"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"bazil.org/fuse/fuseutil"
	"github.com/sahib/brig/catfs"
	log "github.com/sirupsen/logrus"
)

// Handle is an open Entry.
type Handle struct {
	mu sync.Mutex
	fd *catfs.Handle
	m  *Mount
	// number of write-capable handles currently open
	writers uint
	// only valid if writers > 0, data used as a buffer for write operations
	data []byte
	wasModified bool

}

func (hd *Handle) loadData(path string) (error) {
	hd.data = nil
	hd.wasModified = false
	fd, err := hd.m.fs.Open(path)
	if err != nil {
		return errorize("file-loadData", err)
	}
	var bufSize int = 128*1024
	buf := make([]byte, bufSize)
	var data []byte
	for {
		n, err := fd.Read(buf)
		isEOF := (err == io.ErrUnexpectedEOF || err == io.EOF)
		if err != nil && !isEOF {
			return errorize("file-loadData", err)
		}
		data = append(data, buf[:n]...)
		if isEOF {
			break
		}
	}
	hd.data = data
	return nil
}


// Read is called to read a block of data at a certain offset.
func (hd *Handle) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	hd.mu.Lock()
	defer hd.mu.Unlock()
	defer logPanic("handle: read")

	log.WithFields(log.Fields{
		"path":   hd.fd.Path(),
		"offset": req.Offset,
		"size":   req.Size,
	}).Debugf("fuse: handle: read")

	// if we have writers we need to suply response from the data buffer
	if hd.writers != 0 {
		fuseutil.HandleRead(req, resp, hd.data)
		return nil
	}

	// otherwise we will read from the brig filesystem directly
	newOff, err := hd.fd.Seek(req.Offset, io.SeekStart)
	if err != nil {
		return errorize("handle-read-seek", err)
	}

	if newOff != req.Offset {
		log.Warningf("read/seek offset differs (want %d, got %d)", req.Offset, newOff)
	}

	n, err := hd.fd.Read(resp.Data[:req.Size])
	if err != nil && err != io.EOF {
		return errorize("handle-read-io", err)
	}

	resp.Data = resp.Data[:n]
	return nil
}

const maxInt = int(^uint(0) >> 1)

// Write is called to write a block of data at a certain offset.
func (hd *Handle) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	hd.mu.Lock()
	defer hd.mu.Unlock()
	defer logPanic("handle: write")

	log.Debugf(
		"fuse-write: %s (off: %d size: %d)",
		hd.fd.Path(),
		req.Offset,
		len(req.Data),
	)

	// expand the buffer if necessary
	newLen := req.Offset + int64(len(req.Data))
	if newLen > int64(maxInt) {
		return fuse.Errno(syscall.EFBIG)
	}
	if newLen := int(newLen); newLen > len(hd.data) {
		hd.data = append(hd.data, make([]byte, newLen-len(hd.data))...)
	}

	n := copy(hd.data[req.Offset:], req.Data)
	hd.wasModified = true
	resp.Size = n
	return nil
}

// Flush is called to make sure all written contents get synced to disk.
func (hd *Handle) Flush(ctx context.Context, req *fuse.FlushRequest) error {
	return hd.flush()
}

// flush does the actual adding to brig.
func (hd *Handle) flush() error {
	hd.mu.Lock()
	defer hd.mu.Unlock()

	log.Debugf("fuse-flush: %v", hd.fd.Path())
	defer logPanic("handle: flush")

	if !hd.wasModified {
		return nil
	}
	r := bytes.NewReader(hd.data)
	if err := hd.m.fs.Stage(hd.fd.Path(), r); err != nil {
		return errorize("handle-flush", err)
	}
	hd.wasModified = false

	notifyChange(hd.m, 500*time.Millisecond)
	return nil
}

// Release is called to close this handle.
func (hd *Handle) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
	defer logPanic("handle: release")
	log.Debugf("fuse-release: %v", hd.fd.Path())

	if req.Flags.IsReadOnly() {
		// we don't need to track read-only handles
		return nil
	}

	hd.mu.Lock()
	defer hd.mu.Unlock()

	hd.writers--
	if hd.writers == 0 {
		hd.data = nil
	}
	return nil
}


// Truncates (or extends) data to the desired size
func (hd *Handle) truncate(size uint64) error {
	log.Debugf("fuse-truncate: %v to size %d", hd.fd.Path(), size)
	defer logPanic("handle: truncate")

	if size > uint64(maxInt) {
		return fuse.Errno(syscall.EFBIG)
	}
	newLen := int(size)
	switch {
	case newLen > len(hd.data):
		hd.data = append(hd.data, make([]byte, newLen-len(hd.data))...)
		hd.wasModified = true
	case newLen < len(hd.data):
		hd.data = hd.data[:newLen]
		hd.wasModified = true
	}
	return nil
}

// Compiler checks to see if we got all the interfaces right:
var _ = fs.HandleFlusher(&Handle{})
var _ = fs.HandleReader(&Handle{})
var _ = fs.HandleReleaser(&Handle{})
var _ = fs.HandleWriter(&Handle{})
