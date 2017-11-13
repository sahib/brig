package memory

import (
	"fmt"
	"net"

	"github.com/disorganizer/brig/net/peer"
)

type NetBackend struct {
	conns map[string]chan net.Conn
}

func NewNetBackend() *NetBackend {
	return &NetBackend{
		conns: make(map[string]chan net.Conn),
	}
}

func (nb *NetBackend) ResolveName(name peer.Name) ([]peer.Info, error) {
	switch name {
	case "bob":
		return []peer.Info{
			{Name: name, Addr: "bob-addr"},
		}, nil
	case "charlie":
		return []peer.Info{
			{Name: name, Addr: "charlie-addr-right"},
			{Name: name, Addr: "charlie-addr-wrong"},
		}, nil
	case "vincent":
		// Vincent is always offline.
		//
		// This is a work of fiction. Names, characters, places and incidents
		// either are products of the author’s imagination or are used
		// fictitiously. Any resemblance to actual events or locales or persons,
		// living or dead, is entirely coincidental.
		return []peer.Info{
			{Name: name, Addr: "vincent-addr"},
		}, nil
	case "mallory":
		// Mallory is a faker:
		return []peer.Info{
			{Name: name, Addr: "charlie-addr-right"},
			{Name: name, Addr: "bob-addr"},
		}, nil
	default:
		return nil, fmt.Errorf("No such peer: %v", name)
	}
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
	case "vincent":
		return nil, fmt.Errorf("vincent is offline")
	case "bob-addr", "charlie-addr-right":
		// Those are the only valid addrs we may dial.
		break
	case "charlie-addr-wrong":
		return nil, fmt.Errorf("No such peer")
	}

	clConn, srvConn := net.Pipe()
	ch, ok := nb.conns[protocol]
	if !ok {
		return nil, fmt.Errorf("No listener for this protocol: %v", protocol)
	}

	ch <- srvConn
	return clConn, nil
}

func (nb *NetBackend) Listen(protocol string) (net.Listener, error) {
	return &memListener{
		nb:       nb,
		protocol: protocol,
	}, nil
}

type memListener struct {
	nb       *NetBackend
	protocol string
}

func (ml *memListener) Accept() (net.Conn, error) {
	ch, ok := ml.nb.conns[ml.protocol]
	if !ok {
		ch = make(chan net.Conn, 1)
	}

	return <-ch, nil
}

type memAddr string

func (ma memAddr) Network() string {
	return "mem"
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
