// Package protocol implements a encoder and decoder for a protobuf based
// communication protocol. Any protobuf.Message might be send and received.
// Optionally, the messages might be compressed using snappy.
package protocol

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	protobuf "github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
)

const (
	// Maximum size a single message may have:
	MessageSizeLimit = 5 * 1024 * 1024
)

// ErrMalformed is returned when the size tag is missing
// or is too short. Bad data in the payload will return a protobuf error.
var (
	ErrMalformed = errors.New("Malformed protocol data (not enough data)")
	ErrNoReader  = errors.New("Protocol was created without reader part")
	ErrNoWriter  = errors.New("Protocol was created without writer part")
)

// ErrMessageTooBig is returned when the received message is bigger
// than MessageSizeLimit and is therefore refused for security reasons.
type ErrMessageTooBig struct {
	size uint32
}

func (e ErrMessageTooBig) Error() string {
	return fmt.Sprintf("Message is too big (%d bytes, maximum: %d)", e.size, MessageSizeLimit)
}

type Protocol struct {
	r        io.Reader
	w        io.Writer
	compress bool
}

func NewProtocol(rw io.ReadWriter, compress bool) *Protocol {
	return &Protocol{r: rw, w: rw, compress: compress}
}

func NewProtocolReader(r io.Reader, compress bool) *Protocol {
	return &Protocol{r: r, w: nil, compress: compress}
}

func NewProtocolWriter(w io.Writer, compress bool) *Protocol {
	return &Protocol{r: nil, w: w, compress: compress}
}

func (p *Protocol) Send(msg protobuf.Message) error {
	if p.w == nil {
		return ErrNoWriter
	}

	data, err := protobuf.Marshal(msg)
	if err != nil {
		return err
	}

	if p.compress {
		data = snappy.Encode(data, data)
	}

	sizeBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(sizeBuf, uint32(len(data)))
	if _, err := p.w.Write(sizeBuf); err != nil {
		return err
	}

	if _, err := p.w.Write(data); err != nil {
		return err
	}

	return nil
}

func (p *Protocol) Recv(resp protobuf.Message) error {
	if p.r == nil {
		return ErrNoReader
	}

	sizeBuf := make([]byte, 4)
	n, err := p.r.Read(sizeBuf)

	if err != nil {
		return err
	}

	if n < 4 {
		return ErrMalformed
	}

	size := binary.LittleEndian.Uint32(sizeBuf)
	if size > MessageSizeLimit {
		return ErrMessageTooBig{size}
	}

	data := make([]byte, 0, size)
	buf := bytes.NewBuffer(data)

	if _, err = io.CopyN(buf, p.r, int64(size)); err != nil {
		return err
	}

	if p.compress {
		data, err = snappy.Decode(buf.Bytes(), buf.Bytes())
		if err != nil {
			return err
		}
	} else {
		data = buf.Bytes()
	}

	if err := protobuf.Unmarshal(data, resp); err != nil {
		return err
	}

	return nil
}
