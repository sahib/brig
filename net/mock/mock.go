package mock

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	e "github.com/pkg/errors"
	"github.com/sahib/brig/net/backend"
	"github.com/sahib/brig/net/peer"
)

// NetBackend provides a testing backend implementation
// of net.Backend. It only works on a single machine
// by storing some data into a temporary directory.
type NetBackend struct {
	isOnline bool
	conns    map[string]chan net.Conn
	path     string
	name     string
	port     int
}

// NewNetBackend returns a new fake NetBackend
func NewNetBackend(path, name string, port int) *NetBackend {
	return &NetBackend{
		isOnline: true,
		conns:    make(map[string]chan net.Conn),
		name:     name,
		port:     port,
		path:     path,
	}
}

// PublishName is a fake implementation.
func (nb *NetBackend) PublishName(partialName string) error {
	discoveryName := filepath.Join(nb.path, "discovery", partialName, nb.name)
	if err := os.MkdirAll(filepath.Dir(discoveryName), 0744); err != nil {
		return err
	}

	return ioutil.WriteFile(discoveryName, nil, 0644)
}

// ResolveName is a fake implementation.
func (nb *NetBackend) ResolveName(ctx context.Context, partialName string) ([]peer.Info, error) {
	discoDir := filepath.Join(nb.path, "discovery", partialName)
	names, err := ioutil.ReadDir(discoDir)
	if err != nil {
		return nil, err
	}

	if len(names) == 0 {
		return nil, fmt.Errorf("no such peer: %v", partialName)
	}

	infos := []peer.Info{}
	for _, name := range names {
		dnsName := filepath.Join(nb.path, "dns", filepath.Base(name.Name()))
		data, err := ioutil.ReadFile(dnsName) // #nosec
		if err != nil {
			return nil, fmt.Errorf("no such peer: %v", name)
		}

		infos = append(infos, peer.Info{
			Addr: string(data),
			Name: peer.Name(name.Name()),
		})
	}

	return infos, nil
}

// Connect is a fake implementation.
func (nb *NetBackend) Connect() error {
	if nb.isOnline {
		return fmt.Errorf("already online")
	}

	dnsName := filepath.Join(nb.path, "dns", nb.name)
	if err := os.MkdirAll(filepath.Dir(dnsName), 0744); err != nil {
		return err
	}

	dnsTag := fmt.Sprintf("%s@%d", nb.name, nb.port)

	if err := ioutil.WriteFile(dnsName, []byte(dnsTag), 0644); err != nil {
		return e.Wrap(err, "failed to write dns tag")
	}

	nb.isOnline = true
	return nil
}

// Disconnect is a fake implementation.
func (nb *NetBackend) Disconnect() error {
	if !nb.isOnline {
		return fmt.Errorf("already offline")
	}

	nb.isOnline = false
	return nil
}

// IsOnline is a fake implementation.
func (nb *NetBackend) IsOnline() bool {
	return nb.isOnline
}

// Identity is a fake implementation.
func (nb *NetBackend) Identity() (peer.Info, error) {
	dnsTag := fmt.Sprintf("%s@%d", nb.name, nb.port)
	return peer.Info{
		Addr: dnsTag,
		Name: peer.Name(nb.name),
	}, nil
}

func getPortFromAddr(peerAddr string) (int, error) {
	split := strings.SplitN(peerAddr, "@", 2)
	if len(split) < 2 {
		return 0, fmt.Errorf("invalid mock addr: %s", peerAddr)
	}

	port, err := strconv.Atoi(split[1])
	if err != nil {
		return 0, fmt.Errorf("invalid mock addr port: %s %v", peerAddr, err)
	}

	return port, nil
}

// Dial is a fake implementation.
func (nb *NetBackend) Dial(peerAddr, protocol string) (net.Conn, error) {
	port, err := getPortFromAddr(peerAddr)
	if err != nil {
		return nil, err
	}

	return net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
}

// Ping is a fake implementation.
func (nb *NetBackend) Ping(addr string) (backend.Pinger, error) {
	return pingerByName(addr)
}

// Listen is a fake implementation.
func (nb *NetBackend) Listen(protocol string) (net.Listener, error) {
	addr := fmt.Sprintf("localhost:%d", nb.port)
	log.Debugf("Mock listening on %s", addr)
	return net.Listen("tcp", addr)
}
