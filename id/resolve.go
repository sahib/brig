package id

import (
	"net"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/util/ipfsutil"
	ma "github.com/jbenet/go-multiaddr"
	"golang.org/x/net/context"
)

type Peer interface {
	ID() ID
	Hash() string
}

type ipfsPeer struct {
	id   ID
	hash string
}

func NewPeer(id ID, hash string) Peer {
	return &ipfsPeer{id, hash}
}

func (ip *ipfsPeer) ID() ID {
	return ip.id
}

func (ip *ipfsPeer) Hash() string {
	return ip.hash
}

type Addresses []net.Addr

// func (addrs Addresses) Local() ma.Multiaddr {
// 	// TODO: This is so stupid, it hurts.
// 	for _, addr := range addrs {
// 		if strings.Contains(addr.String(), "192.") {
// 			return addr
// 		}
// 	}
//
// 	return nil
// }

type Resolver interface {
	Resolve(ctx context.Context) (Addresses, error)
	Peer() Peer
}

// IPFSResolver tries to resolve a ID to a IP
// by looking up the corresponding providers and peer IDs
// on the network.
type ipfsResolver struct {
	peerHash string
	id       ID
	node     *ipfsutil.Node
}

func NewIpfsResolver(node *ipfsutil.Node, id ID, peerHash string) Resolver {
	return &ipfsResolver{
		node:     node,
		peerHash: peerHash,
		id:       id,
	}
}

func (ir *ipfsResolver) Peer() Peer {
	return NewPeer(ir.id, ir.peerHash)
}

func (ir *ipfsResolver) Resolve(ctx context.Context) (Addresses, error) {
	subCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	timeout := 10 * time.Second
	if deadline, ok := subCtx.Deadline(); ok {
		timeout = deadline.Sub(time.Now())
	}

	// TODO: `20` is a bit random..
	// Pass ir.peerHash directly to Locate?
	peers, err := ipfsutil.Locate(ir.node, ir.id.Hash(), 20, timeout)
	if err != nil {
		return nil, err
	}

	var maddrs []ma.Multiaddr

	// Select the peer with the desired id:
	for _, peer := range peers {
		if peer.ID == ir.peerHash {
			maddrs = peer.Addrs
			break
		}
	}

	// Make handling errors a bit easier; error out when no results found:
	if maddrs == nil {
		return nil, ErrNoAddrs
	}

	// TODO: This sucks balls.
	addrs := Addresses{}
	for _, maddr := range maddrs {
		ipStr, err := maddr.ValueForProtocol(ma.P_IP4)
		if err != nil {
			log.Warningf("Bad protocol: IP4: %v", err)
			continue
		}

		portStr, err := maddr.ValueForProtocol(ma.P_TCP)
		if err != nil {
			log.Warningf("Bad protocol: TCP: %v", err)
			continue
		}

		port, err := strconv.Atoi(portStr)
		if err != nil {
			log.Warningf("Bad port spec (%s): %v", portStr, err)
			continue
		}

		ip := net.ParseIP(ipStr)
		if ip == nil {
			log.Warningf("Bad IP spec: %s", ipStr)
			continue
		}

		addrs = append(addrs, &net.TCPAddr{IP: ip, Port: port})
	}

	return addrs, nil
}
