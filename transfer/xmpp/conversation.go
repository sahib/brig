package xmpp

import (
	"bytes"
	"fmt"
	"io"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/tsuibin/goxmpp2/xmpp"
	"golang.org/x/crypto/otr"
)

const (
	// ChunkSize is the maximum size of one chunk of data.
	// XMPP does not allow arbitary big messages, therefore
	// we need to distribute big writes over several messages.
	ChunkSize = 8 * 1024
)

var (
	// ErrTimeout happens when the partner could not be reached after Config.Timeout.
	ErrTimeout = fmt.Errorf("Timeout reached during OTR I/O")
)

// Conversation represents a point to point connection with a buddy.
// It can be used like a io.ReadWriter over network, encrypted via OTR.
type Conversation struct {
	sync.Mutex

	// Jid of your Conversation.
	Jid xmpp.JID

	// Client is a pointer to the client this cnv belongs to.
	Client *Client

	// recv provides all messages sent from this cnv.
	recv chan []byte

	// send can be used to send arbitrary messages to this cnv.
	send chan []byte

	// the underlying OTR conversation
	conversation *otr.Conversation

	// A backlog of messages send before OTR auth.
	backlog [][]byte

	// used in Read() to compensate against small read-buffers.
	readBuf *bytes.Buffer

	// This is set to a value > 0 if the conversation ended.
	isDead bool

	// Did we initiated the conversation to this cnv?
	initiated bool

	// This cnv completed the auth-game
	authenticated bool
}

func newConversation(ID xmpp.JID, client *Client, privKey *otr.PrivateKey) *Conversation {
	sendChan := make(chan []byte)
	recvChan := make(chan []byte)

	go func() {
		for data := range sendChan {
			if err := client.send(ID, data); err != nil {
				log.Warningf("im-send: %v", err)
			}
		}
	}()

	return &Conversation{
		Jid:     ID,
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

func escape(data []byte) []byte {
	buf := &bytes.Buffer{}

	for _, b := range data {
		switch b {
		case 0:
			buf.Write([]byte{'\\', '0'})
		case '\\':
			buf.Write([]byte{'\\', '\\'})
		default:
			buf.WriteByte(b)
		}
	}

	return buf.Bytes()
}

func unescape(data []byte) []byte {
	// TODO: I'm lazy, this could be more efficient...
	data = bytes.Replace(data, []byte{'\\', '\\'}, []byte{'\\'}, -1)
	data = bytes.Replace(data, []byte{'\\', '0'}, []byte{0}, -1)
	return data
}

// func (b *Conversation) Write(buf []byte) (int, error) {
// 	if len(buf) == 0 {
// 		return 0, nil
// 	}
//
// 	n := 0
// 	chunks := bytes.NewBuffer(buf)
//
// 	sizeBuf := make([]byte, 4)
// 	binary.LittleEndian.PutUint32(sizeBuf, uint32(len(buf)))
// 	// TODO: hash the whole data and send that hash too?
//
// 	fmt.Println("write header", sizeBuf)
// 	if _, err := b.writeChunk(sizeBuf); err != nil {
// 		return 0, err
// 	}
//
// 	for {
// 		chunk := chunks.Next(ChunkSize)
// 		if len(chunk) == 0 {
// 			break
// 		}
//
// 		nn, err := b.writeChunk(chunk)
// 		fmt.Println("write chunk", nn, err, chunk)
// 		if err != nil {
// 			return n, err
// 		}
//
// 		n += nn
// 	}
//
// 	return n, nil
// }

func (b *Conversation) Write(buf []byte) (int, error) {
	if b.Ended() {
		return 0, io.EOF
	}

	ticker := time.NewTicker(b.Client.Timeout)

	// This is retarted and the result of crack-misuse by the OTR designers.
	// Sending nul bytes as part of `buf` is not allowed, since the protocol
	// uses \0 to split off the TLV data (length, type etc.) - which goes
	// wrong when having nul bytes in the actual data. Why the heck didn't
	// they just prepend it to the data?
	select {
	case <-ticker.C:
		return 0, ErrTimeout
	case b.send <- escape(buf):
		return len(buf), nil
	}
}

func (b *Conversation) Read(buf []byte) (int, error) {
	msg, err := b.ReadMessage()
	if err != nil {
		return 0, err
	}

	b.Lock()
	defer b.Unlock()

	n, _ := b.readBuf.Write(msg)
	return b.readBuf.Read(buf[:n])
}

// func (b *Conversation) ReadMessage() ([]byte, error) {
// 	header, err := b.readChunk()
// 	if err != nil {
// 		fmt.Println("Reading header failed")
// 		return nil, err
// 	}
//
// 	fmt.Println("Read header size", header)
//
// 	if len(header) < 4 {
// 		fmt.Println("Weird header", header)
// 		return nil, fmt.Errorf("Bad header")
// 	}
//
// 	size := binary.LittleEndian.Uint32(header)
// 	buf := bytes.NewBuffer(make([]byte, 0, size))
// 	fmt.Println("Read header size in bytes", size)
//
// 	nchunks := int(size / ChunkSize)
// 	if size%ChunkSize != 0 {
// 		nchunks += 1
// 	}
//
// 	for i := 0; i < nchunks; i++ {
// 		chunk, err := b.readChunk()
//
// 		if chunk != nil {
// 			buf.Write(chunk)
// 		}
//
// 		if err != nil {
// 			return buf.Bytes(), err
// 		}
// 	}
//
// 	return buf.Bytes(), nil
// }

// ReadMessage returns exactly one message.
func (b *Conversation) ReadMessage() ([]byte, error) {
	if b.Ended() {
		return nil, io.EOF
	}

	ticker := time.NewTicker(b.Client.Timeout)

	select {
	case <-ticker.C:
		return nil, ErrTimeout
	case msg, ok := <-b.recv:
		if ok {
			// See comment in Write() for explanation.
			return unescape(msg), nil
		}

		return nil, io.EOF
	}
}

func (b *Conversation) adieu() {
	// Make sure Write()/Read() does not block anymore.
	b.Lock()
	defer b.Unlock()

	if b.isDead {
		return
	}

	b.isDead = true
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
	b.Lock()
	defer b.Unlock()

	if !b.isDead {
		b.recv <- msg
	}
}

// Ended returns true when the underlying conversation was ended.
func (b *Conversation) Ended() bool {
	b.Lock()
	defer b.Unlock()

	return b.isDead
}

// Close ends a conversation. You normally do not need to call this directly.
// There is no guarantee that previously send messages will be actually delivered.
func (b *Conversation) Close() error {
	if b.Ended() {
		return nil
	}

	b.adieu()
	b.Client.removeConversation(b.Jid)
	return nil
}
