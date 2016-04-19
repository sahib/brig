package transfer

import (
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/disorganizer/brig/id"
	"github.com/disorganizer/brig/transfer/wire"
)

var (
	// ErrOffline is returned when an online operation is requested during
	// offline mode.
	ErrOffline = errors.New("Transfer layer is offline")
)

// Conversation is a open channel to another peer
// used to exchange metadata over protobuf messages.
type Conversation interface {
	io.Closer

	// Send delivers `req` exactly once to the conversation peer.
	//
	// The message might be any proto.Message,
	// but is usually wire.Request on the client side
	// and wire.Response on the server side.
	// `callback` will not be called if no answer was received.
	// `callback` may be nil for fire-and-forget messages.
	//
	// How requests are actually handled and processed into responses,
	// is depended on the handler you passed to RegisterHandler().
	SendAsync(req *wire.Request, callback AsyncFunc) error

	// Peer returns the peer we're talking to.
	Peer() id.Peer
}

// HandlerFunc handles a single wire.Request and returns
// a fitting wire.Response.
type HandlerFunc func(*wire.Request) (*wire.Response, error)

// AsyncFunc is used as argument to SendAsync
// It will be called whenever a response arrives at the layer.
type AsyncFunc func(resp *wire.Response)

// Dialer implementors define how the layer connects to the outside world.
type Dialer interface {
	// Dial shall create a, possibly unencrypted, connection to the peer
	// at `peer`. If succesfull a working network connection should be returned.
	Dial(peer id.Peer) (net.Conn, error)
}

// Layer is the interface that all metadata-networking layers
// of brig have to fulfill.
type Layer interface {
	// Dial opens a new connection to the peer conn is opened to.
	// Dial() shall return ErrOffline when not in online mode.
	Dial(peer id.Peer) (Conversation, error)

	// IsOnlineMode returns true if the layer is online and may respond
	// to requests or send requests itself. It should be true after
	// a succesful Connect().
	IsInOnlineMode() bool

	// Connect to the net. A freshly created Layer should not be
	// connected upon construction. The passed listener will
	// be used to listen on new network connections from outside
	// and dialer will be used to dial to the outside.
	//
	// A Connect() when IsOnlineMode() is true is a no-op.
	Connect(l net.Listener, d Dialer) error

	// Disconnect from the net.
	// A Disconnect() when IsOnlineMode() is false is a no-op.
	Disconnect() error

	// RegisterHandler will register  a handler for the request type `typ`.
	// `handler` will be called once a request with this type is received.
	RegisterHandler(typ wire.RequestType, handler HandlerFunc)

	// ProtocolID returns a unique protocol id
	// that will be used to differentiate between other protocols.
	// Example: "/brig/mqtt/v1"
	ProtocolID() string
}

// AuthManager shall be passed to a layer upon creation.
// The layer will use it to encrypt the communication
// between the peers and handle the login procedure.
type AuthManager interface {
	// Authenticate decides wether `id` is allowed to access
	// this node when by looking at the credentials in `cred`.
	Authenticate(id string, cred []byte) error

	// Credentials return the login credentials used
	// when talking to other clients.
	Credentials(peer id.Peer) ([]byte, error)

	// TunnelFor should return a AuthTunnel that
	// encrypts the traffic between us and `id`.
	TunnelFor(id id.Peer) (AuthTunnel, error)
}

// AuthTunnel secures the communication between two peers
// in a implementation defined manner.
type AuthTunnel interface {
	// Encrypt encrypts `data`.
	Encrypt(data []byte) ([]byte, error)

	// Decrypt decrypts `data`.
	Decrypt(data []byte) ([]byte, error)
}

// mockAuthManager fullfills AuthManager by doing nothing.
// It is meant for tests. Production code users will be shot.
type mockAuthManager bool

// Authenticate just nods yes to everything.
func (y mockAuthManager) Authenticate(id string, cred []byte) error {
	if y {
		return nil
	} else {
		return fmt.Errorf("You shall not pass")
	}
}

func (_ mockAuthManager) Credentials(id id.Peer) ([]byte, error) {
	return []byte("wald"), nil
}

func (_ mockAuthManager) TunnelFor(id id.Peer) (AuthTunnel, error) {
	return mockAuthTunnel{}, nil
}

type mockAuthTunnel struct{}

// Encrypt does not encrypt. It returns `data`.
func (_ mockAuthTunnel) Encrypt(data []byte) ([]byte, error) {
	return data, nil
}

// Decrypt does not decrypt. It returns `data`.
func (mam mockAuthTunnel) Decrypt(data []byte) ([]byte, error) {
	return data, nil
}

var (
	// MockAuthSuccess is a AuthManager that blindly lets everything through
	MockAuthSuccess = mockAuthManager(true)
	// MockAuthDeny is a AuthManager that denies every access
	MockAuthDeny = mockAuthManager(false)
)
