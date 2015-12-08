package main

import (
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	xmpp "github.com/tsuibin/goxmpp2/xmpp"
	"golang.org/x/crypto/otr"
)

// TODO: Purge dead conversations.

// Client is an xmpp client with OTR support.
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

	// Map of JID to Conversation layer.
	conversations map[xmpp.JID]*otr.Conversation

	// Map of JID to wether we initiated the conversation.
	initiated map[xmpp.JID]bool

	// Map of JID to authorisation state.
	authorised map[xmpp.JID]bool

	blockOtr chan error

	Send chan<- xmpp.Stanza
	Recv <-chan xmpp.Stanza
}

// NewClient returns a ready client or nil on error.
func NewClient(jid xmpp.JID, password string) (*Client, error) {
	recvChan := make(chan xmpp.Stanza, 10)
	sendChan := make(chan xmpp.Stanza, 10)

	c := &Client{
		conversations: make(map[xmpp.JID]*otr.Conversation),
		authorised:    make(map[xmpp.JID]bool),
		KeyPath:       "/tmp/otr.key", // TODO
		initiated:     make(map[xmpp.JID]bool),
		blockOtr:      make(chan error, 1),
		Send:          sendChan,
		Recv:          recvChan,
	}

	// TODO: This tls config is probably a bad idea.
	innerClient, err := xmpp.NewClient(
		&jid,
		password,
		tls.Config{InsecureSkipVerify: true},
		nil,
		xmpp.Presence{},
		c.Status,
	)

	if err != nil {
		log.Fatalf("NewClient(%v): %v", jid, err)
		return nil, err
	}

	c.C = innerClient

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
				// fmt.Printf("--->\n%s: %s\n<---\n", msg.From, msg.Body[0].Chardata)
				if c.doRecv(msg) {
					recvChan <- msg
				}
			}
		}
	}()

	// Send loop: Send incoming messages over the network.
	go func() {
		for stanza := range sendChan {
			if msg, ok := stanza.(*xmpp.Message); ok {
				// TODO:  Join bodies.
				c.doSend(msg.To, msg.Body[0].Chardata)
			}
		}
	}()

	return c, nil
}

func loadPrivateKey(path string) (*otr.PrivateKey, error) {
	key := &otr.PrivateKey{}
	file, err := os.Open(path)

	// Generate a fresh one if it does not exist.
	if os.IsNotExist(err) {
		key.Generate(rand.Reader)

		if err := ioutil.WriteFile(path, key.Serialize(nil), 0600); err != nil {
			return key, err
		}

		log.Infof("Key Generated: %x", key.Serialize(nil))
		return key, nil
	}

	// There was some other error.
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	if _, ok := key.Parse(data); ok {
		return key, nil
	}

	return nil, fmt.Errorf("Parsing otr-key failed.")
}

func (c *Client) lookupConv(jid xmpp.JID) (*otr.Conversation, bool) {
	con, ok := c.conversations[jid]
	if !ok {
		fmt.Printf("NEW CONVERSATION: `%v`\n", string(jid))
		con = &otr.Conversation{}
		c.conversations[jid] = con
		c.authorised[jid] = false
		c.initiated[jid] = false

		privKey, err := loadPrivateKey(c.KeyPath)
		if err != nil {
			log.Errorf("otr-key-gen failed: %v", err)
		}

		con.PrivateKey = privKey
	}

	return con, !ok
}

func truncate(a string, l int) string {
	if len(a) > l {
		return a[:l] + "..." + a[len(a)-l:]
	}

	return a
}

