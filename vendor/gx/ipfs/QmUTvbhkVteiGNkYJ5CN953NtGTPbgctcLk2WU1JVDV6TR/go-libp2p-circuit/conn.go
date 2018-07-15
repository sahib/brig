package relay

import (
	"fmt"
	"net"

	inet "gx/ipfs/QmPjvxTpVH8qJyQDnxnsxF9kv9jezKD1kozz1hs3fCGsNh/go-libp2p-net"
	manet "gx/ipfs/QmV6FjemM1K8oXjrvuq3wuVWWoU2TLDPmNnKrxHzY3v6Ai/go-multiaddr-net"
	ma "gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
	pstore "gx/ipfs/QmZR2XWVVBCtbgBWnQhWk2xcQfaR3W8faQPriAiaaj7rsr/go-libp2p-peerstore"
)

type Conn struct {
	inet.Stream
	remote pstore.PeerInfo
}

type NetAddr struct {
	Relay  string
	Remote string
}

func (n *NetAddr) Network() string {
	return "libp2p-circuit-relay"
}

func (n *NetAddr) String() string {
	return fmt.Sprintf("relay[%s-%s]", n.Remote, n.Relay)
}

func (c *Conn) RemoteAddr() net.Addr {
	return &NetAddr{
		Relay:  c.Conn().RemotePeer().Pretty(),
		Remote: c.remote.ID.Pretty(),
	}
}

func (c *Conn) RemoteMultiaddr() ma.Multiaddr {
	a, err := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s/p2p-circuit/ipfs/%s", c.Conn().RemotePeer().Pretty(), c.remote.ID.Pretty()))
	if err != nil {
		panic(err)
	}
	return a
}

func (c *Conn) LocalMultiaddr() ma.Multiaddr {
	return c.Conn().LocalMultiaddr()
}

func (c *Conn) LocalAddr() net.Addr {
	na, err := manet.ToNetAddr(c.Conn().LocalMultiaddr())
	if err != nil {
		log.Error("failed to convert local multiaddr to net addr:", err)
		return nil
	}
	return na
}
