// Package protocol implements a encoder and decoder for a protobuf based
// communication protocol. Any proto.Message might be send and received.
// Optionally, the messages might be compressed using snappy.
package protocol

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
)

const (
	// Maximum size a single message may have:
	MessageSizeLimit = 5 * 1024 * 1024
)

var (
	ErrNoReader = errors.New("Protocol was created without reader part")
	ErrNoWriter = errors.New("Protocol was created without writer part")
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

func (p *Protocol) Send(msg proto.Message) error {
	if p.w == nil {
		return ErrNoWriter
	}

	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	if p.compress {
		data = snappy.Encode(nil, data)
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

func (p *Protocol) Recv(resp proto.Message) error {
	if p.r == nil {
		return ErrNoReader
	}

	sizeBuf := make([]byte, 4)
	if _, err := io.ReadAtLeast(p.r, sizeBuf, len(sizeBuf)); err != nil {
		return err
	}

	size := binary.LittleEndian.Uint32(sizeBuf)
	if size > MessageSizeLimit {
		return ErrMessageTooBig{size}
	}

	buf := bytes.NewBuffer(make([]byte, 0, size))

	if _, err := io.CopyN(buf, p.r, int64(size)); err != nil {
		return err
	}

	data := buf.Bytes()

	var err error
	if p.compress {
		if data, err = snappy.Decode(nil, data); err != nil {
			return err
		}
	}

	if err := proto.Unmarshal(data, resp); err != nil {
		return err
	}

	return nil
}

// ProtocolEncoder is a utility that uses `Protocol` to
// encode an arbitary protobuf message to a byte slice.
type ProtocolEncoder struct {
	p *Protocol
	b *bytes.Buffer
}

// NewProtocolEncoder returns a valid ProtocolEncoder, which
// will compress it's data when flagged accordingly.
// If tnl is non-nil it will also encrypt the data.
func NewProtocolEncoder(compress bool) *ProtocolEncoder {
	b := &bytes.Buffer{}
	return &ProtocolEncoder{p: NewProtocolWriter(b, compress), b: b}
}

// Encode returns a byte representation of `msg`.
func (pe *ProtocolEncoder) Encode(msg proto.Message) ([]byte, error) {
	if err := pe.p.Send(msg); err != nil {
		return nil, err
	}

	data := pe.b.Bytes()
	pe.b.Reset()
	return data, nil
}

// ProtocolDecoder is a utility that uses `Protocol` to
// decode a byte representation of a message to a protobuf message.
type ProtocolDecoder struct {
	p *Protocol
}

// NewProtocolDecoder returns a new ProtocolDecoder which
// decompresses the data passed into it if needed.
func NewProtocolDecoder(decompress bool) *ProtocolDecoder {
	return &ProtocolDecoder{p: NewProtocolReader(nil, decompress)}
}

// Decode decodes `data` and writes the result into `msg`.
func (pd *ProtocolDecoder) Decode(msg proto.Message, data []byte) error {
	pd.p.r = bytes.NewReader(data)

	if err := pd.p.Recv(msg); err != nil {
		return err
	}

	return nil
}
