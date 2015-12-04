package main

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	xmpp "github.com/tsuibin/goxmpp2/xmpp"
	"golang.org/x/crypto/otr"
)

// Client is an xmpp client with OTR support.
// Before establishing a connection, OTR will be triggered
// and the Socialist Millionaire Protocol is played through,
// using the minilock IDs of the participants.
type Client struct {
	// Embedded client
	C *xmpp.Client

	// Connection Status channel:
	Status chan xmpp.Status

	// Map of JID to Conversation layer
	conversations map[xmpp.JID]*otr.Conversation

	// Map of JID to authorisation state
	authorised map[xmpp.JID]bool
}

// NewClient returns a ready client or nil on error.
func NewClient(jid xmpp.JID, pw string) (*Client, error) {
	client := &Client{}
	client.conversations = make(map[xmpp.JID]*otr.Conversation)
	client.authorised = make(map[xmpp.JID]bool)

	// TODO: This tls config is probably a bad idea.
	tlsConf := tls.Config{InsecureSkipVerify: true}
	innerClient, err := xmpp.NewClient(&jid, pw,
		tlsConf, nil, xmpp.Presence{}, client.Status)

	if err != nil {
		log.Fatalf("NewClient(%v): %v", jid, err)
		return nil, err
	}

	client.C = innerClient

	// Remember to update the status:
	go func() {
		for status := range client.Status {
			fmt.Printf("connection status %d\n", status)
		}
	}()

	return client, nil
}

// TODO: This is only a dummy.
func loadPrivateKey() *otr.PrivateKey {
	key := &otr.PrivateKey{}

	if file, err := os.Open("/tmp/keyfile"); err != nil {
		// Generate fresh one:
		key.Generate(rand.Reader)
		fmt.Printf("Key Generated: %x\n", key.Serialize(nil))

		// Save for next time:
		ioutil.WriteFile("/tmp/keyfile", key.Serialize(nil), 0775)
	} else {
		// TODO: This *seems* to work, assert it does.
		buffer := make([]byte, 4096)
		n, _ := file.Read(buffer)
		_, ok := key.Parse(buffer[:n])
		fmt.Print("Key was cached: ")
		if ok {
			fmt.Println("Success!")
		} else {
			fmt.Println("Nope.")
		}
	}

	return key
}

func (client *Client) initOtr(jid xmpp.JID) {
	client.Send(jid, otr.QueryMessage)
}

func (client *Client) getConversation(jid xmpp.JID) (*otr.Conversation, bool) {
	con, ok := client.conversations[jid]
	if !ok {
		fmt.Printf("NEW CONVERSATION: `%v`\n", string(jid))
		con = &otr.Conversation{}
		con.PrivateKey = loadPrivateKey()
		client.conversations[jid] = con
		client.authorised[jid] = false
	}

	return con, !ok
}

func truncate(a string, l int) string {
	if len(a) > l {
		return a[:l] + "..." + a[len(a)-l:]
	}

	return a
}

func createMessage(from, to string, text []byte) *xmpp.Message {
	xmsg := &xmpp.Message{}
	xmsg.From = xmpp.JID(from)
	xmsg.To = xmpp.JID(to)
	xmsg.Id = xmpp.NextId()

	xmsg.Type = "chat"
	xmsg.Lang = "en"
	xmsg.Body = []xmpp.Text{{XMLName: xml.Name{Local: "body"}, Chardata: string(text)}}

	return xmsg
}

var isServer = false

// Recv receives a single xmpp.Message which is written to *msg.
func (client *Client) Recv(msg *xmpp.Message) {
	con, _ := client.getConversation(msg.From)
	sendBack := [][]byte{}

	for _, field := range msg.Body {
		data, encrypted, state, toSend, err := con.Receive([]byte(field.Chardata))
		if err != nil {
			fmt.Println("\n\n!!!!! ", err)
		}

		sendBack = append(sendBack, toSend...)

		fmt.Printf("RECV: `%v` `%v` (encr: %v %v %v) (state-change: %v)\n",
			truncate(string(data), 20),
			truncate(string(field.Chardata), 20),
			encrypted, con.IsEncrypted(), client.authorised[msg.From], state)

		switch state {
		case otr.NewKeys:
			if isServer {
				authToSend, authErr := con.Authenticate("weis nich?", []byte("eule"))
				fmt.Println("==> AUTH REQUEST")
				if authErr != nil {
					fmt.Println("============ AUTH ==========")
					fmt.Println(authErr)
					fmt.Println("============ AUTH ==========")
				}
				sendBack = append(sendBack, authToSend...)
			}
		case otr.SMPSecretNeeded:
			question := con.SMPQuestion()
			fmt.Printf("[!] Answer a question '%s'\n", question)
			msgs, _ := con.Authenticate(question, []byte("eule"))
			sendBack = append(sendBack, msgs...)
		case otr.SMPComplete:
			fmt.Println("[!] Answer is correct")
			if isServer == false && client.authorised[msg.From] == false {
				authToSend, authErr := con.Authenticate("wer weis nich?", []byte("eule"))
				fmt.Println("==> AUTH REQUEST")
				if authErr != nil {
					fmt.Println("============ AUTH ==========")
					fmt.Println(authErr)
					fmt.Println("============ AUTH ==========")
				}
				sendBack = append(sendBack, authToSend...)
			}
			client.authorised[msg.From] = true
		case otr.SMPFailed:
			fmt.Println("[!] Answer is wrong")
			client.authorised[msg.From] = false
		}

		for _, s := range sendBack {
			fmt.Printf("   SEND(%v) BACK: `%v`\n", con.IsEncrypted(), truncate(string(s), 20))
			client.C.Send <- createMessage(string(client.C.Jid), string(msg.From), s)
		}
	}
}

// Send sends `text` to participant `to`.
// A new otr session will be established if required.
func (client *Client) Send(to xmpp.JID, text string) {
	con, wasNew := client.getConversation(to)

	if wasNew {
		client.initOtr(to)
	}

	base64Texts, err := con.Send([]byte(text))
	fmt.Printf("SEND(%v|%v): %v => %v\n",
		con.IsEncrypted(), client.authorised[to],
		text, truncate(string(base64Texts[0]), 20))

	if err != nil {
		fmt.Println("!! ", err)
		return
	}

	for _, base64Text := range base64Texts {
		client.C.Send <- createMessage(string(client.C.Jid), string(to), base64Text)
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
	jid := flag.String("jid", "", "JID to log in as")
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
			if msg, ok := stanza.(*xmpp.Message); ok {
				// fmt.Printf("--->\n%s: %s\n<---\n", msg.From, msg.Body[0].Chardata)
				client.Recv(msg)
			}
		}
	}(client.C.Recv)

	sendOtr := true
	for {
		if *send {
			isServer = true
			if sendOtr {
				client.Send(xmpp.JID(*to), "Hello me. "+otr.QueryMessage)
				sendOtr = false
			} else {
				client.Send(xmpp.JID(*to), "Hello me.")
			}
		}
		time.Sleep(5 * time.Second)
	}
}
