package config

import (
	"context"
	"fmt"

	bhost "gx/ipfs/QmZ86eLPtXkQ1Dfa992Q8NpXArUoWWh3y728JDcWvzRrvC/go-libp2p/p2p/host/basic"

	inet "gx/ipfs/QmPjvxTpVH8qJyQDnxnsxF9kv9jezKD1kozz1hs3fCGsNh/go-libp2p-net"
	filter "gx/ipfs/QmSW4uNHbvQia8iZDXzbwjiyHQtnyo9aFqfQAMasj3TJ6Y/go-maddr-filter"
	circuit "gx/ipfs/QmUTvbhkVteiGNkYJ5CN953NtGTPbgctcLk2WU1JVDV6TR/go-libp2p-circuit"
	swarm "gx/ipfs/QmVqCSwuzgDfhLMTmFfUePTGX78PBjzuHcbSWWNPrnrmKy/go-libp2p-swarm"
	pnet "gx/ipfs/QmW7Ump7YyBMr712Ta3iEVh3ZYcfVvJaPryfbCnyE826b4/go-libp2p-interface-pnet"
	ifconnmgr "gx/ipfs/QmXuucFcuvAWYAJfhHV2h4BYreHEAsLSsiquosiXeuduTN/go-libp2p-interface-connmgr"
	ma "gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
	pstore "gx/ipfs/QmZR2XWVVBCtbgBWnQhWk2xcQfaR3W8faQPriAiaaj7rsr/go-libp2p-peerstore"
	host "gx/ipfs/Qmb8T6YBBsjYsVGfrihQLfCJveczZnneSBqBKkYEBWDjge/go-libp2p-host"
	tptu "gx/ipfs/QmbzWHWFpwCm7GDgrGyKnqUzh48PnRp86VD2Bb8wq6ZfDP/go-libp2p-transport-upgrader"
	logging "gx/ipfs/QmcVVHfdyv15GVPk7NrxdWjh2hLVccXnoD8j2tyQShiXJb/go-log"
	metrics "gx/ipfs/QmcoBbyTiL9PFjo1GFixJwqQ8mZLJ36CribuqyKmS1okPu/go-libp2p-metrics"
	peer "gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
	crypto "gx/ipfs/Qme1knMqwt1hKZbc1BmQFmnm9f36nyQGwXxPGVpVJ9rMK5/go-libp2p-crypto"
)

var log = logging.Logger("p2p-config")

// AddrsFactory is a function that takes a set of multiaddrs we're listening on and
// returns the set of multiaddrs we should advertise to the network.
type AddrsFactory = bhost.AddrsFactory

// NATManagerC is a NATManager constructor.
type NATManagerC func(inet.Network) bhost.NATManager

// Config describes a set of settings for a libp2p node
//
// This is *not* a stable interface. Use the options defined in the root
// package.
type Config struct {
	PeerKey crypto.PrivKey

	Transports         []TptC
	Muxers             []MsMuxC
	SecurityTransports []MsSecC
	Insecure           bool
	Protector          pnet.Protector

	Relay     bool
	RelayOpts []circuit.RelayOpt

	ListenAddrs  []ma.Multiaddr
	AddrsFactory bhost.AddrsFactory
	Filters      *filter.Filters

	ConnManager ifconnmgr.ConnManager
	NATManager  NATManagerC
	Peerstore   pstore.Peerstore
	Reporter    metrics.Reporter
}

// NewNode constructs a new libp2p Host from the Config.
//
// This function consumes the config. Do not reuse it (really!).
func (cfg *Config) NewNode(ctx context.Context) (host.Host, error) {
	// Check this early. Prevents us from even *starting* without verifying this.
	if pnet.ForcePrivateNetwork && cfg.Protector == nil {
		log.Error("tried to create a libp2p node with no Private" +
			" Network Protector but usage of Private Networks" +
			" is forced by the enviroment")
		// Note: This is *also* checked the upgrader itself so it'll be
		// enforced even *if* you don't use the libp2p constructor.
		return nil, pnet.ErrNotInPrivateNetwork
	}

	if cfg.PeerKey == nil {
		return nil, fmt.Errorf("no peer key specified")
	}

	// Obtain Peer ID from public key
	pid, err := peer.IDFromPublicKey(cfg.PeerKey.GetPublic())
	if err != nil {
		return nil, err
	}

	if cfg.Peerstore == nil {
		return nil, fmt.Errorf("no peerstore specified")
	}

	if !cfg.Insecure {
		cfg.Peerstore.AddPrivKey(pid, cfg.PeerKey)
		cfg.Peerstore.AddPubKey(pid, cfg.PeerKey.GetPublic())
	}

	// TODO: Make the swarm implementation configurable.
	swrm := swarm.NewSwarm(ctx, pid, cfg.Peerstore, cfg.Reporter)
	if cfg.Filters != nil {
		swrm.Filters = cfg.Filters
	}

	// TODO: make host implementation configurable.
	h, err := bhost.NewHost(ctx, swrm, &bhost.HostOpts{
		ConnManager:  cfg.ConnManager,
		AddrsFactory: cfg.AddrsFactory,
		NATManager:   cfg.NATManager,
	})
	if err != nil {
		swrm.Close()
		return nil, err
	}

	upgrader := new(tptu.Upgrader)
	upgrader.Protector = cfg.Protector
	upgrader.Filters = swrm.Filters
	if cfg.Insecure {
		upgrader.Secure = makeInsecureTransport(pid)
	} else {
		upgrader.Secure, err = makeSecurityTransport(h, cfg.SecurityTransports)
		if err != nil {
			h.Close()
			return nil, err
		}
	}

	upgrader.Muxer, err = makeMuxer(h, cfg.Muxers)
	if err != nil {
		h.Close()
		return nil, err
	}

	tpts, err := makeTransports(h, upgrader, cfg.Transports)
	if err != nil {
		h.Close()
		return nil, err
	}
	for _, t := range tpts {
		err = swrm.AddTransport(t)
		if err != nil {
			h.Close()
			return nil, err
		}
	}

	if cfg.Relay {
		err := circuit.AddRelayTransport(swrm.Context(), h, upgrader, cfg.RelayOpts...)
		if err != nil {
			h.Close()
			return nil, err
		}
	}

	// TODO: This method succeeds if listening on one address succeeds. We
	// should probably fail if listening on *any* addr fails.
	if err := h.Network().Listen(cfg.ListenAddrs...); err != nil {
		h.Close()
		return nil, err
	}

	// TODO: Configure routing (it's a pain to setup).
	// TODO: Bootstrapping.

	return h, nil
}

// Option is a libp2p config option that can be given to the libp2p constructor
// (`libp2p.New`).
type Option func(cfg *Config) error

// Apply applies the given options to the config, returning the first error
// encountered (if any).
func (cfg *Config) Apply(opts ...Option) error {
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return err
		}
	}
	return nil
}
