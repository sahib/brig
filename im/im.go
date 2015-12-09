package im

import (
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	xmpp "github.com/tsuibin/goxmpp2/xmpp"
	"golang.org/x/crypto/otr"
)

func init() {
	xmpp.Debug = false
}

// TODO: Purge dead conversations.
// TODO: Make Client be a io.ReadWriter?

type buddyInfo struct {
	// We initiated the conversation to this buddy.
	initiated bool

	// This buddy completed the auth-game
	authorised bool

	// the underlying otr conversation
	conversation *otr.Conversation
}

func (b *buddyInfo) Adieu() {
	if b.conversation != nil {
		b.conversation.End()
	}
}

// Client is an XMPP client with OTR support.
// Before establishing a connection, OTR will be triggered
// and the Socialist Millionaire Protocol is played through,
// using the minilock IDs of the participants.
type Client struct {
	// Embedded client
	C *xmpp.Client

	// Path to a otr-key file. If empty, a new one will be generated.
	KeyPath string

	// Connection Status channel:
	Status chan xmpp.Status

	buddies map[xmpp.JID]*buddyInfo

	blockOtr chan error
	Send     chan<- xmpp.Stanza
	Recv     <-chan xmpp.Stanza
}

// NewClient returns a ready client or nil on error.
func NewClient(jid xmpp.JID, password string) (*Client, error) {
	recvChan := make(chan xmpp.Stanza, 10)
	sendChan := make(chan xmpp.Stanza, 10)

	c := &Client{
		buddies:  make(map[xmpp.JID]*buddyInfo),
		blockOtr: make(chan error, 1),
		KeyPath:  "/tmp/otr.key", // TODO
		Send:     sendChan,
		Recv:     recvChan,
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
			fmt.Printf("connection status %d\n", status)
		}
	}()

	// Recv loop: Handle incoming messages, filter OTR.
	go func() {
		for stanza := range c.C.Recv {
			if msg, ok := stanza.(*xmpp.Message); ok {
				isNoOtrMsg, err := c.recv(msg)
				if err != nil {
					log.Warningf("im-recv: %v", err)
				}

				if isNoOtrMsg {
					recvChan <- msg
				}
			}
		}
	}()

	// Send loop: Send incoming messages over the network.
	go func() {
		for stanza := range sendChan {
			if msg, ok := stanza.(*xmpp.Message); ok {
				// TODO:  Join bodies, check err.
				c.send(msg.To, []byte(msg.Body[0].Chardata))
			}
		}
	}()

	return c, nil
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

func (c *Client) lookupBuddy(jid xmpp.JID) (*buddyInfo, bool, error) {
	_, ok := c.buddies[jid]

	if !ok {
		fmt.Printf("NEW CONVERSATION: `%v`\n", string(jid))
		privKey, err := loadPrivateKey(c.KeyPath)
		if err != nil {
			log.Errorf("otr-key-gen failed: %v", err)
			return nil, false, err
		}

		c.buddies[jid] = &buddyInfo{
			conversation: &otr.Conversation{
				PrivateKey: privKey,
			},
		}
	}

	return c.buddies[jid], !ok, nil
}

// TODO: debug, remove.
func truncate(a string, l int) string {
	if len(a) > l {
		return a[:l] + "..." + a[len(a)-l:]
	}

	return a
}

