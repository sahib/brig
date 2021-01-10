package server

import (
	"sync"

	"github.com/djherbis/buffer"
	"github.com/sahib/brig/catfs"
	"github.com/sahib/brig/server/capnp"
	"gopkg.in/djherbis/nio.v2"
)

const (
	memBufferSize = 64 * 1024
)

var (
	sendBufferPool = sync.Pool{
		New: func() interface{} {
			return buffer.New(memBufferSize)
		},
	}
)

/////////////////
// SERVER SIDE //
/////////////////

type streamServer struct {
	err     *error
	errCond *sync.Cond
	pr      *nio.PipeReader
	pw      *nio.PipeWriter
	buf     buffer.Buffer
}

func (ss *streamServer) doStage(base *base, repoPath string) {
	err := base.withFsFromPath(repoPath, func(url *URL, fs *catfs.FS) error {
		if err := fs.Stage(url.Path, ss.pr); err != nil {
			return err
		}

		base.notifyFsChangeEvent()
		return nil
	})

	ss.errCond.L.Lock()
	defer ss.errCond.L.Unlock()

	ss.err = &err

	// Wake up any calls waiting in Done()
	ss.errCond.Broadcast()
}

func newStreamServer(base *base, repoPath string) *streamServer {
	buf := sendBufferPool.Get().(buffer.Buffer)
	pr, pw := nio.Pipe(buf)
	ss := &streamServer{
		pr:      pr,
		pw:      pw,
		buf:     buf,
		errCond: sync.NewCond(&sync.Mutex{}),
	}

	// already start staging, but fill reader only
	// chunk by chunk as they arrive:
	go ss.doStage(base, repoPath)
	return ss
}

func (ss *streamServer) hasFinished() (bool, error) {
	ss.errCond.L.Lock()
	defer ss.errCond.L.Unlock()
	if ss.err != nil {
		return true, *ss.err
	}

	return false, nil
}

// SendChunk is called when the client sends one block of data.
func (ss *streamServer) SendChunk(call capnp.FS_StageStream_sendChunk) error {
	if finished, err := ss.hasFinished(); finished {
		// return the last error if already done, err might be nil here.
		// This is here to protect against more SendChunk() calls after Done()
		return err
	}

	data, err := call.Params.Chunk()
	if err != nil {
		return err
	}

	// Send the data over the pipe to the stage reader:
	// No actual copying done, here. This waits until the data was read.
	_, err = ss.pw.Write(data)
	return err
}

func (ss *streamServer) Done(call capnp.FS_StageStream_done) error {
	// Closing the pipe writer will trigger a io.EOF in the reader part.
	// This will make Stage() return after some post processing.
	if err := ss.pw.Close(); err != nil {
		return err
	}

	ss.errCond.L.Lock()
	defer ss.errCond.L.Unlock()

	// Wait until Stage() actually returned:
	for ss.err == nil {
		ss.errCond.Wait()
	}

	ss.buf.Reset()
	sendBufferPool.Put(ss.buf)
	return *ss.err
}
