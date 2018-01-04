package mock

import (
	"fmt"
	"net"
	"time"

	"github.com/sahib/brig/net/backend"
	"github.com/sahib/brig/net/peer"
)

type NetBackend struct {
	isOnline bool
	conns    map[string]chan net.Conn
}

func NewNetBackend() *NetBackend {
	return &NetBackend{
		isOnline: true,
		conns:    make(map[string]chan net.Conn),
	}
}

func (nb *NetBackend) ResolveName(name string) ([]peer.Info, error) {
	switch name {
	case "bob":
		return []peer.Info{
			{Name: peer.Name(name), Addr: "bob-addr"},
		}, nil
	case "charlie":
		return []peer.Info{
			{Name: peer.Name(name), Addr: "charlie-addr-right"},
			{Name: peer.Name(name), Addr: "charlie-addr-wrong"},
		}, nil
	case "vincent":
		// Vincent is always offline.
		//
		// This is a work of fiction. Names, characters, places and incidents
		// either are products of the author’s imagination or are used
		// fictitiously. Any resemblance to actual events or locales or persons,
		// living or dead, is entirely coincidental.
		return []peer.Info{
			{Name: peer.Name(name), Addr: "vincent-addr"},
		}, nil
	case "mallory":
		// Mallory is a faker:
		return []peer.Info{
			{Name: peer.Name(name), Addr: "charlie-addr-right"},
			{Name: peer.Name(name), Addr: "bob-addr"},
		}, nil
	default:
		return nil, fmt.Errorf("No such peer: %v", name)
	}
}

func (nb *NetBackend) PublishName(name string) error {
	return nil
}

func (nb *NetBackend) Connect() error {
	if nb.isOnline {
		return fmt.Errorf("Already online")
	}

	nb.isOnline = true
	return nil
}

func (nb *NetBackend) Disconnect() error {
	if !nb.isOnline {
		return fmt.Errorf("Already offline")
	}

	nb.isOnline = false
	return nil
}

func (nb *NetBackend) IsOnline() bool {
	return nb.isOnline
}

func (nb *NetBackend) Identity() (peer.Info, error) {
	return peer.Info{
		Addr: "alice-addr",
		Name: peer.Name("alice"),
	}, nil
}

func (nb *NetBackend) Dial(peerAddr, protocol string) (net.Conn, error) {
	switch peerAddr {
	case "alice-addr":
		return nil, fmt.Errorf("Cannot dial self")
	case "vincent-addr":
		return nil, fmt.Errorf("vincent is offline")
	case "bob-addr", "charlie-addr-right":
		// Those are the only valid addrs we may dial.
		break
	case "charlie-addr-wrong":
		return nil, fmt.Errorf("No such peer")
	}

	// We basically call ourselves with the mock backend,
	// just pretending to be a different peer.
	clConn, srvConn, err := LoopPipe()
	if err != nil {
		return nil, err
	}

	ch, ok := nb.conns[protocol]
	if !ok {
		return nil, fmt.Errorf("No listener for this protocol (offline?): %v", protocol)
	}

	ch <- srvConn
	return clConn, nil
}

func (nb *NetBackend) Ping(addr string) (backend.Pinger, error) {
	return pingerByName(addr)
}

func (nb *NetBackend) Listen(protocol string) (net.Listener, error) {
	return &memListener{
		nb:       nb,
		protocol: protocol,
	}, nil
}

type memListener struct {
	nb          *NetBackend
	hasDeadline bool
	deadline    time.Time
	protocol    string
}

type timeoutError struct{}

func (te *timeoutError) Timeout() bool {
	return true
}

func (te *timeoutError) Error() string {
	return "timeout"
}

func (ml *memListener) Accept() (net.Conn, error) {
	ch, ok := ml.nb.conns[ml.protocol]
	if !ok {
		ch = make(chan net.Conn, 1)
		ml.nb.conns[ml.protocol] = ch
	}

	timeoutCh := make(<-chan time.Time)
	if ml.hasDeadline {
		timeoutCh = time.After(ml.deadline.Sub(time.Now()))
	}

	select {
	case <-timeoutCh:
		return nil, &timeoutError{}
	case conn := <-ch:
		return conn, nil
	}
}

func (ml *memListener) SetDeadline(t time.Time) error {
	ml.deadline = t
	ml.hasDeadline = true
	return nil
}

type memAddr string

func (ma memAddr) Network() string {
	return "mock"
}

func (ma memAddr) String() string {
	return string(ma)
}

func (ml *memListener) Addr() net.Addr {
	return memAddr(ml.protocol)
}

func (ml *memListener) Close() error {
	delete(ml.nb.conns, ml.protocol)
	return nil
}
