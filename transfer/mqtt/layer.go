package mqtt

import (
	"fmt"
	"net"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/id"
	"github.com/disorganizer/brig/transfer"
	"github.com/disorganizer/brig/transfer/wire"
	"github.com/gogo/protobuf/proto"
	"golang.org/x/net/context"
)

type layer struct {
	// self is our own ID and ipfs ID
	self id.Peer

	// srv is a mqtt broker wrapper
	srv *server

	// own is the client connected to srv
	own *client

	// dialing interface for outside connections
	dialer transfer.Dialer

	// tab maps peer hashes to open conversations
	tab map[string]*client

	// ctx is passed to long-running operations that may timeout.
	ctx context.Context

	// cancel interrupts `ctx`.
	cancel context.CancelFunc

	// handler map for RegisterHandler
	handlers map[wire.RequestType]transfer.HandlerFunc
	// Lock for respbox and respctr
	resplock sync.Mutex

	// respbox holds all open channels that may be filled
	// with a response. Channels will be deleted after a certain time.
	// Acess is locked by `mu`.
	respbox map[int64]chan *wire.Response

	// respctr is a running counter for responses
	respctr int64

	// Used to handle authentication and encryption
	authMgr transfer.AuthManager
}

func NewLayer(self id.Peer, authMgr transfer.AuthManager) transfer.Layer {
	ctx, cancel := context.WithCancel(context.Background())
	return &layer{
		self:     self,
		ctx:      ctx,
		cancel:   cancel,
		tab:      make(map[string]*client),
		handlers: make(map[wire.RequestType]transfer.HandlerFunc),
		respbox:  make(map[int64]chan *wire.Response),
		respctr:  0,
		authMgr:  authMgr,
	}
}

func (lay *layer) Talk(peer id.Peer) (transfer.Conversation, error) {
	if !lay.IsOnlineMode() {
		return nil, transfer.ErrOffline
	}

	tunnel, err := lay.authMgr.TunnelFor(peer)
	if err != nil {
		return nil, err
	}

	cnv, ok := lay.tab[peer.Hash()]
	if ok {
		return cnv, nil
	}

	cnv, err = newClient(lay, tunnel, peer, false)
	if err != nil {
		return nil, err
	}

	conn, err := lay.dialer.Dial(peer)
	if err != nil {
		return nil, err
	}

	if err := cnv.connect(conn); err != nil {
		return nil, err
	}

	lay.tab[peer.Hash()] = cnv
	return cnv, nil
}

func (lay *layer) IsOnline(peer id.Peer) bool {
	if !lay.IsOnlineMode() {
		return false
	}

	if peer.Hash() == lay.self.Hash() {
		return true
	}

	client, ok := lay.tab[peer.Hash()]
	if !ok {
		return false
	}

	reachable, err := client.ping()
	if err != nil {
		log.Debugf("Ping failed: %v", err)
		return false
	}

	return reachable
}

func (lay *layer) IsOnlineMode() bool {
	return lay.srv != nil
}

func (lay *layer) Broadcast(req *wire.Request) error {
	req.ID = proto.Int64(-1)

	data, err := lay.own.protoToPayload(req)
	if err != nil {
		return err
	}

	return lay.own.publish(data, lay.own.peerTopic("broadcast"))
}

func (lay *layer) Connect(listener net.Listener, dialer transfer.Dialer) (err error) {
	if lay.IsOnlineMode() {
		return nil
	}

	lay.dialer = dialer

	// Get a encrypted tunnel for the own client:
	tunnel, tnlErr := lay.authMgr.TunnelFor(lay.self)
	if tnlErr != nil {
		return tnlErr
	}

	srv, err := newServer(lay, listener)
	if err != nil {
		return
	}

	if err = srv.connect(); err != nil {
		return
	}

	lay.srv = srv

	var lastError error
	own, err := newClient(lay, tunnel, lay.self, true)
	log.Debugf("Own MQTT Client connecting")

	for i := 0; i < 10; i++ {
		conn, err := dialer.Dial(lay.self)
		if err == nil && conn != nil {
			if err = own.connect(conn); err == nil {
				fmt.Println("Connected!")
				break
			}
		}

		if conn != nil {
			conn.Close()
		}

		fmt.Println("Connect", err, conn)
		lastError = err
		time.Sleep(100 * time.Millisecond)
	}

	if lastError != nil {
		return lastError
	}

	lay.own = own
	return
}

func (lay *layer) Disconnect() (err error) {
	if !lay.IsOnlineMode() {
		return nil
	}

	if err = lay.own.disconnect(); err != nil {
		return
	}

	if err = lay.srv.disconnect(); err != nil {
		return
	}

	lay.tab = make(map[string]*client)
	lay.cancel()
	lay.ctx, lay.cancel = context.WithCancel(context.Background())
	lay.srv = nil
	return nil
}

func (lay *layer) Close() error {
	return lay.Disconnect()
}

func (lay *layer) Self() id.Peer {
	return lay.self
}

func (lay *layer) RegisterHandler(typ wire.RequestType, handler transfer.HandlerFunc) {
	lay.handlers[typ] = handler
}

func (lay *layer) addReqRespPair(req *wire.Request) chan *wire.Response {
	lay.resplock.Lock()
	defer lay.resplock.Unlock()

	respnotify := make(chan *wire.Response)
	oldCounter := lay.respctr
	lay.respctr++

	lay.respbox[oldCounter] = respnotify
	req.ID = proto.Int64(oldCounter)

	// Remember to purge the channel after a timeout
	// if no response was received in a timely fashion.
	go func() {
		time.Sleep(20 * time.Second)
		lay.resplock.Lock()
		delete(lay.respbox, oldCounter)
		lay.resplock.Unlock()
	}()

	return respnotify
}

func (lay *layer) forwardResponse(resp *wire.Response) error {
	lay.resplock.Lock()
	defer lay.resplock.Unlock()

	respnotify, ok := lay.respbox[resp.GetID()]

	if !ok {
		return fmt.Errorf("No request for ID %d", resp.GetID())
	}

	delete(lay.respbox, resp.GetID())
	respnotify <- resp

	// Remove the result channel again:
	return nil
}

func (lay *layer) Wait() error {
	for {
		lay.resplock.Lock()
		// Timeout in addReqRespPair ensures that
		// lay.respbox will drain with time.
		if len(lay.respbox) == 0 {
			return nil
		}
		lay.resplock.Unlock()

		time.Sleep(50 * time.Millisecond)
	}

	return nil
}
