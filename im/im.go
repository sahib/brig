package im

import (
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	log "github.com/Sirupsen/logrus"
	xmpp "github.com/tsuibin/goxmpp2/xmpp"

	"golang.org/x/crypto/otr"
)

// Debug is a flag that enables some debug prints when set to true.
var Debug bool

func init() {
	Debug = false
	xmpp.Debug = false
}

// TODO: Prevent send to unavailable partner?
// TODO: Compare fingerprints. (store to file with key)
// TODO: Provide Config for Client

type Buddy struct {
	// Jid of your Buddy.
	Jid xmpp.JID

	// Client is a pointer to the client this buddy belongs to.
	Client *Client

	// Recv provides all messages sent from this buddy.
	Recv chan []byte

	// Send can be used to send arbitary messages to this buddy.
	Send chan []byte

	// We initiated the conversation to this buddy.
	initiated bool

	// This buddy completed the auth-game
	authorised bool

	// the underlying otr conversation
	conversation *otr.Conversation

	// A backlog of messages send before otr auth.
	backlog [][]byte

	// used in Read() to compensate against small read-buffers.
	readBuf *bytes.Buffer
}

func newBuddy(jid xmpp.JID, client *Client, privKey *otr.PrivateKey) *Buddy {
	sendChan := make(chan []byte)
	recvChan := make(chan []byte)

	go func() {
		for data := range sendChan {
			if err := client.send(jid, data); err != nil {
				log.Warningf("im-send: %v", err)
			}
		}
	}()

	return &Buddy{
		Jid:     jid,
		Client:  client,
		Recv:    recvChan,
		Send:    sendChan,
		backlog: make([][]byte, 0),
		readBuf: &bytes.Buffer{},
		conversation: &otr.Conversation{
			PrivateKey: privKey,
		},
	}
}

func (b *Buddy) Write(buf []byte) (int, error) {
	b.Send <- buf
	return len(buf), nil
}

func (b *Buddy) Read(buf []byte) (int, error) {
	msg := <-b.Recv
	b.readBuf.Write(msg)
	return b.readBuf.Read(buf)
}

func (b *Buddy) adieu() {
	b.authorised = false

	if b.conversation != nil {
		// End() returns some messages that can be used to revert the connection
		// back to a normal non-OTR connection. We just don't send those.
		b.conversation.End()
	}

	if b.Send != nil {
		close(b.Send)
		b.Send = nil
	}
}

// Client is an XMPP client with OTR support.
// Before establishing a connection, OTR will be triggered
// and the Socialist Millionaire Protocol is played through,
// using the minilock IDs of the participants.
type Client struct {
	sync.Mutex

	// Embedded client
	C *xmpp.Client

	// Path to a otr-key file. If empty, a new one will be generated.
	KeyPath string

	// Connection Status channel:
	Status chan xmpp.Status

	// JID to each individual buddy.
	// Only active connections are stored here.
	buddies map[xmpp.JID]*Buddy

	// buddies that send initial messages to us are pushed to this chan.
	incomingBuddies chan *Buddy
}

// locked buddy lookup
func (c *Client) lookupBuddy(jid xmpp.JID) (*Buddy, bool) {
	c.Lock()
	defer c.Unlock()

	buddy, ok := c.buddies[jid]
	return buddy, ok
}

func (c *Client) removeBuddy(jid xmpp.JID) {
	c.Lock()
	defer c.Unlock()

	if buddy, ok := c.buddies[jid]; ok {
		buddy.adieu()
	}

	delete(c.buddies, jid)
}

// NewClient returns a ready client or nil on error.
func NewClient(jid xmpp.JID, password, keyPath string) (*Client, error) {
	c := &Client{
		buddies:         make(map[xmpp.JID]*Buddy),
		incomingBuddies: make(chan *Buddy),
		KeyPath:         keyPath,
	}

	xmppClient, err := xmpp.NewClient(
		&jid,
		password,
		// TODO: This tls config is probably a bad idea.
		tls.Config{
			InsecureSkipVerify: true,
		},
		nil,
		xmpp.Presence{},
		c.Status,
	)

	if err != nil {
		log.Fatalf("NewClient(%v): %v", jid, err)
		return nil, err
	}

	c.C = xmppClient

	// Remember to update the status:
	go func() {
		for status := range c.Status {
			log.Debugf("connection status %d", status)
		}
	}()

	// Recv loop: Handle incoming messages, filter OTR.
	go func() {
		for stanza := range c.C.Recv {
			switch msg := stanza.(type) {
			case *xmpp.Message:
				response, err := c.recv(msg)
				if err != nil {
					log.Warningf("im-recv: %v", err)
				}

				if response != nil {
					if buddy, ok := c.lookupBuddy(msg.From); ok {
						fmt.Println("recv", joinBodies(response), response)
						buddy.Recv <- joinBodies(response)
					}
				}
			case *xmpp.Presence:
				if msg.Type == "unavailable" {
					if _, ok := c.lookupBuddy(msg.From); ok {
						log.Infof("Removed otr conversation with %v", msg.From)
						c.removeBuddy(msg.From)
					}
				}
			}
		}
	}()

	return c, nil
}

