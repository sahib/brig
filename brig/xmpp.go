package main

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	xmpp "github.com/tsuibin/goxmpp2/xmpp"
	"golang.org/x/crypto/otr"
)

type XMPPClient struct {
	// Embedded client
	C *xmpp.Client

	// Connection Status channel:
	Status chan xmpp.Status

	conversations map[xmpp.JID]*otr.Conversation
}

func NewXMPPClient(jid xmpp.JID, pw string) (*XMPPClient, error) {
	client := &XMPPClient{}
	client.conversations = make(map[xmpp.JID]*otr.Conversation)

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
			fmt.Println("connection status %d", status)
		}
	}()

	return client, nil
}

// TODO: This is only a dummy.
func loadPrivateKey() *otr.PrivateKey {
	key := &otr.PrivateKey{}
	key.Generate(rand.Reader)

	// if file, err := os.Open("/tmp/keyfile"); err != nil {
	// 	// Generate fresh one:
	// 	key.Generate(rand.Reader)
	// 	fmt.Println("Key Generated", key.Serialize(nil))

	// 	// Save for next time:
	// 	ioutil.WriteFile("/tmp/keyfile", key.Serialize(nil), 0775)
	// } else {
	// 	buffer := make([]byte, 4096)
	// 	n, _ := file.Read(buffer)
	// 	_, ok := key.Parse(buffer[:n])
	// 	fmt.Print("Key was cached: ")
	// 	if ok {
	// 		fmt.Println("Success!")
	// 	} else {
	// 		fmt.Println("Nope.")
	// 	}
	// }

	return key
}

func (client *XMPPClient) getConversation(jid xmpp.JID) *otr.Conversation {
	con, ok := client.conversations[jid]
	if !ok {
		con = &otr.Conversation{}
		con.PrivateKey = loadPrivateKey()
		client.conversations[jid] = con
	}

	return con
}

func truncate(a string, l int) string {
	if len(a) > l {
		return a[:l]
	}

	return a
}

func (client *XMPPClient) Recv(msg *xmpp.Message) {
	con := client.getConversation(msg.From)

	for _, data := range msg.Body {
		data, encrypted, state, toSend, err := con.Receive([]byte(data.Chardata))
		if err != nil {
			fmt.Println("\n\n!!!!! ", err)
		}

		fmt.Printf("RECV: `%v` (encr: %v %v) (state-change: %v)\n",
			truncate(string(data), 20), encrypted, con.IsEncrypted(), state)

		if state == otr.NewKeys {
			authToSend, authErr := con.Authenticate("weis nich?", []byte("eule"))
			fmt.Println("==> AUTH REQUEST")
			if authErr != nil {
				fmt.Println("============ AUTH ==========")
				fmt.Println(authErr)
				fmt.Println("============ AUTH ==========")
			}
			toSend = append(toSend, authToSend...)
		}

		for _, s := range toSend {
			fmt.Printf("   SEND BACK: `%v`\n", truncate(string(s), 20))
			// client.Send(msg.From, string(s))
			xmsg := xmpp.Message{}
			xmsg.From = client.C.Jid
			xmsg.To = msg.From
			xmsg.Id = xmpp.NextId()

			xmsg.Type = "chat"
			xmsg.Lang = "en"
			xmsg.Body = []xmpp.Text{{XMLName: xml.Name{Local: "body"}, Chardata: string(s)}}

			client.C.Send <- &xmsg
		}
	}
}

func (client *XMPPClient) Send(to xmpp.JID, text string) {
	// Encrypt the message via OTR
	// (TODO: Abort if not encrypted: con.IsEncrypted)
	con := client.getConversation(to)
	base64Texts, err := con.Send([]byte(text))

	if err != nil {
		fmt.Println("!! ", err)
		return
	}

	fmt.Printf("SEND(%v): %v %v\n", con.IsEncrypted(), text, truncate(string(base64Texts[0]), 20))

	for _, base64Text := range base64Texts {
		xmsg := xmpp.Message{}
		xmsg.From = client.C.Jid
		xmsg.To = to
		xmsg.Id = xmpp.NextId()

		xmsg.Type = "chat"
		xmsg.Lang = "en"
		xmsg.Body = []xmpp.Text{{XMLName: xml.Name{Local: "body"}, Chardata: string(base64Text)}}

		client.C.Send <- &xmsg
	}
}

func (client *XMPPClient) Close() {
	// TODO: Iterate over all conversations and close them
	for _, conversation := range client.conversations {
		conversation.End()
	}
	client.C.Close()
}

func init() {
	// xmpp.Debug = true
}

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

	client, err := NewXMPPClient(xmpp.JID(*jid), *pwd)
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
				// client.Send(msg.From, "pong!")
				// client.Send(msg.From, otr.QueryMessage)
			}
		}
	}(client.C.Recv)

	sendOtr := true
	for {
		if *send {
			if sendOtr {
				client.Send(xmpp.JID(*to), "Hello me. "+otr.QueryMessage)
				sendOtr = false
			} else {
				client.Send(xmpp.JID(*to), "Hello me. ")
			}
		}
		time.Sleep(5 * time.Second)
	}

	// stat := make(chan xmpp.Status)
	// go func() {
	// 	for s := range stat {
	// 		log.Printf("connection status %d", s)
	// 	}
	// }()

	// tlsConf := tls.Config{InsecureSkipVerify: true}
	// c, err := xmpp.NewClient(&jid, *pw, tlsConf, nil, xmpp.Presence{}, stat)
	// if err != nil {
	// 	log.Fatalf("NewClient(%v): %v", jid, err)
	// }
	// defer c.Close()

	// go func(ch <-chan xmpp.Stanza) {
	// 	for obj := range ch {
	// 		fmt.Printf("s: %v\n", obj)
	// 	}
	// 	fmt.Println("done reading")
	// }(c.Recv)

	//msg := createMessage("sahib@jabber.nullcat.de/xxx", "christoph@jabber.nullcat.de/xxx", "Hello Kitteh")
	//c.Send <- msg

	// roster := c.Roster.Get()
	// fmt.Printf("%d roster entries:\n", len(roster))
	// for i, entry := range roster {
	// 	fmt.Printf("%d: %v %v %v\n", i, entry.Jid, entry.Name, entry.Subscription)
	// }

	// p := make([]byte, 1024)
	// for {
	// 	nr, _ := os.Stdin.Read(p)
	// 	if nr == 0 {
	// 		break
	// 	}
	// 	s := string(p)
	// 	dec := xml.NewDecoder(strings.NewReader(s))
	// 	t, err := dec.Token()
	// 	if err != nil {
	// 		fmt.Printf("token: %s\n", err)
	// 		break
	// 	}
	// 	var se *xml.StartElement
	// 	var ok bool
	// 	if se, ok = t.(*xml.StartElement); !ok {
	// 		fmt.Println("Couldn't find start element")
	// 		break
	// 	}
	// 	var stan xmpp.Stanza
	// 	switch se.Name.Local {
	// 	case "iq":
	// 		stan = &xmpp.Iq{}
	// 	case "message":
	// 		stan = &xmpp.Message{}
	// 	case "presence":
	// 		stan = &xmpp.Presence{}
	// 	default:
	// 		fmt.Println("Can't parse non-stanza.")
	// 		continue
	// 	}
	// 	err = dec.Decode(stan)
	// 	if err == nil {
	// 		c.Send <- stan
	// 	} else {
	// 		fmt.Printf("Parse error: %v\n", err)
	// 		break
	// 	}
	// }
	// fmt.Println("done sending")
}
