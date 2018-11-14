package ipfs

import (
	"errors"
	"fmt"
	"net"
	"os"
	"sync"

	log "github.com/Sirupsen/logrus"
	e "github.com/pkg/errors"

	core "github.com/ipfs/go-ipfs/core"
	coreapi "github.com/ipfs/go-ipfs/core/coreapi"
	coreiface "github.com/ipfs/go-ipfs/core/coreapi/interface"
	fsrepo "github.com/ipfs/go-ipfs/repo/fsrepo"
	migrate "github.com/ipfs/go-ipfs/repo/fsrepo/migrations"

	"golang.org/x/net/context"
)

// Find the next free tcp port near to `port` (possibly equal to `port`).
// Only `maxTries` number of trials will be made.
// This method is (of course...) racy since the port might be already
// taken again by another process until we startup our service on that port.
func findFreePortAfter(port int, maxTries int) int {
	for idx := 0; idx < maxTries; idx++ {
		addr := fmt.Sprintf("localhost:%d", port+idx)
		lst, err := net.Listen("tcp", addr)
		if err != nil {
			continue
		}

		if err := lst.Close(); err != nil {
			// continue, this port might be burned.
			// should not happen most likely though.
			continue
		}

		return port + idx
	}

	return port
}

var (
	// ErrIsOffline is returned when an online operation was done offline.
	ErrIsOffline = errors.New("Node is offline")
)

// Node remembers the settings needed for accessing the ipfs daemon.
type Node struct {
	Path      string
	SwarmPort int

	mu       sync.Mutex
	ipfsNode *core.IpfsNode

	// Root context used for all operations.
	ctx            context.Context
	cancel         context.CancelFunc
	bootstrapAddrs []string
	api            coreiface.CoreAPI
}

func createNode(path string, minSwarmPort int, ctx context.Context, online bool, bootstrapAddrs []string) (*core.IpfsNode, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Infof("Creating new ipfs repo at %s since it does not exist yet.", path)
		if err := Init(path, 2048); err != nil {
			return nil, err
		}
	}

	rp, err := fsrepo.Open(path)
	if err == fsrepo.ErrNeedMigration {
		log.Infof("the ipfs repo version changed. We need to run a migration now.")
		if err := migrate.RunMigration(fsrepo.RepoVersion); err != nil {
			log.Errorf("migration failed: %v", err)
			return nil, e.Wrapf(err, "migration failed")
		}

		// Try re-opening it:
		rp, err = fsrepo.Open(path)
		if err != nil {
			return nil, e.Wrapf(err, "failed to open repo after migration")
		}
	}

	if len(bootstrapAddrs) > 0 && rp != nil {
		cfg, err := rp.Config()
		if err != nil {
			return nil, err
		}

		bootstrapMap := make(map[string]struct{})
		for _, entry := range cfg.Bootstrap {
			bootstrapMap[entry] = struct{}{}
		}

		for _, addr := range bootstrapAddrs {
			fullAddr := "/dnsaddr/bootstrap.libp2p.io/ipfs/" + addr
			if _, ok := bootstrapMap[fullAddr]; ok {
				continue
			}

			cfg.Bootstrap = append(cfg.Bootstrap, fullAddr)
		}

		if err := rp.SetConfig(cfg); err != nil {
			return nil, err
		}
	}

	if err != nil {
		log.Errorf("Unable to open repo `%s`: %v", path, err)
		return nil, err
	}

	swarmPort := findFreePortAfter(minSwarmPort, 100)

	log.Debugf(
		"ipfs node configured to run on swarm port %d",
		swarmPort,
	)

	// Resource on the config keys can be found here:
	// https://github.com/ipfs/go-ipfs/blob/master/docs/config.md
	config := map[string]interface{}{
		"Addresses.Swarm": []string{
			fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", swarmPort),
			fmt.Sprintf("/ip6/::/tcp/%d", swarmPort),
		},
		"Addresses.API":        "",
		"Addresses.Gateway":    "",
		"Reprovider.Interval":  "2h",
		"Swarm.EnableRelayHop": true,
		"Experimental.QUIC":    true,
	}

	for key, value := range config {
		if err := rp.SetConfigKey(key, value); err != nil {
			return nil, err
		}
	}

	cfg := &core.BuildCfg{
		Repo:   rp,
		Online: online,
		ExtraOpts: map[string]bool{
			"pubsub": true,
		},
	}

	ipfsNode, err := core.NewNode(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return ipfsNode, nil
}

// New creates a new ipfs node manager. No daemon is started yet.
func New(ipfsPath string, bootstrapAddrs []string) (*Node, error) {
	return NewWithPort(ipfsPath, bootstrapAddrs, 4001)
}

func NewWithPort(ipfsPath string, bootstrapAddrs []string, swarmPort int) (*Node, error) {
	ctx, cancel := context.WithCancel(context.Background())
	ipfsNode, err := createNode(ipfsPath, swarmPort, ctx, true, bootstrapAddrs)
	if err != nil {
		return nil, err
	}

	return &Node{
		Path:           ipfsPath,
		SwarmPort:      swarmPort,
		ipfsNode:       ipfsNode,
		api:            coreapi.NewCoreAPI(ipfsNode),
		ctx:            ctx,
		cancel:         cancel,
		bootstrapAddrs: bootstrapAddrs,
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

func (nd *Node) Connect() error {
	nd.mu.Lock()
	defer nd.mu.Unlock()

	if nd.isOnline() {
		return nil
	}

	var err error
	nd.ipfsNode, err = createNode(nd.Path, nd.SwarmPort, nd.ctx, true, nd.bootstrapAddrs)
	if err != nil {
		return err
	}

	nd.api = coreapi.NewCoreAPI(nd.ipfsNode)
	return nil
}

func (nd *Node) Disconnect() error {
	nd.mu.Lock()
	defer nd.mu.Unlock()

	if !nd.isOnline() {
		return ErrIsOffline
	}

	var err error
	nd.ipfsNode, err = createNode(nd.Path, nd.SwarmPort, nd.ctx, false, nd.bootstrapAddrs)
	if err != nil {
		return err
	}

	nd.api = coreapi.NewCoreAPI(nd.ipfsNode)
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
