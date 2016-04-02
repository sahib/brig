package mqtt

import (
	"bytes"
	"fmt"
	"net"
	"sync/atomic"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/bkaradzic/go-lz4"
	"github.com/disorganizer/brig/id"
	"github.com/disorganizer/brig/transfer"
	"github.com/disorganizer/brig/transfer/wire"
	"github.com/disorganizer/surgemq/service"
	"github.com/gogo/protobuf/proto"
	"github.com/surgemq/message"
)

type client struct {
	layer  *layer
	client *service.Client
	peer   id.Peer

	tunnel transfer.AuthTunnel

	// execRequests is only true for the `own` client
	// of `layer`. It will receive commands and process
	// them and also process status messages.
	execRequests bool

	// A unique client id, used for the mqtt id.
	clientIdx uint32

	// Last time we heard from our peer
	// (not only for ping, but for all operations)
	lastHearbeat time.Time
}

var GlobalClientIdx = uint32(0)

func newClient(lay *layer, tunnel transfer.AuthTunnel, peer id.Peer, execRequests bool) (*client, error) {
	idx := atomic.AddUint32(&GlobalClientIdx, 1)

	return &client{
		layer:        lay,
		client:       nil,
		execRequests: execRequests,
		peer:         peer,
		lastHearbeat: time.Now(),
		tunnel:       tunnel,
		clientIdx:    idx,
	}, nil
}

func (cv *client) peerTopic(sub string) []byte {
	return []byte(fmt.Sprintf("%s/%s", cv.peer.Hash(), sub))
}

func (cv *client) formatClientID() []byte {
	return []byte(fmt.Sprintf("%s%d", cv.peer.Hash(), cv.clientIdx))
}

