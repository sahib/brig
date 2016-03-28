package transfer

import (
	"errors"
	"io"

	"github.com/disorganizer/brig/id"
	"github.com/disorganizer/brig/transfer/wire"
)

var (
	// ErrOffline is returned when an online operation is requested during
	// offline mode.
	ErrOffline = errors.New("Transfer layer is offline")
)

// AsyncFunc is used as argument to SendAsync
// It will be called whenever a response arrives at the layer.
type AsyncFunc func(resp *wire.Response)

// Conversation is a open channel to another peer
// used to exchange metadata over protobuf messages.
type Conversation interface {
	io.Closer

	// Send delivers `req` exactly once to the conversation peer.
	// TODO: handle commands docs?
	//
	// The message might be any proto.Message,
	// but is usually wire.Request on the client side
	// and wire.Response on the server side.
	// `callback` will not be called if no answer was received.
	// `callback` may be nil for fire-and-forget messages.
	SendAsync(req *wire.Request, callback AsyncFunc) error

	// Peer returns the peer we're talking to.
	Peer() id.Peer
}

// HandlerFunc handles a single wire.Request and returns
// a fitting wire.Response.
type HandlerFunc func(*wire.Request) (*wire.Response, error)

// Layer is the interface that all metadata-networking layers
// of brig have to fulfill.
type Layer interface {
	io.Closer

	// Talk opens a new connection to the peer pointed to by `id`.
	// The peer should have the peer id presented in `rslv.Peer().ID()`
	// in order to authenticate itself.
	//
	// Talk() shall return ErrOffline when not in online mode.
	Talk(rslv id.Resolver) (Conversation, error)

	// IsOnline shall return true if the peer knows as `peer` is online and
	// responding. It is allowed that the implementation may cache the
	// answer for a short time, therefore changes might be visible only
	// after a certain timeout.
	//
	// IsOnline only can give sensible answers when IsOnlineMode() is true
	// and we're currently talking to this peer.
	// If you want to "peek" if the other client is online, you have
	// to do a Talk() before.
	IsOnline(peer id.Peer) bool

	// IsOnlineMode returns true if the layer is online and may respond
	// to requests or send requests itself. It should be true after
	// a succesful Connect().
	IsOnlineMode() bool

	// Connect to the net. A freshly created Layer should not be
	// connected upon construction.
	//
	// A Connect() when IsOnlineMode() is true is a no-op.
	Connect() error

	// Disconnect from the net.
	// A Disconnect() when IsOnlineMode() is false is a no-op.
	Disconnect() error

	// RegisterHandler will register  a handler for the request type `typ`.
	// `handler` will be called once a request with this type is received.
	RegisterHandler(typ wire.RequestType, handler HandlerFunc)

	// Broadcast sends a request to all connected peers.
	// No answers will be collected.
	// It's usecase is to send quick updates to all peers.
	Broadcast(req *wire.Request) error

	// Self returns the peer we're acting as.
	Self() id.Peer

	// Wait blocks until all requests were handled.
	// It is mainly useful for debugging.
	Wait() error
}

// AuthManager shall be passed to a layer upon creation.
// The layer will use it to encrypt the communication
// between the peers and handle the login procedure.
type AuthManager interface {
	Authenticate(id id.Peer, cred interface{}) error
	Encrypt(data []byte) ([]byte, error)
	Decrypt(data []byte) ([]byte, error)
}

// MockAuthManager fullfills AuthManager by doing nothing.
// It is meant for tests. Production code users will be shot.
type MockAuthManager struct{}

// Authenticate just nods yes to everything.
func (mam MockAuthManager) Authenticate(id id.Peer, cred interface{}) error {
	return nil
}

// Encrypt does not encrypt. It returns `data`.
func (mam MockAuthManager) Encrypt(data []byte) ([]byte, error) {
	return data, nil
}

// Decrypt does not decrypt. It returns `data`.
func (mam MockAuthManager) Decrypt(data []byte) ([]byte, error) {
	return data, nil
}
