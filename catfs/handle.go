package catfs

import (
	"sync"

	"github.com/disorganizer/brig/catfs/mio"
	h "github.com/disorganizer/brig/util/hashlib"
)

// TODO: Implement, using the implemenation found in fuse.
type Handle struct {
	path   string
	hash   h.Hash
	lock   sync.Mutex
	stream mio.Stream
}

func newHandle(stream mio.Stream) (*Handle, error) {
	return nil, nil
}

func (hdl *Handle) Read(buf []byte) (int, error) {
	return 0, nil
}

func (hdl *Handle) Write(buf []byte) (int, error) {
	return 0, nil
}

func (hdl *Handle) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
}

func (hdl *Handle) Truncate(size uint64) error {
	return nil
}

func (hdl *Handle) Flush() error {
	return nil
}

func (hdl *Handle) Close() error {
	return hdl.Flush()
}
