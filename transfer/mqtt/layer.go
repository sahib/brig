package mqtt

import (
	"fmt"

	"github.com/disorganizer/brig/id"
	"github.com/disorganizer/brig/transfer"
	"github.com/disorganizer/brig/transfer/proto"
	protobuf "github.com/gogo/protobuf/proto"
	"golang.org/x/net/context"
)

type Layer struct {
	// self is our own ID and ipfs ID
	self id.Peer
	// srv is a mqtt broker wrapper
	srv *server
	// own is the client connected to srv
	own *client
	// tab maps IDs top open conversations
	tab map[id.ID]*client
	// ctx is passed to long-running operations that may timeout.
	ctx context.Context
	// cancel interrupts `ctx`.
	cancel context.CancelFunc
	//
	handlers map[proto.RequestType]transfer.HandlerFunc
}

func NewLayer(self id.Peer) transfer.Layer {
	ctx, cancel := context.WithCancel(context.Background())
	return &Layer{
		self:     self,
		ctx:      ctx,
		cancel:   cancel,
		tab:      make(map[id.ID]*client),
		handlers: make(map[proto.RequestType]transfer.HandlerFunc),
	}
}

func (lay *Layer) Talk(rslv id.Resolver) (transfer.Conversation, error) {
	if !lay.IsOnlineMode() {
		return nil, transfer.ErrOffline
	}

	addrs, err := rslv.Resolve(lay.ctx)
	if err != nil {
		return nil, err
	}

	cnv, err := newClient(lay, rslv.Peer(), false)
	if err != nil {
		return nil, err
	}

	// TODO: This is brute force.
	var lastError error

	for _, addr := range addrs {
		if err := cnv.connect(addr); err != nil {
			lastError = err
		} else {
			break
		}
	}

	if lastError != nil {
		cnv = nil
	}

	return cnv, lastError
}

func (lay *Layer) IsOnline(peer id.ID) (bool, error) {
	if !lay.IsOnlineMode() {
		return false, transfer.ErrOffline
	}

	if peer == lay.self.ID() {
		return true, nil
	}

	client, ok := lay.tab[peer]
	if !ok {
		return false, fmt.Errorf("No peer with that ID: %s", peer)
	}

	reachable, err := client.ping()
	if err != nil {
		return false, err
	}

	return reachable, nil
}

func (lay *Layer) IsOnlineMode() bool {
	return lay.srv != nil
}

func (lay *Layer) Broadcast(req *proto.Request) error {
	data, err := protobuf.Marshal(req)
	if err != nil {
		return err
	}

	return lay.own.publish(data, lay.own.peerTopic("broadcast"))
}

func (lay *Layer) Connect() (err error) {
	if lay.IsOnlineMode() {
		return nil
	}

	// TODO: Pass correct port in
	srv, err := newServer(1883)
	if err != nil {
		return
	}

	lay.srv = srv

	own, err := newClient(lay, lay.self, true)
	if err = own.connect(lay.srv.addr()); err != nil {
		return
	}

	lay.own = own
	return
}

func (lay *Layer) Disconnect() (err error) {
	if !lay.IsOnlineMode() {
		return nil
	}

	if err = lay.srv.disconnect(); err != nil {
		return
	}

	lay.tab = make(map[id.ID]*client)
	lay.cancel()
	lay.ctx, lay.cancel = context.WithCancel(context.Background())
	return nil
}

func (lay *Layer) Close() error {
	return lay.Disconnect()
}

func (lay *Layer) RegisterHandler(typ proto.RequestType, handler transfer.HandlerFunc) {
	lay.handlers[typ] = handler
}
