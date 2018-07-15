package stream

import (
	"fmt"

	inet "gx/ipfs/QmPjvxTpVH8qJyQDnxnsxF9kv9jezKD1kozz1hs3fCGsNh/go-libp2p-net"
	transport "gx/ipfs/QmW5LxJm2Yo3S7uVsfLM7NsJn2QnKbvZD7uYsZVYR7YViE/go-libp2p-transport"
	smux "gx/ipfs/QmY9JXR3FupnYAYJWK9aMr9bCpqWKcToQ1tz8DVGTrHpHw/go-stream-muxer"
)

type transportConn struct {
	smux.Conn
	inet.ConnMultiaddrs
	inet.ConnSecurity
	transport transport.Transport
}

func (t *transportConn) Transport() transport.Transport {
	return t.transport
}

func (t *transportConn) String() string {
	ts := ""
	if s, ok := t.transport.(fmt.Stringer); ok {
		ts = "[" + s.String() + "]"
	}
	return fmt.Sprintf(
		"<stream.Conn%s %s (%s) <-> %s (%s)>",
		ts,
		t.LocalMultiaddr(),
		t.LocalPeer(),
		t.RemoteMultiaddr(),
		t.RemotePeer(),
	)
}
