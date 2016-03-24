package mqtt

import (
	"bytes"
	"fmt"
	"net"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/id"
	"github.com/disorganizer/brig/transfer"
	"github.com/disorganizer/brig/transfer/wire"
	"github.com/gogo/protobuf/proto"
	"github.com/surgemq/message"
	"github.com/surgemq/surgemq/service"
)

type client struct {
	layer         *Layer
	client        *service.Client
	peer          id.Peer
	execRequests  bool
	responseTopic string

	// Last time we heard from our peer
	// (not only for ping, but for all operations)
	lastHearbeat time.Time

	// I guess, that's a very "WAT" inducing type signature.
	respbox map[int64]chan *wire.Response
	respctr int64
}

func newClient(lay *Layer, peer id.Peer, execRequests bool) (*client, error) {
	return &client{
		layer:         lay,
		client:        nil,
		execRequests:  execRequests,
		peer:          peer,
		respbox:       make(map[int64]chan *wire.Response),
		responseTopic: "response/" + lay.self.Hash(),
		respctr:       0,
		lastHearbeat:  time.Now(),
	}, nil
}

func (cv *client) peerTopic(sub string) []byte {
	return []byte(fmt.Sprintf("%s/%s", cv.peer.Hash, sub))
}

func (cv *client) formatClientID() []byte {
	return []byte(fmt.Sprintf("brig:%s", cv.peer.Hash))
}

func (cv *client) heartbeat(msg, ack message.Message, err error) error {
	if err != nil {
		return err
	}

	// BEAT IT, JUST BEAT IT!
	cv.lastHearbeat = time.Now()
	return nil
}

func (cv *client) publish(data []byte, topic []byte) error {
	pubmsg := message.NewPublishMessage()
	pubmsg.SetTopic(topic)
	pubmsg.SetPayload(data)
	pubmsg.SetQoS(2)

	return cv.client.Publish(pubmsg, cv.heartbeat)
}

func (cv *client) notifyStatus(status string) error {
	return cv.publish(
		[]byte(status),
		cv.peerTopic("status/"+cv.layer.self.Hash()),
	)
}

func (cv *client) processRequest(msg *message.PublishMessage, answer bool) error {
	if !cv.execRequests {
		return nil
	}

	reqData := msg.Payload()
	req := &wire.Request{}

	if err := proto.Unmarshal(reqData, req); err != nil {
		return err
	}

	handler, ok := cv.layer.handlers[req.GetType()]
	if !ok {
		return fmt.Errorf("No such request type: %d", req.GetType())
	}

	resp, err := handler(cv.layer, req)
	if err != nil {
		return err
	}

	if !answer {
		return nil
	}

	respData, err := proto.Marshal(resp)
	if err != nil {
		return err
	}

	// Publish response:
	if err := cv.publish(respData, cv.peerTopic(cv.responseTopic)); err != nil {
		return err
	}

	return nil
}

func (cv *client) handleStatus(msg *message.PublishMessage) error {
	data := msg.Payload()

	parsed := bytes.SplitN(msg.Topic(), []byte("/"), 1)
	if len(parsed) != 2 {
		return fmt.Errorf("Invalid online notification: %s", msg.Topic())
	}

	// TODO: Somehow update Layer's online infos.
	fmt.Println("%s is going %s", parsed[1], string(data))
	return nil
}

func (cv *client) handleRequests(msg *message.PublishMessage) error {
	return cv.processRequest(msg, true)
}

func (cv *client) handleBroadcast(msg *message.PublishMessage) error {
	return cv.processRequest(msg, false)
}

func (cv *client) handleResponse(msg *message.PublishMessage) error {
	resp := &wire.Response{}
	if err := proto.Unmarshal(msg.Payload(), resp); err != nil {
		return err
	}

	respnotify, ok := cv.respbox[resp.GetID()]
	if !ok {
		// Probably not from us.
		return nil
	}

	respnotify <- resp
	return nil
}

func (cv *client) connect(addr net.Addr) error {
	msg := message.NewConnectMessage()
	msg.SetVersion(4)
	msg.SetCleanSession(true)
	msg.SetClientId(cv.formatClientID())
	msg.SetKeepAlive(300)
	msg.SetWillQos(1)

	// Where to publish our death:
	msg.SetWillTopic(cv.peerTopic("status"))
	msg.SetWillMessage([]byte(
		fmt.Sprintf(
			"%s-%s",
			cv.layer.self.Hash(),
			"offline",
		),
	))

	client := &service.Client{}
	if err := client.Connect("tcp://"+addr.String(), msg); err != nil {
		return err
	}

	topicHandlers := map[string]func(msg *message.PublishMessage) error{
		"status/+":       cv.handleStatus,
		"broadcast":      cv.handleBroadcast,
		cv.responseTopic: cv.handleResponse,
	}

	if cv.execRequests {
		topicHandlers["request"] = cv.handleRequests
	}

	for name, handler := range topicHandlers {
		submsg := message.NewSubscribeMessage()
		submsg.AddTopic(cv.peerTopic(name), 2)

		// There does not seem to be an easier way to register
		// different callbacks per
		if err := client.Subscribe(submsg, cv.heartbeat, handler); err != nil {
			return err
		}
	}

	cv.client = client

	if err := cv.notifyStatus("online"); err != nil {
		log.Warningf("Could not publish an online notify: %v", err)
	}

	return nil
}

func (cv *client) disconnect() error {
	if cv.client != nil {
		return nil
	}

	if err := cv.notifyStatus("offline"); err != nil {
		log.Warningf("Could not publish an offline notify: %v", err)
	}

	cv.client.Disconnect()
	cv.client = nil
	return nil
}

func (cv *client) SendAsync(req *wire.Request, handler transfer.AsyncFunc) error {
	data, err := proto.Marshal(req)
	if err != nil {
		return err
	}

	// Guard against mixed up sends and responses:
	respnotify := make(chan *wire.Response)
	cv.respbox[cv.respctr] = respnotify
	cv.respctr++

	// Start before publish to fix a very unlikely race.
	go func() {
		// Guard with a timeout to protect against too many go routines.
		ticker := time.NewTicker(30 * time.Second)

		select {
		case resp, ok := <-respnotify:
			if resp != nil && ok {
				handler(resp)
			}
		case <-ticker.C:
		}

		// Remove the result channel again:
		delete(cv.respbox, cv.respctr)
	}()

	if err := cv.publish(data, cv.peerTopic("request")); err != nil {
		return err
	}

	return nil
}

func (cv *client) Close() error {
	return cv.disconnect()
}

func (cv *client) ping() (bool, error) {
	if cv.client == nil {
		return false, transfer.ErrOffline
	}

	if time.Since(cv.lastHearbeat) < 10*time.Second {
		return true, nil
	}

	// Ping() seems to wait for the ACK.
	if err := cv.client.Ping(cv.heartbeat); err != nil {
		return false, err
	}

	return true, nil
}

func (cv *client) Peer() id.Peer {
	return cv.peer
}
