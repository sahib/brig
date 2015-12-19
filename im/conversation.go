package im

import (
	"bytes"
	"fmt"
	"sync/atomic"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/tsuibin/goxmpp2/xmpp"
	"golang.org/x/crypto/otr"
)

var (
	// ErrTimeout happens when the partner could not be reached after Config.Timeout.
	ErrTimeout = fmt.Errorf("Timeout reached during OTR io")
	// ErrDeadConversation happens when the underlying OTR conversation was ended.
	ErrDeadConversation = fmt.Errorf("Conversation ended already")
)

// Conversation represents a point to point connection with a buddy.
// It can be used like a io.ReadWriter over network, encrypted via OTR.
type Conversation struct {
	// Jid of your Conversation.
	Jid xmpp.JID

	// Client is a pointer to the client this cnv belongs to.
	Client *Client

	// recv provides all messages sent from this cnv.
	recv chan []byte

	// send can be used to send arbitary messages to this cnv.
	send chan []byte

	// the underlying otr conversation
	conversation *otr.Conversation

	// A backlog of messages send before otr auth.
	backlog [][]byte

	// used in Read() to compensate against small read-buffers.
	readBuf *bytes.Buffer

	// This is set to a value > 0 if the conversation ended.
	cnvIsDead uint32

	// Did we initiated the conversation to this cnv?
	initiated bool

	// This cnv completed the auth-game
	authenticated bool
}

func newConversation(jid xmpp.JID, client *Client, privKey *otr.PrivateKey) *Conversation {
	sendChan := make(chan []byte)
	recvChan := make(chan []byte)

	go func() {
		for data := range sendChan {
			if err := client.send(jid, data); err != nil {
				log.Warningf("im-send: %v", err)
			}
		}
	}()

	return &Conversation{
		Jid:     jid,
		Client:  client,
		recv:    recvChan,
		send:    sendChan,
		backlog: make([][]byte, 0),
		readBuf: &bytes.Buffer{},
		conversation: &otr.Conversation{
			PrivateKey: privKey,
		},
	}
}

func (b *Conversation) Write(buf []byte) (int, error) {
	if b.Ended() {
		return 0, ErrDeadConversation
	}

	ticker := time.NewTicker(b.Client.Timeout)

	select {
	case <-ticker.C:
		return 0, ErrTimeout
	case b.send <- buf:
		return len(buf), nil
	}
}

func (b *Conversation) Read(buf []byte) (int, error) {
	msg, err := b.ReadMessage()
	if err != nil {
		return 0, err
	}

	n, _ := b.readBuf.Write(msg)
	return b.readBuf.Read(buf[:n])
}

// ReadMessage returns exactly one message.
func (b *Conversation) ReadMessage() ([]byte, error) {
	if b.Ended() {
		return nil, ErrDeadConversation
	}

	ticker := time.NewTicker(b.Client.Timeout)

	select {
	case <-ticker.C:
		return nil, ErrTimeout
	case msg, ok := <-b.recv:
		if ok {
			return msg, nil
		}

		return nil, ErrDeadConversation
	}
}

// NOTE: adieu() is called with c.Lock() hold.
func (b *Conversation) adieu() {
	// Make sure Write()/Read() does not block anymore.
	atomic.StoreUint32(&b.cnvIsDead, 1)

	b.authenticated = false

	if b.conversation != nil {
		// End() returns some messages that can be used to revert the connection
		// back to a normal non-OTR connection. We just don't send those.
		b.conversation.End()
	}

	// Wakeup any Write/Read calls.
	close(b.send)
	close(b.recv)
}

// Add a message to the conversation
func (b *Conversation) add(msg []byte) {
	if !b.Ended() {
		b.recv <- msg
	}
}

// Ended returns true when the underlying conversation was ended.
func (b *Conversation) Ended() bool {
	return atomic.LoadUint32(&b.cnvIsDead) > 0
}
