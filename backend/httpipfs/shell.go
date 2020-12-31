package httpipfs

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/blang/semver"
	"github.com/patrickmn/go-cache"
	"github.com/sahib/brig/repo/setup"
	shell "github.com/ipfs/go-ipfs-api"
	log "github.com/sirupsen/logrus"
)

var (
	// ErrOffline is returned by operations that need online support
	// to work when the backend is in offline mode.
	ErrOffline = errors.New("backend is in offline mode")
)

// IpfsStateCache contains various backend related caches
type IpfsStateCache struct {
	locallyCached *cache.Cache // shows if the hash and its children is locally cached by ipfs
}

// Node is the struct that holds the httpipfs backend together.
// It is a shallow type that has not much own state and is very light.
type Node struct {
	sh             *shell.Shell
	mu             sync.Mutex
	cachedIdentity string
	allowNetOps    bool
	fingerprint    string
	version        *semver.Version
	cache          *IpfsStateCache
}

func getExperimentalFeatures(sh *shell.Shell) (map[string]bool, error) {
	ctx := context.Background()
	resp, err := sh.Request("config/show").Send(ctx)
	if err != nil {
		return nil, err
	}

	defer resp.Close()

	if resp.Error != nil {
		return nil, resp.Error
	}

	raw := struct {
		Experimental map[string]bool
	}{}

	if err := json.NewDecoder(resp.Output).Decode(&raw); err != nil {
		return nil, err
	}

	return raw.Experimental, nil
}

// NewNode returns a new http based IPFS backend.
func NewNode(ipfsPath, fingerprint string) (*Node, error) {
	addr, err := setup.GetAPIAddrForPath(ipfsPath)
	if err != nil {
		return nil, err
	}

	log.Infof("Connecting to IPFS HTTP API at %s", addr)
	sh := shell.NewShell(addr)

	versionString, _, err := sh.Version()
	if err != nil {
		log.Warningf("failed to get version: %v", err)
	}

	version, err := semver.Parse(versionString)
	if err != nil {
		log.Warningf("failed to parse version string of IPFS (»%s«): %v", versionString, err)
	}

	log.Infof("The IPFS version is »%s«.", version)
	if version.LT(semver.MustParse("0.4.18")) {
		log.Warningf("This version is quite old. Please update, if possible.\n")
		log.Warningf("We only test on newer versions (>= 0.4.18).\n")
	}

	features, err := getExperimentalFeatures(sh)
	if err != nil {
		log.Warningf("Failed to get experimental feature list: %v", err)
	} else {
		if !features["Libp2pStreamMounting"] {
			log.Warningf("Stream mounting does not seem to be enabled.")
			log.Warningf("Please execute the following to change that:")
			log.Warningf("$ ipfs config --json Experimental.Libp2pStreamMounting true")
		}
	}

	return &Node{
		sh:          sh,
		allowNetOps: true,
		fingerprint: fingerprint,
		version:     &version,
		cache: &IpfsStateCache{
			locallyCached: cache.New(5*time.Minute, 10*time.Minute),
		},
	}, nil
}

// IsOnline returns true if the node is in online mode and the daemon is reachable.
func (nd *Node) IsOnline() bool {
	nd.mu.Lock()
	allowNetOps := nd.allowNetOps
	nd.mu.Unlock()

	return nd.sh.IsUp() && allowNetOps
}

// Connect implements Backend.Connect
func (nd *Node) Connect() error {
	nd.mu.Lock()
	defer nd.mu.Unlock()

	nd.allowNetOps = true
	return nil
}

// Disconnect implements Backend.Disconnect
func (nd *Node) Disconnect() error {
	nd.mu.Lock()
	defer nd.mu.Unlock()

	nd.allowNetOps = false
	return nil
}

func (nd *Node) isOnline() bool {
	nd.mu.Lock()
	defer nd.mu.Unlock()

	return nd.allowNetOps
}

// Close implements Backend.Close
func (nd *Node) Close() error {
	return nil
}

// Name returns "httpipfs" as name of the backend.
func (nd *Node) Name() string {
	return "httpipfs"
}