func CreateMessage(from, to, text string) *xmpp.Message {
	xmsg := &xmpp.Message{}
	xmsg.From = xmpp.JID(from)
	xmsg.To = xmpp.JID(to)
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

func (c *Client) recv(msg *xmpp.Message) (bool, error) {
	buf := &bytes.Buffer{}
	for _, field := range msg.Body {
		buf.Write([]byte(field.Chardata))
	}

	sendBack, isNoOtrMsg, err := c.recvRaw(buf.Bytes(), msg.From)
	if err != nil {
		log.Warningf("recv failed: %v", err)
		return false, err
	}

	// Turn every fragment into a separate xmpp message:
	for _, outMsg := range sendBack {
		fmt.Printf("   SEND BACK: `%v`\n", truncate(string(outMsg), 20))
		c.C.Send <- CreateMessage(string(c.C.Jid), string(msg.From), string(outMsg))
	}

	return isNoOtrMsg, nil
}

func (c *Client) recvRaw(input []byte, from xmpp.JID) ([][]byte, bool, error) {
	buddy, isNew, err := c.lookupBuddy(from)
	if err != nil {
		return nil, false, err
	}

	// We talk to this buddy the first time.
	if isNew {
		buddy.initiated = false

		// First message should be the otr query. Validate.
		if !bytes.Contains(input, []byte(otr.QueryMessage)) {
			err := fmt.Errorf("First message was no OTR query.")
			c.blockOtr <- err
			return nil, false, err
		}
	}

	// Pipe input through the conversation:
	cnv := buddy.conversation
	data, encrypted, state, response, err := cnv.Receive(input)
	if err != nil {
		fmt.Println("\n\n!!!!! ", err)
		c.blockOtr <- err
		return nil, false, err
	}

	fmt.Printf("RECV: `%v` `%v` (encr: %v %v %v) (state-change: %v)\n",
		truncate(string(data), 20),
		truncate(string(input), 20),
		encrypted,
		cnv.IsEncrypted(),
		buddy.authorised,
		state,
	)

	sendBack := [][]byte{}
	sendBack = append(sendBack, response...)

	auth := func(question string, answer []byte) error {
		authToSend, err := cnv.Authenticate(question, answer)
		fmt.Println("==> AUTH REQUEST")
		if err != nil {
			fmt.Println(err)
			c.blockOtr <- err
			return err
		}

		sendBack = append(sendBack, authToSend...)
		return nil
	}

	// Handle any otr conversation state change:
	switch state {
	case otr.NewKeys: // We exchanged keys, channel is encrypted now.
		if buddy.initiated {
			if err := auth("weis nich?", []byte("eule")); err != nil {
				return nil, false, err
			}
		}
	case otr.SMPSecretNeeded: // We received a question and have to answer.
		question := cnv.SMPQuestion()
		fmt.Printf("[!] Answer a question '%s'\n", question)
		if err := auth(question, []byte("eule")); err != nil {
			return nil, false, err
		}
	case otr.SMPComplete: // We completed their quest, ask partner now.
		fmt.Println("[!] Answer is correct")
		if buddy.initiated == false && buddy.authorised == false {
			if err := auth("wer weis nich?", []byte("eule")); err != nil {
				return nil, false, err
			}
		}

		buddy.authorised = true
		fmt.Println("BEFORE BLOCK")
		if buddy.initiated {
			c.blockOtr <- nil
		}
		fmt.Println("AFTER BLOCK")
	case otr.SMPFailed: // We failed with our answer.
		fmt.Println("[!] Answer is wrong")
		buddy.authorised = false
		if buddy.initiated {
			c.blockOtr <- nil
		}
	}

	return sendBack, state == otr.NoChange && encrypted, nil
}

// Send sends `text` to participant `to`.
// A new otr session will be established if required.
func (c *Client) send(to xmpp.JID, text []byte) error {
	buddy, isNew, err := c.lookupBuddy(to)
	if err != nil {
		return err
	}

	if isNew {
		// Do the OTR dance first:
		buddy.initiated = true
		if err := c.sendRaw(to, []byte(otr.QueryMessage), buddy); err != nil {
			return err
		}

		timeout := 1 * time.Minute
		ticker := time.NewTicker(timeout)

		// Wait until the otr connection is established:
		select {
		case <-ticker.C:
			return fmt.Errorf("OTR init took too long: %v", timeout)
		case err := <-c.blockOtr:
			if err != nil {
				log.Warningf("blockOtr: %v", err)
				return err
			}
		}
	}

	return c.sendRaw(to, text, buddy)
}

func (c *Client) sendRaw(to xmpp.JID, text []byte, buddy *buddyInfo) error {
	base64Texts, err := buddy.conversation.Send(text)

	// TODO: DEBUG
	fmt.Printf("SEND(%v|%v): %v => %v\n",
		buddy.conversation.IsEncrypted(), buddy.authorised,
		text, truncate(string(base64Texts[0]), 20))

	if err != nil {
		fmt.Println("!! ", err)
		return err
	}

	for _, base64Text := range base64Texts {
		c.C.Send <- CreateMessage(string(c.C.Jid), string(to), string(base64Text))
	}

	return nil
}

// Close terminates all open connections.
func (c *Client) Close() {
	for jid, buddy := range c.buddies {
		fmt.Println("Closing OTR conversation to", jid)
		buddy.Adieu()
	}
	c.C.Close()
}
