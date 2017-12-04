package ipfs

import (
	"errors"
	"fmt"
	"os"
	"sync"

	log "github.com/Sirupsen/logrus"

	core "github.com/ipfs/go-ipfs/core"
	fsrepo "github.com/ipfs/go-ipfs/repo/fsrepo"

	"golang.org/x/net/context"
)

var (
	// ErrIsOffline is returned when an online operation was done offline.
	ErrIsOffline = errors.New("Node is offline")
)

// Node remembers the settings needed for accessing the ipfs daemon.
type Node struct {
	Path      string
	SwarmPort int

	mu sync.Mutex

	ipfsNode *core.IpfsNode

	// Root context used for all operations.
	ctx    context.Context
	cancel context.CancelFunc
}

func createNode(path string, swarmPort int, ctx context.Context, online bool) (*core.IpfsNode, error) {

	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Infof("Creating new ipfs repo at %s since it does not exist yet.", path)
		if err := Init(path, 2048); err != nil {
			return nil, err
		}
	}

	rp, err := fsrepo.Open(path)
	if err != nil {
		log.Errorf("Unable to open repo `%s`: %v", path, err)
		return nil, err
	}

	swarmAddrs := []string{
		fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", swarmPort),
		fmt.Sprintf("/ip6/::/tcp/%d", swarmPort),
	}

	if err := rp.SetConfigKey("Addresses.Swarm", swarmAddrs); err != nil {
		return nil, err
	}

	cfg := &core.BuildCfg{
		Repo:   rp,
		Online: online,
	}

	ipfsNode, err := core.NewNode(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return ipfsNode, nil
}

// New creates a new ipfs node manager.
// No daemon is started yet.
func New(ipfsPath string) (*Node, error) {
	return NewWithPort(ipfsPath, 4001)
}

func NewWithPort(ipfsPath string, swarmPort int) (*Node, error) {
	ctx, cancel := context.WithCancel(context.Background())
	ipfsNode, err := createNode(ipfsPath, swarmPort, ctx, true)
	if err != nil {
		return nil, err
	}

	return &Node{
		Path:      ipfsPath,
		SwarmPort: swarmPort,
		ipfsNode:  ipfsNode,
		ctx:       ctx,
		cancel:    cancel,
	}, nil
}

func (nd *Node) IsOnline() bool {
	nd.mu.Lock()
	defer nd.mu.Unlock()

	return nd.isOnline()
}

func (nd *Node) isOnline() bool {
	return nd.ipfsNode.OnlineMode()
}

func (nd *Node) Online() error {
	nd.mu.Lock()
	defer nd.mu.Unlock()

	if nd.isOnline() {
		return nil
	}

	var err error
	nd.ipfsNode, err = createNode(nd.Path, nd.SwarmPort, nd.ctx, true)
	if err != nil {
		return err
	}

	return nil
}

func (nd *Node) Offline() error {
	nd.mu.Lock()
	defer nd.mu.Unlock()

	if !nd.isOnline() {
		return ErrIsOffline
	}

	var err error
	nd.ipfsNode, err = createNode(nd.Path, nd.SwarmPort, nd.ctx, false)
	if err != nil {
		return err
	}

	return nil
}

// Close shuts down the ipfs node.
// It may not be used afterwards.
func (nd *Node) Close() error {
	nd.cancel()
	return nd.ipfsNode.Close()
}

func (nd *Node) Name() string {
	return "ipfs"
}
