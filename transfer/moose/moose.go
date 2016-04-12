package moose

import (
	"net"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/id"
	"github.com/disorganizer/brig/transfer"
	"github.com/disorganizer/brig/transfer/wire"
	"github.com/disorganizer/brig/util"
	"github.com/disorganizer/brig/util/ipfsutil"
	"github.com/disorganizer/brig/util/protocol"
	"github.com/disorganizer/brig/util/tunnel"
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

func wrapConnAsProto(conn net.Conn) (*protocol.Protocol, error) {
	tnl, err := tunnel.NewEllipticTunnel(conn)
	if err != nil {
		return nil, err
	}

	if err := tnl.Exchange(); err != nil {
		return nil, err
	}

	// TODO: auth
	return protocol.NewProtocol(tnl, true), nil
}

func NewConversation(conn net.Conn, node *ipfsutil.Node, peer id.Peer) (*Conversation, error) {
	proto, err := wrapConnAsProto(conn)
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
			if err := cnv.proto.Recv(&resp); err != nil {
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
	// TODO: Rate limiting to counteract DDOS:
	//       only let cnv.notifees grow up to a certain size
	cnv.Lock()
	defer cnv.Unlock()

	cnv.notifees[req.GetID()] = callback
	return cnv.proto.Send(req)
}

func (cnv *Conversation) Peer() id.Peer {
	return cnv.peer
}

type Layer struct {
	node     *ipfsutil.Node
	dialer   transfer.Dialer
	listener net.Listener

	open     map[id.Peer]Conversation
	handlers map[wire.RequestType]transfer.HandlerFunc
}

func NewLayer(node *ipfsutil.Node) *Layer {
	return &Layer{
		node: node,
		open: make(map[id.Peer]Conversation),
	}
}

func (lay *Layer) Close() error {
	var errs util.Errors

	for _, cnv := range lay.open {
		if err := cnv.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func (lay *Layer) Dial(peer id.Peer) (transfer.Conversation, error) {
	conn, err := lay.dialer.Dial(peer)
	if err != nil {
		return nil, err
	}

	return NewConversation(conn, lay.node, peer)
}

func (lay *Layer) IsInOnlineMode() bool {
	return lay.listener == nil
}

func (lay *Layer) handleServerConn(conn net.Conn) {
	prot := protocol.NewProtocol(conn, true)
	for {
		req := wire.Request{}
		if err := prot.Recv(&req); err != nil {
			log.Warning("Server side recv: %v", err)
			break
		}

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
			// No response is valid too.
			continue
		}

		if err := prot.Send(resp); err != nil {
			log.Warningf("Unable to send back response: %v", err)
			break
		}
	}
}

func (lay *Layer) Connect(l net.Listener, d transfer.Dialer) error {
	lay.dialer = d
	lay.listener = l

	// Listen for incoming connections as long the listener is open:
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				log.Warningf("Listener: %v", err)
				break
			}

			// Handle conn in server mode:
			go lay.handleServerConn(conn)
		}
	}()

	return nil
}

func (lay *Layer) Disconnect() error {
	l := lay.listener
	lay.listener = nil
	lay.dialer = nil

	return l.Close()
}

func (lay *Layer) RegisterHandler(typ wire.RequestType, handler transfer.HandlerFunc) {
	lay.handlers[typ] = handler
}

func (lay *Layer) Wait() error {
	// TODO
	return nil
}

func (lay *Layer) ProtocolID() string {
	return "/brig/moose/v1"
}
