package server

import (
	"fmt"
	"io"
	"net"

	log "github.com/Sirupsen/logrus"
	"github.com/sahib/brig/catfs"
)

func getNextFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}

	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func bootTransferServer(fs *catfs.FS, bindHost, path string) (int, error) {
	port, err := getNextFreePort()
	if err != nil {
		return 0, err
	}

	stream, err := fs.Cat(path)
	if err != nil {
		return 0, err
	}

	lst, err := net.Listen("tcp", fmt.Sprintf("%s:%d", bindHost, port))

	if err != nil {
		stream.Close()
		return 0, err
	}

	go func() {
		defer lst.Close()
		defer stream.Close()

		conn, err := lst.Accept()
		if err != nil {
			log.Warningf("Failed to accept connection on %d: %v", port, err)
			return
		}

		defer conn.Close()

		n, err := io.Copy(conn, stream)
		if err != nil {
			log.Warningf("IO failed for path %s on %d: %v", path, port, err)
			return
		}

		log.Infof("Wrote %d bytes of `%s` over port %d", n, path, port)
	}()

	return port, nil
}

func bootReceiveServer(bindHost string, fn func(conn net.Conn) error) (int, error) {
	port, err := getNextFreePort()
	if err != nil {
		return -1, err
	}

	lst, err := net.Listen("tcp", fmt.Sprintf("%s:%d", bindHost, port))
	if err != nil {
		return -1, err
	}

	go func() {
		defer lst.Close()

		conn, err := lst.Accept()
		if err != nil {
			log.Warningf("Failed to accept connection on %d: %v", port, err)
			return
		}

		defer conn.Close()

		if err := fn(conn); err != nil {
			log.Debugf("handling of data at port %d failed: %v", port, err)
		}
	}()

	return port, nil
}
