package backend

import (
	stdnet "net"
	"time"

	"github.com/sahib/brig/net/peer"
)

// Pinger is a watcher for a single peer that will actively ping
// the peer until closed. Time between pings is chosen by the backend.
type Pinger interface {
	// LastSeen returns a timestamp of when this peer last responded.
	LastSeen() time.Time

	// Roundtrip returns the time needed to send a small package to a peer.
	Roundtrip() time.Duration

	// Err returns a non-nil value if the last try to contact this peer failed.
	Err() error

	// Close shuts down this pinger.
	Close() error
}

// Backend defines all required methods needed from the underyling
// implementation in order to talk with other nodes.
type Backend interface {
	// ResolveName resolves a human readable `name` to a list of peers.
	// Each of these can be later contacted to check their credentials.
	// The operation may take at max `timeoutSec`.
	ResolveName(name string, timeoutSec int) ([]peer.Info, error)

	// PublishName announces to the network that this node is known as `name`.
	// If possible also the group and domain name of the name should be
	// announced.
	PublishName(name string) error

	// Identity resolves our own name to an addr that we could pass to Dial.
	// It is used as part of the brig identifier for others.
	Identity() (peer.Info, error)

	// Dial builds up a connection to another peer.
	// If only ever one protocol is used, just pass the same string always.
	Dial(peerAddr, protocol string) (stdnet.Conn, error)

	// Listen returns a listener, that will yield incoming connections
	// from other peers when calling Accept.
	Listen(protocol string) (stdnet.Listener, error)

	// Ping returns a Pinger interface for the peer at `peerAddr`.
	// It should not create a full
	Ping(peerAddr string) (Pinger, error)

	// Connect will connect to the common network
	Connect() error

	// Disconnect will reject incoming connections and disallow outgoing.
	Disconnect() error

	// IsOnline should return true if the node is currently able to contact
	// or receive connections from other peers.
	IsOnline() bool
}
