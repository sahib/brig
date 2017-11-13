package backend

import (
	"net"

	"github.com/disorganizer/brig/net/peer"
)

type Backend interface {
	// ResolveName resolves a human readable `name` to a list of peers.
	// Each of these can be contacted to check their credentials.
	// If the backend support exact lookups, this method will only
	// return one peer on success always.
	ResolveName(name peer.Name) ([]peer.Info, error)

	// Identity resolves our own name to an addr that we could pass to Dial.
	// It is used as part of the brig identifier for others.
	Identity() (peer.Info, error)

	// Dial builds up a connection to another peer.
	// If only ever one protocol is used, just pass the same string always.
	Dial(peerAddr, protocol string) (net.Conn, error)

	// Listen returns a listener, that will yield incoming connections
	// from other peers when calling Accept.
	Listen(protocol string) (net.Listener, error)
}