func (cv *client) heartbeat(msg, ack message.Message, err error) error {
	if err != nil {
		return err
	}

	// BEAT IT, JUST BEAT IT! (Sorry, catchy tune.)
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

func (cv *client) statusMessage(status string) ([]byte, []byte) {
	data := []byte(fmt.Sprintf("%s:%s", status, cv.layer.self.ID()))
	topic := cv.peerTopic("status/" + cv.layer.self.Hash())
	return data, topic
}

func (cv *client) notifyStatus(status string) error {
	data, topic := cv.statusMessage(status)
	return cv.publish(data, topic)
}

func (cv *client) processRequest(msg *message.PublishMessage, answer bool) error {
	if !cv.execRequests {
		return nil
	}

	parts := bytes.SplitN(msg.Topic(), []byte{'/'}, 3)
	if len(parts) < 3 {
		return fmt.Errorf("Bad topic: %v", msg.Topic())
	}

	reqData := msg.Payload()
	req := &wire.Request{}

	if err := cv.payloadToProto(req, reqData); err != nil {
		return err
	}

	handler, ok := cv.layer.handlers[req.GetReqType()]
	if !ok {
		return fmt.Errorf("No such request handler: %d", req.GetReqType())
	}

	resp, err := handler(req)
	if err != nil {
		return err
	}

	if !answer || (resp == nil && err == nil) {
		return nil
	}

	// Respond error back if any:
	if resp == nil {
		resp = &wire.Response{
			Error: proto.String(err.Error()),
		}
	}

	// Autofill the internal fields:
	resp.ID = proto.Int64(req.GetID())
	resp.ReqType = req.GetReqType().Enum()

	respData, err := cv.protoToPayload(resp)
	if err != nil {
		log.Debugf("Invalid proto response: %v", err)
		return err
	}

	respTopic := fmt.Sprintf(
		"%s/response/%s",
		parts[2],
		cv.layer.self.Hash(),
	)

	// Publish response:
	if err := cv.publish(respData, []byte(respTopic)); err != nil {
		return err
	}

	return nil
}

func (cv *client) handleStatus(msg *message.PublishMessage) error {
	parsedTopic := bytes.SplitN(msg.Topic(), []byte{'/'}, 3)
	if len(parsedTopic) != 3 {
		return fmt.Errorf("Invalid online notification: %s", msg.Topic())
	}

	data := msg.Payload()
	parsedData := bytes.SplitN(data, []byte{':'}, 2)
	if len(parsedData) != 2 {
		return fmt.Errorf("Invalid online notification data: %s", data)
	}

	hash, status := string(parsedTopic[2]), string(parsedData[0])

	switch status {
	case "offline":
		// Remove the conversation from the tab and close it.
		// User will get a EOF on the next operation.
		if client, ok := cv.layer.tab[hash]; ok {
			if err := client.Close(); err != nil {
				log.Warningf("Could not close offline client: %v", err)
			}
			delete(cv.layer.tab, hash)
		}
	default:
		return fmt.Errorf("Invalid status message: %s", status)
	}

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
	if err := cv.payloadToProto(resp, msg.Payload()); err != nil {
		return err
	}

	// Send the response to the requesting client:
	if err := cv.layer.forwardResponse(resp); err != nil {
		log.Warningf("forward failed: %v", err)
	}

	return nil
}

func (cv *client) connect(conn net.Conn) error {
	cred, err := cv.layer.authMgr.Credentials(cv.peer)
	if err != nil {
		return err
	}

	msg := message.NewConnectMessage()
	msg.SetVersion(4)
	msg.SetCleanSession(true)
	msg.SetClientId(cv.formatClientID())
	msg.SetKeepAlive(300)

	// Set login credentials:
	msg.SetUsername([]byte(cv.layer.self.ID()))
	msg.SetPassword(cred)

	// Plan for our own death:
	statusData, statusTopic := cv.statusMessage("offline")
	msg.SetWillTopic(statusTopic)
	msg.SetWillMessage(statusData)
	msg.SetWillQos(1)

	client := &service.Client{}

	if err := client.ConnectOverConn(conn, msg); err != nil {
		return err
	}

	topicHandlers := map[string]func(msg *message.PublishMessage) error{
		"broadcast":  cv.handleBroadcast,
		"response/+": cv.handleResponse,
	}

	if cv.execRequests {
		topicHandlers["status/+"] = cv.handleStatus
		topicHandlers["request/+"] = cv.handleRequests
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

	statusData, statusTopic := cv.statusMessage("offline")
	pubmsg := message.NewPublishMessage()
	pubmsg.SetTopic(statusTopic)
	pubmsg.SetPayload(statusData)
	pubmsg.SetQoS(2)

	done := make(chan bool)

	err := cv.client.Publish(pubmsg, func(_, _ message.Message, _ error) error {
		// OnCompleteFunc is called once broker acknowledged the message.
		// Still, surgemq seems to a little bit of time before disconnecting
		// fully to prevent weird messages (about closed connections) in the log.
		time.Sleep(100 * time.Millisecond)
		done <- true
		return nil
	})

	if err != nil {
		log.Warningf("Could not publish status notify.")
		log.Warningf("Client might appear still online to others.")

		// Just in case:
		done <- false
	}

	delete(cv.layer.tab, cv.peer.Hash())

	// Give notifyStatus a short bit of time to finish
	timeout := time.NewTimer(2 * time.Second)

	select {
	case <-done:
		break
	case <-timeout.C:
		log.Warningf("Could not send offline notify: timeout")
	}

	cv.client.Disconnect()
	cv.client = nil
	return nil
}

func (cv *client) SendAsync(req *wire.Request, handler transfer.AsyncFunc) error {
	if cv.client == nil {
		return transfer.ErrOffline
	}

	respnotify := cv.layer.addReqRespPair(req)

	data, err := cv.protoToPayload(req)
	if err != nil {
		return err
	}

	// Start before publish to fix a very unlikely race.
	go func() {
		// Guard with a timeout to protect against too many go routines.
		ticker := time.NewTicker(30 * time.Second)

		select {
		case resp, ok := <-respnotify:
			if resp != nil && ok && handler != nil {
				handler(resp)
			}
		case <-ticker.C:
		}
	}()

	reqTopic := cv.peerTopic("request/" + cv.layer.self.Hash())
	log.Debugf("Publish request on %s", reqTopic)
	if err := cv.publish(data, reqTopic); err != nil {
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

func (cv *client) payloadToProto(msg proto.Message, data []byte) error {
	decryptData, err := cv.tunnel.Decrypt(data)
	if err != nil {
		return err
	}

	decompData, err := lz4.Decode(decryptData, decryptData)
	if err != nil {
		return err
	}

	if err := proto.Unmarshal(decompData, msg); err != nil {
		return err
	}

	return nil
}

func (cv *client) protoToPayload(msg proto.Message) ([]byte, error) {
	data, err := proto.Marshal(msg)
	if err != nil {
		return nil, err
	}

	compData, err := lz4.Encode(data, data)
	if err != nil {
		return nil, err
	}

	log.Debugf(
		"Compressed message from %.1fKB to %1.fKB (%.1f%%)",
		float64(len(data))/1024,
		float64(len(compData))/1024,
		float64(len(compData))/float64(len(data))*100,
	)

	return cv.tunnel.Encrypt(compData)
}
