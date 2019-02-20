package httpipfs

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	shell "github.com/ipfs/go-ipfs-api"
	netBackend "github.com/sahib/brig/net/backend"
)

// TODO: Move this to util.
func findFreePort() int {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0
	}

	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port
}

//////////////////////////

type connWrapper struct {
	net.Conn

	peer       string
	protocol   string
	targetAddr string
	sh         *shell.Shell
}

func (cw *connWrapper) LocalAddr() net.Addr {
	return &addrWrapper{
		protocol: cw.protocol,
		peer:     "",
	}
}

func (cw *connWrapper) RemoteAddr() net.Addr {
	return &addrWrapper{
		protocol: cw.protocol,
		peer:     cw.peer,
	}
}

func (cw *connWrapper) Close() error {
	defer cw.Conn.Close()
	return closeStream(cw.sh, cw.protocol, "", cw.targetAddr)
}

func (nd *Node) Dial(peerHash, protocol string) (net.Conn, error) {
	port := findFreePort()
	addr := fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", port)
	if err := forward(nd.sh, protocol, addr, peerHash); err != nil {
		return nil, err
	}

	tcpAddr := fmt.Sprintf("127.0.0.1:%d", port)
	conn, err := net.Dial("tcp", tcpAddr)
	if err != nil {
		return nil, err
	}

	return &connWrapper{
		Conn:       conn,
		peer:       peerHash,
		protocol:   protocol,
		targetAddr: addr,
		sh:         nd.sh,
	}, nil
}

//////////////////////////

func forward(sh *shell.Shell, protocol, targetAddr, peerID string) error {
	ctx := context.Background()
	peerID = "/ipfs/" + peerID

	rb := sh.Request("p2p/forward", protocol, targetAddr, peerID)
	rb.Option("allow-custom-protocol", true)
	resp, err := rb.Send(ctx)
	if err != nil {
		return err
	}

	defer resp.Close()
	if resp.Error != nil {
		return resp.Error
	}

	return nil
}

func openListener(sh *shell.Shell, protocol, targetAddr string) error {
	ctx := context.Background()
	rb := sh.Request("p2p/listen", protocol, targetAddr)
	rb.Option("allow-custom-protocol", true)
	resp, err := rb.Send(ctx)
	if err != nil {
		return err
	}

	defer resp.Close()
	if err := resp.Error; err != nil {
		return err
	}

	return nil
}

func closeStream(sh *shell.Shell, protocol, targetAddr, listenAddr string) error {
	ctx := context.Background()
	rb := sh.Request("p2p/close")
	rb.Option("protocol", protocol)

	if targetAddr != "" {
		rb.Option("target-address", targetAddr)
	}

	if listenAddr != "" {
		rb.Option("listen-address", listenAddr)
	}

	resp, err := rb.Send(ctx)
	if err != nil {
		return err
	}

	defer resp.Close()
	if resp.Error != nil {
		return resp.Error
	}

	return nil
}

type addrWrapper struct {
	protocol string
	peer     string
}

func (sa *addrWrapper) Network() string {
	return sa.protocol
}

func (sa *addrWrapper) String() string {
	return sa.peer
}

type listenerWrapper struct {
	net.Listener
	protocol   string
	peer       string
	targetAddr string
	sh         *shell.Shell
}

func (lw *listenerWrapper) Addr() net.Addr {
	return &addrWrapper{
		protocol: lw.protocol,
		peer:     lw.peer,
	}
}

func (lw *listenerWrapper) Close() error {
	defer lw.Listener.Close()
	return closeStream(lw.sh, lw.protocol, lw.targetAddr, "")
}

func (nd *Node) Listen(protocol string) (net.Listener, error) {
	self, err := nd.Identity()
	if err != nil {
		return nil, err
	}

	port := findFreePort()
	addr := fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", port)
	if err := openListener(nd.sh, protocol, addr); err != nil {
		return nil, err
	}

	localAddr := fmt.Sprintf("127.0.0.1:%d", port)
	lst, err := net.Listen("tcp", localAddr)
	if err != nil {
		return nil, err
	}

	return &listenerWrapper{
		Listener:   lst,
		protocol:   protocol,
		peer:       self.Addr,
		targetAddr: addr,
		sh:         nd.sh,
	}, nil
}

/////////////////////////////////

type pinger struct {
	lastSeen  time.Time
	roundtrip time.Duration
	err       error

	mu     sync.Mutex
	cancel func()
	sh     *shell.Shell
}

// LastSeen returns the time we pinged the remote last time.
func (p *pinger) LastSeen() time.Time {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.lastSeen
}

// Roundtrip returns the time needed send a single package to
// the remote and receive the answer.
func (p *pinger) Roundtrip() time.Duration {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.roundtrip
}

// Err will return a non-nil error when the current ping did not succeed.
func (p *pinger) Err() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.err
}

// Close will clean up the pinger.
func (p *pinger) Close() error {
	p.cancel()
	return nil
}

func (p *pinger) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel

	for {
		select {
		case <-ctx.Done():
			break
		default:
		}

		p.mu.Lock()
		roundtrip, err := ping(p.sh)
		if err != nil {
			p.err = err
		} else {
			p.lastSeen = time.Now()
			p.roundtrip = roundtrip
		}

		p.mu.Unlock()
	}
}

// TODO: Make a PR with those functions.
func ping(sh *shell.Shell) (time.Duration, error) {
	ctx := context.Background()
	resp, err := sh.Request("ping").Send(ctx)
	if err != nil {
		return 0, err
	}

	defer resp.Close()

	if resp.Error != nil {
		return 0, resp.Error
	}

	raw := struct {
		Success bool
		Time    int64
	}{}

	if err := json.NewDecoder(resp.Output).Decode(&raw); err != nil {
		return 0, err
	}

	if raw.Success {
		return time.Duration(raw.Time), nil
	}

	return 0, fmt.Errorf("no ping")
}

func (nd *Node) Ping(addr string) (netBackend.Pinger, error) {
	p := &pinger{sh: nd.sh}
	go p.Run()
	return p, nil
}
