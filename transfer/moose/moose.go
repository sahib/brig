package moose

import (
	"io"
	"math/rand"
	"net"
	"sync"

	"golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/id"
	"github.com/disorganizer/brig/transfer"
	"github.com/disorganizer/brig/transfer/wire"
	"github.com/disorganizer/brig/util/protocol"
	"github.com/disorganizer/brig/util/security"
)

// Conversation implements layer.Conversation
// by using ipfs' swarming functionality and building a protocol on top.
type Conversation struct {
	sync.Mutex
	conn     net.Conn
	backend  Backend
	proto    *protocol.Protocol
	peer     id.Peer
	notifees map[int64]transfer.AsyncFunc
}

// isEOFError checks if `err` means something like io.EOF.
// Sadly, we need to match the error string since no distinct error exists.
func isEOFError(err error) bool {
	return err == io.EOF || (err != nil && err.Error() == "stream closed")
}

// wrapConnAsProto establishes the moose protocol on the raw ipfs connection
func wrapConnAsProto(conn net.Conn, bk Backend, peerHash string) (*protocol.Protocol, error) {
	pub, err := bk.PublicKeyFor(peerHash)
	if err != nil {
		return nil, err
	}

	priv, err := bk.PrivateKey()
	if err != nil {
		return nil, err
	}

	authrw := security.NewAuthReadWriter(conn, priv, pub)
	if err := authrw.Trigger(); err != nil {
		return nil, err
	}

	return protocol.NewProtocol(authrw, true), nil
}

// NewConversation returns a conversation that exchanges data over `conn`.
func NewConversation(conn net.Conn, backend Backend, peer id.Peer) (*Conversation, error) {
	proto, err := wrapConnAsProto(conn, bk, peer.Hash())
	if err != nil {
		return nil, err
	}

	cnv := &Conversation{
		conn:     conn,
		backend:  backend,
		peer:     peer,
		proto:    proto,
		notifees: make(map[int64]transfer.AsyncFunc),
	}

	// Cater responses:
	go func() {
		for {
			resp := wire.Response{}
			err := cnv.proto.Recv(&resp)

			if isEOFError(err) {
				break
			}

			if err != nil {
				log.Warningf("Error while receiving data: %v", err)
				continue
			}

			respID := resp.ID

			cnv.Lock()
			fn, ok := cnv.notifees[respID]
			if !ok {
				log.Warningf("No such id: %v", respID)
				cnv.Unlock()
				continue
			}

			// Remove the callback
			delete(cnv.notifees, respID)
			cnv.Unlock()

			fn(&resp)
		}
	}()

	return cnv, nil
}

// Close terminates the conversation by closing the underlying connection
func (cnv *Conversation) Close() error {
	return cnv.conn.Close()
}

// SendAsync sends `req` to the other end and calls callback on the response.
func (cnv *Conversation) SendAsync(req *wire.Request, callback transfer.AsyncFunc) error {
	cnv.Lock()
	defer cnv.Unlock()

	// Add a nonce so that the same message is guaranteed to result
	// in a different ciphertext:
	req.Nonce = rand.Int63()

	// Broadcast messages usually do not register a callback.
	// (it wouldn't have been called anyways)
	if callback != nil {
		cnv.notifees[req.ID] = callback
	}

	return cnv.proto.Send(req)
}

// Peer returns the remote peer this conversation is connected to.
func (cnv *Conversation) Peer() id.Peer {
	return cnv.peer
}

// typedef that so we don't need to repeat that long type...
type handlerMap map[wire.RequestType]transfer.HandlerFunc

// Layer implements the moose protocol by handling incoming requests
// and creating Conversations to other peers.
type Layer struct {
	// Core functionality:
	backend  Backend
	dialer   transfer.Dialer
	listener net.Listener
	handlers handlerMap

	// Cancellation related:
	parentCtx context.Context
	childCtx  context.Context
	cancel    context.CancelFunc

	// Locking for functions that are not
	// inherently threadsafe.
	mu sync.Mutex
}

// NewLayer returns a freshly setup layer that is not connected yet.
func NewLayer(backend Backend, parentCtx context.Context) *Layer {
	childCtx, cancel := context.WithCancel(parentCtx)
	return &Layer{
		backend:   Backend,
		parentCtx: parentCtx,
		childCtx:  childCtx,
		cancel:    cancel,
		handlers:  make(handlerMap),
	}
}

