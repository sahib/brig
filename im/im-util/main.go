package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/im"
	colorlog "github.com/disorganizer/brig/util/log"
	"github.com/tsuibin/goxmpp2/xmpp"
)

func init() {
	log.SetOutput(os.Stderr)

	// Only log the warning severity or above.
	log.SetLevel(log.DebugLevel)

	// Log pretty text
	log.SetFormatter(&colorlog.ColorfulLogFormatter{})
}

func main() {
	sendFlag := flag.Bool("send", false, "Send otr query")

	flag.Parse()

	var jid xmpp.JID
	var partnerJid xmpp.JID
	var password string

	aliceJid := xmpp.JID("alice@jabber.nullcat.de/laptop")
	bobJid := xmpp.JID("bob@jabber.nullcat.de/desktop")

	if *sendFlag {
		jid, partnerJid, password = aliceJid, bobJid, "ThiuJ9wesh"
	} else {
		jid, partnerJid, password = bobJid, aliceJid, "eecot3oXan"
	}

	client, err := im.NewClient(&im.Config{
		Jid:          jid,
		KeyPath:      "/tmp/otr.key",
		KeyStorePath: "/tmp/otr.buddies",
		Password:     password,
		TLSConfig:    tls.Config{ServerName: jid.Domain()},
	})

	if err != nil {
		log.Fatalf("Could not create client: %v", err)
		return
	}

	defer client.Close()

	if *sendFlag {
		cnv, err := client.Talk(partnerJid)
		if err != nil {
			log.Errorf("Talk: %v", err)
			return
		}

		for i := 0; !cnv.Ended(); i++ {
			log.Println("Alice: PING")
			cnv.Write([]byte(fmt.Sprintf("PING %d", i)))
			log.Println("Alice: RECV")
			fmt.Println(cnv.ReadMessage())
			time.Sleep(2 * time.Second)
		}
	} else {
		for {
			cnv := client.Listen()
			log.Println("Talking to", cnv.Jid)
			go func() {
				for i := 0; !cnv.Ended(); i++ {
					log.Println("Bob: RECV")
					fmt.Println(cnv.ReadMessage())
					log.Println("Bob: PONG")
					cnv.Write([]byte(fmt.Sprintf("PONG %d", i)))
				}
			}()
		}
	}
}