// Talk opens a conversation with another peer.
func (c *Client) Talk(jid xmpp.JID) (*Buddy, error) {
	if err := c.send(jid, nil); err != nil {
		return nil, err
	}

	if buddy, ok := c.lookupBuddy(jid); ok {
		return buddy, nil
	}

	return nil, nil
}

// Listen waits for new buddies that talk to us.
func (c *Client) Listen() *Buddy {
	return <-c.incomingBuddies
}

func genPrivateKey(key *otr.PrivateKey, path string) error {
	key.Generate(rand.Reader)
	keyDump := key.Serialize(nil)

	if err := ioutil.WriteFile(path, keyDump, 0600); err != nil {
		return err
	}

	log.Infof("Key Generated: %x", key.Serialize(nil))
	return nil
}

// loadPrivateKey generates a valid otr.PrivateKey.
// This function should never fail in normal cases since it
// will attempt to generate a new key and write it to path as fallback.
func loadPrivateKey(path string) (*otr.PrivateKey, error) {
	key := &otr.PrivateKey{}

	// Try to load an existing one:
	if file, err := os.Open(path); err == nil {
		if data, err := ioutil.ReadAll(file); err == nil {
			if _, ok := key.Parse(data); ok {
				return key, nil
			}
		}
	}

	// Generate a new one as fallback or initial case:
	if err := genPrivateKey(key, path); err != nil {
		return nil, err
	}

	return key, nil
}

// NOTE: This function has to be called with c.Lock() held!
func (c *Client) lookupOrInitBuddy(jid xmpp.JID) (*Buddy, bool, error) {
	_, ok := c.buddies[jid]

	if !ok {
		log.Infof("new otr-conversation: `%v`", string(jid))
		privKey, err := loadPrivateKey(c.KeyPath)
		if err != nil {
			log.Errorf("otr-key-gen failed: %v", err)
			return nil, false, err
		}

		c.buddies[jid] = newBuddy(jid, c, privKey)
	}

	return c.buddies[jid], !ok, nil
}

// TODO: move those to im/common.go
func truncate(a string, l int) string {
	if len(a) > l {
		return a[:l] + "..." + a[len(a)-l:]
	}

	return a
}

func createMessage(from, to xmpp.JID, text string) *xmpp.Message {
	xmsg := &xmpp.Message{}
	xmsg.From = from
	xmsg.To = to
	xmsg.Id = xmpp.NextId()

	xmsg.Type = "chat"
	xmsg.Lang = "en"
	xmsg.Body = []xmpp.Text{
		{
			XMLName:  xml.Name{Local: "body"},
			Chardata: text,
		},
	}

	return xmsg
}

func joinBodies(msg *xmpp.Message) []byte {
	if msg == nil {
		return nil
	}

	buf := &bytes.Buffer{}
	for _, field := range msg.Body {
		buf.Write([]byte(field.Chardata))
	}

	return buf.Bytes()
}

func (c *Client) recv(msg *xmpp.Message) (*xmpp.Message, error) {
	plain, responses, isNoOtrMsg, err := c.recvRaw(joinBodies(msg), msg.From)
	if err != nil {
		return nil, err
	}

	// Turn every fragment into a separate xmpp message:
	for _, outMsg := range responses {
		if Debug {
			fmt.Printf("  SEND BACK: %v\n", truncate(string(outMsg), 30))
		}
		c.C.Send <- createMessage(c.C.Jid, msg.From, string(outMsg))
	}

	response := createMessage(msg.From, c.C.Jid, string(plain))
	if isNoOtrMsg {
		return response, nil
	}

	return nil, nil
}