// Dial creates a new and ready Conversation to `peer`
func (lay *Layer) Dial(peer id.Peer) (transfer.Conversation, error) {
	if !lay.IsInOnlineMode() {
		return nil, transfer.ErrOffline
	}

	lay.mu.Lock()
	defer lay.mu.Unlock()

	conn, err := lay.dialer.Dial(peer)
	if err != nil {
		return nil, err
	}

	return NewConversation(conn, lay.backend, peer)
}

// IsInOnlineMode returns true after an success Connect()
func (lay *Layer) IsInOnlineMode() bool {
	lay.mu.Lock()
	defer lay.mu.Unlock()
	return lay.listener != nil
}

// handleServerConn handles incoming messages and calls the respective
// handler on the request. Response are transmitted back afterwards.
func (lay *Layer) handleServerConn(prot *protocol.Protocol) {
	for lay.loopServerConn(prot) {
		// ...
	}
}

func (lay *Layer) loopServerConn(prot *protocol.Protocol) bool {
	// Check if we need to quit:
	select {
	case <-lay.childCtx.Done():
		return false
	default:
		break
	}

	req := wire.Request{}
	if err := prot.Recv(&req); err != nil {
		if err != io.EOF {
			log.Warningf("Server side recv: %v", err)
		}

		return false
	}

	log.Debugf("Got request: %v", req)
	fn, ok := lay.handlers[req.ReqType]
	if !ok {
		log.Warningf("Received packet without registerd handler (%d)", req.ReqType)
		log.Warningf("Package will be dropped.")
		return true
	}

	resp, err := fn(&req)
	if err != nil {
		resp = &wire.Response{
			Error: err.Error(),
		}
	}

	if resp == nil {
		// '0' is the ID for broadcast. Empty response are valid there.
		if req.ID != 0 {
			log.Warningf("Handle for `%d` failed to return a response or error", req.ReqType)
		}

		return true
	}

	// Auto-fill the type and ID fields from the response:
	resp.ReqType = req.ReqType
	resp.ID = req.ID
	resp.Nonce = req.Nonce

	log.Debugf("Sending back %v", resp)
	if err := prot.Send(resp); err != nil {
		log.Warningf("Unable to send back response: %v", err)
		return false
	}

	return true
}

// Connect will start listening on incoming connections and remember
// dialer so that calls to Dial() can succeed.
func (lay *Layer) Connect(l net.Listener, d transfer.Dialer) error {
	lay.mu.Lock()
	defer lay.mu.Unlock()

	lay.dialer = d
	lay.listener = l
	lay.childCtx, lay.cancel = context.WithCancel(lay.parentCtx)

	// Listen for incoming connections as long the listener is open:
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				// *sigh* Again, not my fault.
				if err != transfer.ErrListenerWasClosed {
					if err.Error() != "context canceled" {
						log.Warningf("Listener: %T '%v'", err, err)
					}
				}

				break
			}

			// We currently rely on an ipfs connection here,
			// so testing it without ipfs is not directly possible.
			streamConn, ok := conn.(StreamConn)
			if !ok {
				log.Warningf("Denying non-stream conn connection, sorry.")
				return
			}

			// Attempt to establish a full authenticated connection:
			hash := streamConn.PeerHash()
			proto, err := wrapConnAsProto(conn, lay.backend, hash)
			if err != nil {
				log.Warningf(
					"Could not establish incoming connection to %s: %v",
					hash, err,
				)
				return
			}

			// Handle protocol in server mode:
			go lay.handleServerConn(proto)
		}
	}()

	return nil
}

// Disconnect brings down all resources needed to server responses.
func (lay *Layer) Disconnect() error {
	lay.mu.Lock()
	defer lay.mu.Unlock()

	// Bring down the server-side handlers:
	lay.cancel()

	// This should break the loop in Connect()
	if err := lay.listener.Close(); err != nil {
		return err
	}

	lay.listener = nil
	lay.dialer = nil

	return nil
}

// RegisterHandler remembers to call `handler` for `typ`.
// It is no error to call RegisterHandler twice for the same typ.
func (lay *Layer) RegisterHandler(typ wire.RequestType, handler transfer.HandlerFunc) {
	lay.handlers[typ] = handler
}

// ProtocolID returns the name of the moose protocol.
func (lay *Layer) ProtocolID() string {
	return "/brig/moose/v1"
}
