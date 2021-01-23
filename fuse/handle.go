// +build !windows

package fuse

import (
	"io"
	"sync"
	"syscall"
	"time"

	"context"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/sahib/brig/catfs"
	log "github.com/sirupsen/logrus"
)

// Handle is an open Entry.
type Handle struct {
	mu                    sync.Mutex
	fd                    *catfs.Handle
	m                     *Mount
	wasModified           bool
	currentFileReadOffset int64
}

// Read is called to read a block of data at a certain offset.
func (hd *Handle) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	hd.mu.Lock()
	defer hd.mu.Unlock()
	defer logPanic("handle: read")

	log.Debugf(
		"fuse-Read: %s (off: %d size: %d)",
		hd.fd.Path(),
		req.Offset,
		req.Size,
	)

	newOff := hd.currentFileReadOffset
	if req.Offset != hd.currentFileReadOffset {
		var err error
		newOff, err = hd.fd.Seek(req.Offset, io.SeekStart)
		if err != nil {
			return errorize("handle-read-seek", err)
		}
	}

	if newOff != req.Offset {
		log.Warningf("read/seek offset differs (want %d, got %d)", req.Offset, newOff)
	}

	n, err := hd.fd.Read(resp.Data[:req.Size])
	if err != nil && err != io.EOF {
		return errorize("handle-read-io", err)
	}
	hd.currentFileReadOffset = newOff + int64(n)

	resp.Data = resp.Data[:n]
	return nil
}

const maxInt = int(^uint(0) >> 1)

// Write is called to write a block of data at a certain offset.
// Note: do not assume that Write requests come in `fifo` order from the OS level!!!
// I.e. during `cp largeFile /brig-fuse-mount/newFile`
// the kernel might occasionally send write requests with blocks out of order!!!
// In other words stream-like optimizations are not possible .
func (hd *Handle) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	start := time.Now()
	hd.mu.Lock()
	defer hd.mu.Unlock()
	defer logPanic("handle: write")

	log.Debugf(
		"fuse-Write: %s (off: %d size: %d)",
		hd.fd.Path(),
		req.Offset,
		len(req.Data),
	)

	// Offset seems to be always provided from the start (i.e. 0)
	n, err := hd.writeAt(req.Data, req.Offset)
	resp.Size = n
	if err != nil {
		return errorize("handle-write-io", err)
	}
	if n != len(req.Data) {
		log.Panicf("written amount %d is not equal to requested %d", n, len(req.Data))
		return err
	}
	log.Infof("fuse: Write time %v for %d bytes", time.Since(start), n)
	hd.wasModified = true
	return nil
}

// Writes data from `buf` at offset `off` counted from the start (0 offset).
// Mimics `WriteAt` from `io` package https://golang.org/pkg/io/#WriterAt
// Main idea is not bother with Seek pointer, since underlying `overlay` works
// with intervals in memory and we do not need to `Seek` the backend which is very time expensive.
func (hd *Handle) writeAt(buf []byte, off int64) (n int, err error) {
	n, err = hd.fd.WriteAt(buf, off)
	if n != len(buf) || err != nil {
		log.Errorf("fuse: were not able to save %d bytes at offset %d", len(buf), off)
	}
	return n, err
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
	start := time.Now()
	if err := hd.fd.Flush(); err != nil {
		return errorize("handle-flush", err)
	}
	log.Infof("fuse: Flashed `%s` in %v", hd.fd.Path(), time.Since(start))
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

	err := hd.flush()
	if err != nil {
		return errorize("handle-release", err)
	}

	return nil
}

// Truncates (or extends) data to the desired size
func (hd *Handle) truncate(size uint64) error {
	log.Debugf("fuse-truncate: %v to size %d", hd.fd.Path(), size)
	defer logPanic("handle: truncate")
	err := hd.fd.Truncate(size)

	return err
}

// Poll checks that the handle is ready for I/O or not
func (hd *Handle) Poll(ctx context.Context, req *fuse.PollRequest, resp *fuse.PollResponse) error {
	// Comment taken verbatim from fs/serve.go of bazil.org/fuse:
	// Poll checks whether the handle is currently ready for I/O, and
	// may request a wakeup when it is.
	//
	// Poll should always return quickly. Clients waiting for
	// readiness can be woken up by passing the return value of
	// PollRequest.Wakeup to fs.Server.NotifyPollWakeup or
	// fuse.Conn.NotifyPollWakeup.
	//
	// To allow supporting poll for only some of your Nodes/Handles,
	// the default behavior is to report immediate readiness. If your
	// FS does not support polling and you want to minimize needless
	// requests and log noise, implement NodePoller and return
	// syscall.ENOSYS.
	//
	// The Go runtime uses epoll-based I/O whenever possible, even for
	// regular files.

	// Here we implement a dummy response which reports "I am ready".
	// The access separation is handled by mutex, so go-rutines
	// will have to be blocked but its ok. We do not expect many
	// processes working with the same file

	// default always ready mask
	resp.REvents = fuse.DefaultPollMask

	// We also return ENOSYS error, which sort of invalidate our response,
	// the ENOSYS indicates that this call is not supported
	return syscall.ENOSYS
}

// Compiler checks to see if we got all the interfaces right:
var _ = fs.HandleFlusher(&Handle{})
var _ = fs.HandleReader(&Handle{})
var _ = fs.HandleReleaser(&Handle{})
var _ = fs.HandleWriter(&Handle{})
var _ = fs.HandlePoller(&Handle{})
