package daemon

import (
	"encoding/binary"
	"fmt"
	"io"

	protobuf "github.com/gogo/protobuf/proto"
)

// send transports a msg over conn with a size header.
func send(conn io.Writer, msg protobuf.Message) error {
	data, err := protobuf.Marshal(msg)
	if err != nil {
		return err
	}

	sizeBuf := make([]byte, binary.MaxVarintLen64)
	binary.PutUvarint(sizeBuf, uint64(len(data)))

	n, err := conn.Write(sizeBuf)
	if err != nil {
		return err
	}

	if n < len(sizeBuf) {
		return io.ErrShortWrite
	}

	n, err = conn.Write(data)
	if err != nil {
		return err
	}

	if n < len(data) {
		return io.ErrShortWrite
	}

	return nil
}

// TODO: use little endian, not varint.

// recv reads a size-prefixed protobuf buffer
func recv(conn io.Reader, msg protobuf.Message) error {
	sizeBuf := make([]byte, binary.MaxVarintLen64)
	n, err := conn.Read(sizeBuf)
	if err != nil {
		return err
	}

	size, _ := binary.Uvarint(sizeBuf[:n])
	if size > 1*1024*1024 {
		return fmt.Errorf("Message too large: %d", size)
	}

	buf := make([]byte, size)
	_, err = conn.Read(buf)
	if err != nil {
		return err
	}

	err = protobuf.Unmarshal(buf, msg)
	if err != nil {
		return err
	}

	return nil
}