func createMessage(from, to, text string) *xmpp.Message {
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

func (c *Client) doRecv(msg *xmpp.Message) bool {
	buf := &bytes.Buffer{}

	for _, field := range msg.Body {
		buf.Write([]byte(field.Chardata))
	}

	sendBack, wasNormal, err := c.recv(buf.Bytes(), msg.From)
	if err != nil {
		log.Warningf("recv failed: %v", err)
	}

	for _, outMsg := range sendBack {
		fmt.Printf("   SEND BACK: `%v`\n", truncate(string(outMsg), 20))
		c.C.Send <- createMessage(string(c.C.Jid), string(msg.From), string(outMsg))
	}

	return wasNormal
}

func (c *Client) recv(input []byte, from xmpp.JID) ([][]byte, bool, error) {
	con, isNew := c.lookupConv(from)
	if isNew {
		c.initiated[from] = false
	}

	weStarted := c.initiated[from]
	sendBack := [][]byte{}

	data, encrypted, state, toSend, err := con.Receive(input)
	if err != nil {
		fmt.Println("\n\n!!!!! ", err)
	}

	sendBack = append(sendBack, toSend...)

	wasNormal := encrypted

	fmt.Printf("RECV: `%v` `%v` (encr: %v %v %v) (state-change: %v)\n",
		truncate(string(data), 20),
		truncate(string(input), 20),
		encrypted,
		con.IsEncrypted(),
		c.authorised[from],
		state,
	)

	if state != otr.NoChange {
		wasNormal = false
	}

	switch state {
	case otr.NewKeys:
		if weStarted {
			authToSend, authErr := con.Authenticate("weis nich?", []byte("eule"))
			fmt.Println("==> AUTH REQUEST")
			if authErr != nil {
				fmt.Println(authErr)
			}
			sendBack = append(sendBack, authToSend...)
		}
	case otr.SMPSecretNeeded:
		question := con.SMPQuestion()
		fmt.Printf("[!] Answer a question '%s'\n", question)
		msgs, _ := con.Authenticate(question, []byte("eule"))
		sendBack = append(sendBack, msgs...)
	case otr.SMPComplete: // We completed their quest, ask partner now.
		fmt.Println("[!] Answer is correct")
		if weStarted == false && c.authorised[from] == false {
			authToSend, authErr := con.Authenticate("wer weis nich?", []byte("eule"))
			fmt.Println("==> AUTH REQUEST")
			if authErr != nil {
				fmt.Println(authErr)
			}
			sendBack = append(sendBack, authToSend...)
		}
		c.authorised[from] = true
		fmt.Println("BEFORE BLOCK")
		if weStarted {
			c.blockOtr <- nil
		}
		fmt.Println("AFTER BLOCK")
	case otr.SMPFailed:
		fmt.Println("[!] Answer is wrong")
		c.authorised[from] = false
		if weStarted {
			c.blockOtr <- nil
		}
	}

	return sendBack, wasNormal, nil
}

// Send sends `text` to participant `to`.
// A new otr session will be established if required.
func (c *Client) doSend(to xmpp.JID, text string) {
	con, isNew := c.lookupConv(to)
	if isNew {
		// Do the OTR dance first:
		c.initiated[to] = true
		c.send(to, otr.QueryMessage, con)

		// Wait till connection is fully established. TODO: timeout.
		if err := <-c.blockOtr; err != nil {
			log.Warningf("blockOtr: %v", err)
			// TODO: return err
		}
	}

	c.send(to, text, con)
}

func (c *Client) send(to xmpp.JID, text string, con *otr.Conversation) {
	base64Texts, err := con.Send([]byte(text))
	fmt.Printf("SEND(%v|%v): %v => %v\n",
		con.IsEncrypted(), c.authorised[to],
		text, truncate(string(base64Texts[0]), 20))

	if err != nil {
		fmt.Println("!! ", err)
		return
	}

	for _, base64Text := range base64Texts {
		c.C.Send <- createMessage(string(c.C.Jid), string(to), string(base64Text))
	}
}

// Close terminates all open connections.
func (client *Client) Close() {
	for jid, conversation := range client.conversations {
		fmt.Println("Closing OTR conversation to", jid)
		conversation.End()
	}
	client.C.Close()
}

func init() {
	// xmpp.Debug = true
}

// Verbindungsaufbau:
//   - Sender und Empfänger
//     wer weiß, wer wer ist?
//   - "Echte Nachrichten" zurück halten bis OTR auth fertig ist?
//   - Was passiert bei einem disconnect?

func main() {
	jid := flag.String("jid", "alice@jabber.nullcat.de/laptop", "JID to log in as")
	pwd := flag.String("pw", "", "password")
	to := flag.String("to", "bob@jabber.nullcat.de/desktop", "Receiver")
	send := flag.Bool("send", false, "Send otr query")

	flag.Parse()

	if *jid == "" || *pwd == "" {
		flag.Usage()
		os.Exit(2)
	}

	client, err := NewClient(xmpp.JID(*jid), *pwd)
	if err != nil {
		log.Fatalf("Could not create client: %v", err)
		return
	}

	defer client.Close()

	go func(ch <-chan xmpp.Stanza) {
		for stanza := range ch {
			if _, ok := stanza.(*xmpp.Message); ok {
				// fmt.Printf("--->\n%s: %s\n<---\n", msg.From, msg.Body[0].Chardata)
			}
		}
	}(client.Recv)

	sendOtr := true
	for {
		text := "Hello me. "
		if *send {
			if sendOtr {
				text += otr.QueryMessage
				sendOtr = false
			}
			client.Send <- createMessage(*jid, *to, text)
			time.Sleep(5 * time.Second)
		} else {
			fmt.Println(<-client.Recv)
			client.Send <- createMessage(*jid, *to, "PONG")
		}
	}
}
