package mock

import (
	"fmt"
	"net"
	"time"
)

// LoopPipe is similar to net.Pipe() but sets up
// a fully buffered asynchronous net.Conns.
//
// The problem with that net.Pipe() is that it will
// block upon a write until the other end reads.
// This means it's not full duplex and does thus
// not work well with AuthReadWriter, where both
// ends first send their pubkey.
func LoopPipe() (net.Conn, net.Conn, error) {
	port := 12345
	addr := fmt.Sprintf("localhost:%d", port)
	lst, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, nil, err
	}

	connCh := make(chan net.Conn)

	listen := func() {
		conn, err := lst.Accept()
		if err != nil {
			panic(err)
		}

		if err := lst.Close(); err != nil {
			panic(err)
		}

		connCh <- conn
	}

	go listen()

	time.Sleep(50 * time.Millisecond)
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return nil, nil, err
	}

	return conn, <-connCh, nil
}
