package httpipfs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/sahib/brig/catfs/mio"
	h "github.com/sahib/brig/util/hashlib"
	shell "github.com/sahib/go-ipfs-api"
)

func cat(s *shell.Shell, path string, offset int64) (io.ReadCloser, error) {
	rb := s.Request("cat", path)
	rb.Option("offset", offset)
	resp, err := rb.Send(context.Background())
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, resp.Error
	}

	return resp.Output, nil
}

type streamWrapper struct {
	io.ReadCloser
	nd   *Node
	hash h.Hash
	off  int64
	size int64
}

func (sw *streamWrapper) Read(buf []byte) (int, error) {
	n, err := sw.ReadCloser.Read(buf)
	if err != nil {
		return n, err
	}

	sw.off += int64(n)
	return n, err
}

func (sw *streamWrapper) WriteTo(w io.Writer) (int64, error) {
	return io.Copy(w, sw)
}

func (sw *streamWrapper) cachedSize() (int64, error) {
	ctx := context.Background()
	resp, err := sw.nd.sh.Request(
		"files/stat",
		"/ipfs/"+sw.hash.B58String(),
	).Send(ctx)

	if err != nil {
		return -1, err
	}

	defer resp.Close()

	if resp.Error != nil {
		return -1, resp.Error
	}

	raw := struct {
		Size int64
	}{}

	if err := json.NewDecoder(resp.Output).Decode(&raw); err != nil {
		return -1, err
	}

	return raw.Size, nil
}

func (sw *streamWrapper) getAbsOffset(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		sw.off = offset
		return offset, nil
	case io.SeekCurrent:
		sw.off += offset
		return sw.off, nil
	case io.SeekEnd:
		size, err := sw.cachedSize()
		if err != nil {
			return -1, err
		}

		sw.off = size + offset
		return sw.off, nil
	default:
		return -1, fmt.Errorf("invalid whence: %v", whence)
	}
}

func (sw *streamWrapper) Seek(offset int64, whence int) (int64, error) {
	absOffset, err := sw.getAbsOffset(offset, whence)
	if err != nil {
		return -1, err
	}

	rc, err := cat(sw.nd.sh, sw.hash.B58String(), absOffset)
	if err != nil {
		return -1, err
	}

	if sw.ReadCloser != nil {
		sw.ReadCloser.Close()
	}

	sw.off = absOffset
	sw.ReadCloser = rc
	return absOffset, nil
}

// Cat returns a stream associated with `hash`.
func (nd *Node) Cat(hash h.Hash) (mio.Stream, error) {
	rc, err := cat(nd.sh, hash.B58String(), 0)
	if err != nil {
		return nil, err
	}

	return &streamWrapper{
		nd:         nd,
		hash:       hash,
		ReadCloser: rc,
		off:        0,
		size:       -1,
	}, nil
}

// Add puts the contents of `r` into IPFS and returns its hash.
func (nd *Node) Add(r io.Reader) (h.Hash, error) {
	hs, err := nd.sh.Add(r)
	if err != nil {
		return nil, err
	}

	return h.FromB58String(hs)
}