func (c *Client) recvRaw(input []byte, from xmpp.JID) ([]byte, [][]byte, bool, error) {
	buddy, isNew, err := c.lookupOrInitBuddy(from)
	if err != nil {
		return nil, nil, false, err
	}

	// We talk to this buddy the first time.
	if isNew {
		buddy.initiated = false
		c.incomingBuddies <- buddy

		// TODO: This does not seem to work reliable yet.
		// First received message should be the otr query. Validate.
		// if !bytes.Contains(input, []byte(otr.QueryMessage)) {
		// 	err := fmt.Errorf("First message was no OTR query: %v", truncate(string(input), 20))
		// 	return nil, nil, false, err
		// }
	}

	// Pipe input through the conversation:
	cnv := buddy.conversation
	data, encrypted, stateChange, responses, err := cnv.Receive(input)
	if err != nil {
		return nil, nil, false, err
	}

	if Debug {
		fmt.Printf("RECV: `%v` `%v` (encr: %v %v %v) (state-change: %v)\n",
			truncate(string(data), 30),
			truncate(string(input), 30),
			encrypted,
			cnv.IsEncrypted(),
			buddy.authorised,
			stateChange,
		)
	}

	auth := func(question string, answer []byte) error {
		authResp, err := cnv.Authenticate(question, answer)
		if err != nil {
			log.Warningf("im: Authentication error: %v", err)
			return err
		}

		responses = append(responses, authResp...)
		return nil
	}

	// Handle any otr conversation state change:
	switch stateChange {
	case otr.NewKeys: // We exchanged keys, channel is encrypted now.
		if buddy.initiated {
			if err := auth("weis nich?", []byte("eule")); err != nil {
				return nil, nil, false, err
			}
		}
	case otr.SMPSecretNeeded: // We received a question and have to answer.
		question := cnv.SMPQuestion()
		fmt.Printf("[!] Answer a question '%s'\n", question)
		if err := auth(question, []byte("eule")); err != nil {
			return nil, nil, false, err
		}
	case otr.SMPComplete: // We completed their quest, ask partner now.
		fmt.Println("[!] Answer is correct")
		if buddy.initiated == false && buddy.authorised == false {
			if err := auth("wer weis nich?", []byte("eule")); err != nil {
				return nil, nil, false, err
			}
		}

		if buddy.initiated == true && buddy.authorised {
			responses = append(responses, buddy.backlog...)
			buddy.backlog = make([][]byte, 0)
		}

		buddy.authorised = true
	case otr.SMPFailed: // We failed with our answer.
		fmt.Println("[!] Answer is wrong")
		buddy.authorised = false
	case otr.ConversationEnded:
		buddy.adieu()
		delete(c.buddies, buddy.Jid)
	}

	return data, responses, stateChange == otr.NoChange && encrypted, nil
}

// Send sends `text` to participant `to`.
// A new otr session will be established if required.
// It is allowed that `text` to be nil. This might trigger the otr exchange,
// but does not send any real messages.
func (c *Client) send(to xmpp.JID, text []byte) error {
	c.Lock()
	defer c.Unlock()

	buddy, isNew, err := c.lookupOrInitBuddy(to)
	if err != nil {
		return err
	}

	if isNew {
		buddy.initiated = true

		// Send the initial ?OTRv2? query:
		if err := c.sendRaw(to, []byte(otr.QueryMessage), buddy); err != nil {
			return fmt.Errorf("im: OTR Authentication failed: %v", err)
		}
	}

	if text == nil {
		return nil
	}

	if !buddy.authorised {
		buddy.backlog = append(buddy.backlog, text)
		return nil
	}

	return c.sendRaw(to, text, buddy)
}

func (c *Client) sendRaw(to xmpp.JID, text []byte, buddy *Buddy) error {
	base64Texts, err := buddy.conversation.Send(text)

	if Debug {
		fmt.Printf("SEND(%v|%v): %v => %v\n",
			buddy.conversation.IsEncrypted(), buddy.authorised,
			string(text), truncate(string(base64Texts[0]), 30),
		)
	}

	if err != nil {
		log.Warningf("im: send:", err)
		return err
	}

	for _, base64Text := range base64Texts {
		c.C.Send <- createMessage(c.C.Jid, to, string(base64Text))
	}

	return nil
}

// Close terminates all open connections.
func (c *Client) Close() {
	c.Lock()
	defer c.Unlock()

	for _, buddy := range c.buddies {
		buddy.adieu()
	}
	c.C.Close()
}
