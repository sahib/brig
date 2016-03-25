package mqtt

import (
	"fmt"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/id"
	"github.com/disorganizer/brig/transfer"
	"github.com/disorganizer/brig/transfer/wire"
	"github.com/gogo/protobuf/proto"
	"golang.org/x/net/context"
)

type Layer struct {
	// self is our own ID and ipfs ID
	self id.Peer
	// srv is a mqtt broker wrapper
	srv *server
	// port of the mqtt broker
	port int
	// own is the client connected to srv
	own *client
	// tab maps IDs top open conversations
	tab map[id.ID]*client
	// ctx is passed to long-running operations that may timeout.
	ctx context.Context
	// cancel interrupts `ctx`.
	cancel context.CancelFunc
	// handler map for RegisterHandler
	handlers map[wire.RequestType]transfer.HandlerFunc
	// Lock for respbox and respctr
	mu sync.Mutex
	// respbox holds all open channels that may be filled
	// with a response. Channels will be deleted after a certain time.
	// Acess is locked by `mu`.
	respbox map[int64]chan *wire.Response
	// respctr is a running counter for responses
	respctr int64
}

func NewLayer(self id.Peer, brokerPort int) transfer.Layer {
	ctx, cancel := context.WithCancel(context.Background())
	return &Layer{
		port:     brokerPort,
		self:     self,
		ctx:      ctx,
		cancel:   cancel,
		tab:      make(map[id.ID]*client),
		handlers: make(map[wire.RequestType]transfer.HandlerFunc),
		respbox:  make(map[int64]chan *wire.Response),
		respctr:  0,
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

func (lay *Layer) Broadcast(req *wire.Request) error {
	data, err := proto.Marshal(req)
	if err != nil {
		return err
	}

	return lay.own.publish(data, lay.own.peerTopic("broadcast"))
}

func (lay *Layer) Connect() (err error) {
	if lay.IsOnlineMode() {
		return nil
	}

	srv, err := newServer(lay.port)
	if err != nil {
		return
	}

	if err = srv.connect(); err != nil {
		return
	}

	lay.srv = srv
	log.Debugf("MQTT broker running on port %d", lay.port)

	var lastError error
	own, err := newClient(lay, lay.self, true)
	log.Debugf("Own MQTT Client connecting to %s", lay.srv.addr())

	for i := 0; i < 10; i++ {
		if err = own.connect(lay.srv.addr()); err != nil {
			lastError = err
			time.Sleep(100 * time.Millisecond)
			continue
		}

		break
	}

	if lastError != nil {
		return lastError
	}

	lay.own = own
	return
}

func (lay *Layer) Disconnect() (err error) {
	if !lay.IsOnlineMode() {
		return nil
	}

	if err = lay.own.disconnect(); err != nil {
		return
	}

	if err = lay.srv.disconnect(); err != nil {
		return
	}

	lay.tab = make(map[id.ID]*client)
	lay.cancel()
	lay.ctx, lay.cancel = context.WithCancel(context.Background())
	lay.srv = nil
	return nil
}

func (lay *Layer) Close() error {
	return lay.Disconnect()
}

func (lay *Layer) Self() id.Peer {
	return lay.self
}

func (lay *Layer) RegisterHandler(typ wire.RequestType, handler transfer.HandlerFunc) {
	lay.handlers[typ] = handler
}

func (lay *Layer) addReqRespPair(req *wire.Request) chan *wire.Response {
	lay.mu.Lock()
	defer lay.mu.Unlock()

	respnotify := make(chan *wire.Response)
	oldCounter := lay.respctr
	lay.respctr++

	lay.respbox[oldCounter] = respnotify
	req.ID = proto.Int64(oldCounter)

	// Remember to purge the channel after a timeout
	// if no response was received in a timely fashion.
	go func() {
		time.Sleep(20 * time.Second)
		lay.mu.Lock()
		delete(lay.respbox, oldCounter)
		lay.mu.Unlock()
	}()

	return respnotify
}

func (lay *Layer) forwardResponse(resp *wire.Response) error {
	lay.mu.Lock()
	defer lay.mu.Unlock()

	respnotify, ok := lay.respbox[resp.GetID()]

	if !ok {
		return fmt.Errorf("No request for ID %d", resp.GetID())
	}

	delete(lay.respbox, resp.GetID())
	respnotify <- resp

	// Remove the result channel again:
	return nil
}

func (lay *Layer) Wait() error {
	for {
		lay.mu.Lock()
		// Timeout in addReqRespPair ensures that
		// lay.respbox will drain with time.
		if len(lay.respbox) == 0 {
			return nil
		}
		lay.mu.Unlock()

		time.Sleep(50 * time.Millisecond)
	}

	return nil
}
