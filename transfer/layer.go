package transfer

import (
	"errors"
	"io"

	"github.com/disorganizer/brig/id"
	"github.com/disorganizer/brig/transfer/wire"
	"github.com/gogo/protobuf/proto"
)

var (
	// ErrOffline is returned when an online operation is requested during
	// offline mode.
	ErrOffline = errors.New("Transfer layer is offline")
)

type AsyncFunc func(resp proto.Message)

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
	SendAsync(req *wire.Request, callback AsyncFunc) error

	// Peer returns the peer we're talking to.
	Peer() id.Peer
}

// HandlerFunc handles a single wire.Request and returns
// a fitting wire.Response.
type HandlerFunc func(Layer, *wire.Request) (*wire.Response, error)

// Layer is the interface that all metadata-networking layers
// of brig have to fulfill.
type Layer interface {
	io.Closer

	// Talk opens a new connection to the peer pointed to by `id`.
	// The peer should have the peer id presented in `rslv.Peer().ID()`
	// in order to authenticate itself.
	//
	// Talk() shall return ErrOffline when not in online mode.
	// TODO pass additional credentials.
	Talk(rslv id.Resolver) (Conversation, error)

	// IsOnline shall return true if the peer knows as `id` is online and
	// responding. It is allowed that the implementation may cache the
	// answer for a short time.
	IsOnline(ident id.ID) (bool, error)

	// IsOnlineMode returns true if the layer is online and may respond
	// to requests or send requests itself. It should be true after
	// a succesful Connect().
	IsOnlineMode() bool

	// Connect to the net. A freshly created Layer should not be
	// connected upon construction.
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
}

// TODO: Interface for authentication?
