package server

import (
	"fmt"
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

func bootTransferServer(fs *catfs.FS, bindHost string, copyFn func(conn net.Conn)) (int, error) {
	port, err := getNextFreePort()
	if err != nil {
		return 0, err
	}

	lst, err := net.Listen("tcp", fmt.Sprintf("%s:%d", bindHost, port))
	if err != nil {
		return 0, err
	}

	go func() {
		defer lst.Close()

		conn, err := lst.Accept()
		if err != nil {
			log.Warningf("Failed to accept connection on %d: %v", port, err)
			return
		}

		defer conn.Close()
		copyFn(conn)
	}()

	return port, nil
}
