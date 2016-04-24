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
	"github.com/disorganizer/brig/util/ipfsutil"
	"github.com/disorganizer/brig/util/protocol"
	"github.com/disorganizer/brig/util/security"
	"github.com/gogo/protobuf/proto"
)

type Conversation struct {
	sync.Mutex
	conn     net.Conn
	node     *ipfsutil.Node
	proto    *protocol.Protocol
	peer     id.Peer
	notifees map[int64]transfer.AsyncFunc
}

func isEOFError(err error) bool {
	return err == io.EOF || (err != nil && err.Error() == "stream closed")
}

func wrapConnAsProto(conn net.Conn, node *ipfsutil.Node, peerHash string) (*protocol.Protocol, error) {
	pub, err := node.PublicKeyFor(peerHash)
	if err != nil {
		return nil, err
	}

	priv, err := node.PrivateKey()
	if err != nil {
		return nil, err
	}

	// TODO: also sign messages?
	authrw := security.NewAuthReadWriter(conn, priv, pub)
	if err := authrw.Trigger(); err != nil {
		return nil, err
	}

	return protocol.NewProtocol(authrw, true), nil
}

func NewConversation(conn net.Conn, node *ipfsutil.Node, peer id.Peer) (*Conversation, error) {
	proto, err := wrapConnAsProto(conn, node, peer.Hash())
	if err != nil {
		return nil, err
	}

	cnv := &Conversation{
		conn:     conn,
		node:     node,
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

			respID := resp.GetID()

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

func (cnv *Conversation) Close() error {
	return cnv.conn.Close()
}

func (cnv *Conversation) SendAsync(req *wire.Request, callback transfer.AsyncFunc) error {
	cnv.Lock()
	defer cnv.Unlock()

	// Add a nonce so that the same message is guaranteed to result
	// in a different ciphertext:
	req.Nonce = proto.Int64(rand.Int63())

	// Broadcast messages usually do not register a callback.
	// (it wouldn't have been called anyways)
	if callback != nil {
		cnv.notifees[req.GetID()] = callback
	}

	return cnv.proto.Send(req)
}

func (cnv *Conversation) Peer() id.Peer {
	return cnv.peer
}

type handlerMap map[wire.RequestType]transfer.HandlerFunc

type Layer struct {
	node     *ipfsutil.Node
	dialer   transfer.Dialer
	listener net.Listener
	handlers handlerMap

	// Cancellation related:
	parentCtx context.Context
	childCtx  context.Context
	cancel    context.CancelFunc

	// Locking for functions that are not
	// inherently threadsafe
	mu sync.Mutex
}

func NewLayer(node *ipfsutil.Node, parentCtx context.Context) *Layer {
	childCtx, cancel := context.WithCancel(parentCtx)
	return &Layer{
		node:      node,
		parentCtx: parentCtx,
		childCtx:  childCtx,
		cancel:    cancel,
		handlers:  make(handlerMap),
	}
}

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

	return NewConversation(conn, lay.node, peer)
}

func (lay *Layer) IsInOnlineMode() bool {
	lay.mu.Lock()
	defer lay.mu.Unlock()
	return lay.listener != nil
}

func (lay *Layer) handleServerConn(prot *protocol.Protocol) {
	for {
		// Check if we need to quit:
		select {
		case <-lay.childCtx.Done():
			return
		default:
			break
		}

		req := wire.Request{}
		err := prot.Recv(&req)
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Warningf("Server side recv: %v", err)
			break
		}

		log.Debugf("Got request: %v", req)

		typ := req.GetReqType()
		fn, ok := lay.handlers[typ]
		if !ok {
			log.Warningf("Received packet without registerd handler (%d)", typ)
			log.Warningf("Package will be dropped silently.")
			continue
		}

		resp, err := fn(&req)
		if err != nil {
			resp = &wire.Response{
				Error: proto.String(err.Error()),
			}
		}

		if resp == nil {
			// '0' is the ID for broadcast:
			if req.GetID() != 0 {
				log.Warningf("Handle for `%d` failed to return a response or error", typ)
			}
			continue
		}

		// Auto-fill the type and ID fields from the response:
		resp.ReqType = req.ReqType
		resp.ID = req.ID
		resp.Nonce = req.Nonce

		log.Debugf("Sending back %v", resp)

		if err := prot.Send(resp); err != nil {
			log.Warningf("Unable to send back response: %v", err)
			break
		}
	}
}

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
			streamConn, ok := conn.(*ipfsutil.StreamConn)
			if !ok {
				log.Warningf("Denying non-stream conn connection, sorry.")
				return
			}

			// Attempt to establish a full authenticated connection:
			hash := streamConn.PeerHash()
			proto, err := wrapConnAsProto(conn, lay.node, hash)
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

func (lay *Layer) RegisterHandler(typ wire.RequestType, handler transfer.HandlerFunc) {
	lay.handlers[typ] = handler
}

func (lay *Layer) ProtocolID() string {
	return "/brig/moose/v1"
}
