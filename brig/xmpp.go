package main

import (
	"crypto/tls"
	"encoding/xml"
	"flag"
	"fmt"
	xmpp "github.com/tsuibin/goxmpp2/xmpp"
	"log"
	"os"
	"strings"
)

func createMessage(from, to xmpp.JID, text string) *xmpp.Message {
	xmsg := xmpp.Message{}
	xmsg.From = from
	xmsg.To = to
	xmsg.Id = xmpp.NextId()

	xmsg.Type = "chat"
	xmsg.Lang = "en"
	xmsg.Body = []xmpp.Text{{XMLName: xml.Name{Local: "body"}, Chardata: text}}

	// TODO: omitted:
	// xmpp.Thread
	return &xmsg
}

func init() {
	// xmpp.Debug = true
}

// Demonstrate the API, and allow the user to interact with an XMPP
// server via the terminal.
func main() {
	jidStr := flag.String("jid", "", "JID to log in as")
	pw := flag.String("pw", "", "password")
	flag.Parse()
	jid := xmpp.JID(*jidStr)
	if jid.Domain() == "" || *pw == "" {
		flag.Usage()
		os.Exit(2)
	}

	stat := make(chan xmpp.Status)
	go func() {
		for s := range stat {
			log.Printf("connection status %d", s)
		}
	}()

	tlsConf := tls.Config{InsecureSkipVerify: true}
	c, err := xmpp.NewClient(&jid, *pw, tlsConf, nil, xmpp.Presence{}, stat)
	if err != nil {
		log.Fatalf("NewClient(%v): %v", jid, err)
	}
	defer c.Close()

	go func(ch <-chan xmpp.Stanza) {
		for obj := range ch {
			fmt.Printf("s: %v\n", obj)
		}
		fmt.Println("done reading")
	}(c.Recv)

	msg := createMessage("sahib@jabber.nullcat.de/home", "christoph@jabber.nullcat.de/home", "Hello Kitteh")
	c.Send <- msg

	roster := c.Roster.Get()
	fmt.Printf("%d roster entries:\n", len(roster))
	for i, entry := range roster {
		fmt.Printf("%d: %v %v %v\n", i, entry.Jid, entry.Name, entry.Subscription)
	}

	p := make([]byte, 1024)
	for {
		nr, _ := os.Stdin.Read(p)
		if nr == 0 {
			break
		}
		s := string(p)
		dec := xml.NewDecoder(strings.NewReader(s))
		t, err := dec.Token()
		if err != nil {
			fmt.Printf("token: %s\n", err)
			break
		}
		var se *xml.StartElement
		var ok bool
		if se, ok = t.(*xml.StartElement); !ok {
			fmt.Println("Couldn't find start element")
			break
		}
		var stan xmpp.Stanza
		switch se.Name.Local {
		case "iq":
			stan = &xmpp.Iq{}
		case "message":
			stan = &xmpp.Message{}
		case "presence":
			stan = &xmpp.Presence{}
		default:
			fmt.Println("Can't parse non-stanza.")
			continue
		}
		err = dec.Decode(stan)
		if err == nil {
			c.Send <- stan
		} else {
			fmt.Printf("Parse error: %v\n", err)
			break
		}
	}
	fmt.Println("done sending")
}
