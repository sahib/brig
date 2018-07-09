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

	"github.com/sahib/brig/net/backend"
	"github.com/sahib/brig/net/peer"

	e "github.com/pkg/errors"
)

const NetMockRoot = "/tmp/local-mock-ipfs"

type NetBackend struct {
	isOnline bool
	conns    map[string]chan net.Conn
	name     string
	port     int
}

func NewNetBackend(name string, port int) (*NetBackend, error) {
	dnsName := filepath.Join(NetMockRoot, "dns", name)
	if err := os.MkdirAll(filepath.Dir(dnsName), 0744); err != nil {
		return nil, err
	}

	dnsTag := fmt.Sprintf("%s@%d", name, port)

	if err := ioutil.WriteFile(dnsName, []byte(dnsTag), 0644); err != nil {
		return nil, e.Wrap(err, "failed to write dns tag")
	}

	return &NetBackend{
		isOnline: true,
		conns:    make(map[string]chan net.Conn),
		name:     name,
		port:     port,
	}, nil
}

func (nb *NetBackend) PublishName(partialName string) error {
	discoveryName := filepath.Join(NetMockRoot, "discovery", partialName, nb.name)
	if err := os.MkdirAll(filepath.Dir(discoveryName), 0744); err != nil {
		return err
	}

	return ioutil.WriteFile(discoveryName, nil, 0644)
}

func (nb *NetBackend) ResolveName(ctx context.Context, partialName string) ([]peer.Info, error) {
	discoDir := filepath.Join(NetMockRoot, "discovery", partialName)
	names, err := ioutil.ReadDir(discoDir)
	if err != nil {
		return nil, err
	}

	if len(names) == 0 {
		return nil, fmt.Errorf("no such peer: %v", partialName)
	}

	infos := []peer.Info{}
	for _, name := range names {
		dnsName := filepath.Join(NetMockRoot, "dns", filepath.Base(name.Name()))
		data, err := ioutil.ReadFile(dnsName)
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

func (nb *NetBackend) Connect() error {
	if nb.isOnline {
		return fmt.Errorf("already online")
	}

	nb.isOnline = true
	return nil
}

func (nb *NetBackend) Disconnect() error {
	if !nb.isOnline {
		return fmt.Errorf("already offline")
	}

	nb.isOnline = false
	return nil
}

func (nb *NetBackend) IsOnline() bool {
	return nb.isOnline
}

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

func (nb *NetBackend) Dial(peerAddr, protocol string) (net.Conn, error) {
	port, err := getPortFromAddr(peerAddr)
	if err != nil {
		return nil, err
	}

	return net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
}

func (nb *NetBackend) Ping(addr string) (backend.Pinger, error) {
	return pingerByName(addr)
}

func (nb *NetBackend) Listen(protocol string) (net.Listener, error) {
	return net.Listen("tcp", fmt.Sprintf("localhost:%d", nb.port))
}
