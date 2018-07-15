package relay

import (
	"fmt"
	"net"

	pb "gx/ipfs/QmUTvbhkVteiGNkYJ5CN953NtGTPbgctcLk2WU1JVDV6TR/go-libp2p-circuit/pb"

	manet "gx/ipfs/QmV6FjemM1K8oXjrvuq3wuVWWoU2TLDPmNnKrxHzY3v6Ai/go-multiaddr-net"
	ma "gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
)

var _ manet.Listener = (*RelayListener)(nil)

type RelayListener Relay

func (l *RelayListener) Relay() *Relay {
	return (*Relay)(l)
}

func (r *Relay) Listener() *RelayListener {
	// TODO: Only allow one!
	return (*RelayListener)(r)
}

func (l *RelayListener) Accept() (manet.Conn, error) {
	select {
	case c := <-l.incoming:
		err := l.Relay().writeResponse(c.Stream, pb.CircuitRelay_SUCCESS)
		if err != nil {
			log.Debugf("error writing relay response: %s", err.Error())
			c.Stream.Reset()
			return nil, err
		}

		// TODO: Pretty print.
		log.Infof("accepted relay connection: %s", c)

		return c, nil
	case <-l.ctx.Done():
		return nil, l.ctx.Err()
	}
}

func (l *RelayListener) Addr() net.Addr {
	return &NetAddr{
		Relay:  "any",
		Remote: "any",
	}
}

func (l *RelayListener) Multiaddr() ma.Multiaddr {
	a, err := ma.NewMultiaddr(fmt.Sprintf("/p2p-circuit/ipfs/%s", l.self.Pretty()))
	if err != nil {
		panic(err)
	}
	return a
}

func (l *RelayListener) Close() error {
	// TODO: noop?
	return nil
}
