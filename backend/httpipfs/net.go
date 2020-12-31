package httpipfs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	netBackend "github.com/sahib/brig/net/backend"
	"github.com/sahib/brig/util"
	shell "github.com/ipfs/go-ipfs-api"
	log "github.com/sirupsen/logrus"
)

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

// Dial will open a connection to the peer identified by `peerHash`,
// running `protocol` over it.
func (nd *Node) Dial(peerHash, fingerprint, protocol string) (net.Conn, error) {
	if !nd.isOnline() {
		return nil, ErrOffline
	}

	self, err := nd.Identity()
	if err != nil {
		return nil, err
	}

	if self.Addr == peerHash {
		// Special case:
		// When we use the same IPFS daemon for different
		// brig repositiories, we want still to be able to dial
		// other brig instances. Since we cannot dial over ipfs
		// we simply have the port written to /tmp where
		// we can pick it up on Dial()
		addr, err := readLocalAddr(peerHash, fingerprint)
		if err != nil {
			return nil, err
		}

		return net.Dial("tcp", addr)
	}

	protocol = path.Join(protocol, peerHash)

	port := util.FindFreePort()
	addr := fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", port)
	if err := forward(nd.sh, protocol, addr, peerHash); err != nil {
		return nil, err
	}

	tcpAddr := fmt.Sprintf("127.0.0.1:%d", port)
	log.Debugf("dial to »%s« over port %d", peerHash, port)
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
	lst         net.Listener
	protocol    string
	peer        string
	targetAddr  string
	fingerprint string
	sh          *shell.Shell
}

func (lw *listenerWrapper) Accept() (net.Conn, error) {
	conn, err := lw.lst.Accept()
	if err != nil {
		return nil, err
	}

	return &connWrapper{
		Conn:       conn,
		peer:       lw.peer,
		protocol:   lw.protocol,
		targetAddr: lw.targetAddr,
		sh:         lw.sh,
	}, nil
}

func (lw *listenerWrapper) Addr() net.Addr {
	return &addrWrapper{
		protocol: lw.protocol,
		peer:     lw.peer,
	}
}

func (lw *listenerWrapper) Close() error {
	defer lw.lst.Close()
	defer deleteLocalAddr(lw.peer, lw.fingerprint)
	return closeStream(lw.sh, lw.protocol, lw.targetAddr, "")
}

func buildLocalAddrPath(id, fingerprint string) string {
	return filepath.Join(os.TempDir(), fmt.Sprintf("brig-%s:%s.addr", id, fingerprint))
}

func readLocalAddr(id, fingerprint string) (string, error) {
	path := buildLocalAddrPath(id, fingerprint)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func deleteLocalAddr(id, fingerprint string) error {
	path := buildLocalAddrPath(id, fingerprint)
	return os.RemoveAll(path)
}

func writeLocalAddr(id, fingerprint, addr string) error {
	path := buildLocalAddrPath(id, fingerprint)
	return ioutil.WriteFile(path, []byte(addr), 0644)
}

// Listen will listen to the protocol
func (nd *Node) Listen(protocol string) (net.Listener, error) {
	if !nd.isOnline() {
		return nil, ErrOffline
	}

	self, err := nd.Identity()
	if err != nil {
		return nil, err
	}

	// TODO: Is this even needed still?
	// Do we want support for having more than one brig per ipfs.
	// Append the id to the protocol:
	protocol = path.Join(protocol, self.Addr)

	port := util.FindFreePort()
	addr := fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", port)

	// Prevent errors by closing any previously opened listeners:
	if err := closeStream(nd.sh, protocol, "", ""); err != nil {
		return nil, err
	}

	log.Debugf("backend: listening for %s over port %d", protocol, port)
	if err := openListener(nd.sh, protocol, addr); err != nil {
		return nil, err
	}

	localAddr := fmt.Sprintf("127.0.0.1:%d", port)
	lst, err := net.Listen("tcp", localAddr)
	if err != nil {
		return nil, err
	}

	if err := writeLocalAddr(self.Addr, nd.fingerprint, localAddr); err != nil {
		return nil, err
	}

	return &listenerWrapper{
		lst:         lst,
		protocol:    protocol,
		peer:        self.Addr,
		targetAddr:  addr,
		fingerprint: nd.fingerprint,
		sh:          nd.sh,
	}, nil
}

/////////////////////////////////

type pinger struct {
	lastSeen  time.Time
	roundtrip time.Duration
	err       error

	mu     sync.Mutex
	cancel func()
	nd     *Node
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
	if p.cancel != nil {
		p.cancel()
		p.cancel = nil
	}

	return nil
}

func (p *pinger) update(ctx context.Context, addr, self string) {
	// Edge case: test setups where we ping ourselves.
	if self == addr {
		p.mu.Lock()
		p.err = nil
		p.lastSeen = time.Now()
		p.roundtrip = time.Duration(0)
		p.mu.Unlock()
		return
	}

	// Do the network op without a lock:
	roundtrip, err := ping(p.nd.sh, addr)

	p.mu.Lock()
	if err != nil {
		p.err = err
	} else {
		p.err = nil
		p.lastSeen = time.Now()
		p.roundtrip = roundtrip
	}

	p.mu.Unlock()
}

func (p *pinger) Run(ctx context.Context, addr string) error {
	self, err := p.nd.Identity()
	if err != nil {
		return err
	}

	p.update(ctx, addr, self.Addr)
	tckr := time.NewTicker(10 * time.Second)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-tckr.C:
			p.update(ctx, addr, self.Addr)
		}
	}
}

func ping(sh *shell.Shell, peerID string) (time.Duration, error) {
	ctx := context.Background()
	resp, err := sh.Request("ping", peerID).Send(ctx)
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

// ErrWaiting is the initial error state of a pinger.
// The error will be unset once a successful ping was made.
var ErrWaiting = errors.New("waiting for route")

// Ping will return a pinger for `addr`.
func (nd *Node) Ping(addr string) (netBackend.Pinger, error) {
	if !nd.isOnline() {
		return nil, ErrOffline
	}

	log.Debugf("backend: start ping »%s«", addr)
	p := &pinger{
		nd:  nd,
		err: ErrWaiting,
	}

	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel
	go p.Run(ctx, addr)
	return p, nil
}
