package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/disorganizer/brig/transfer/xmpp"
	colorlog "github.com/disorganizer/brig/util/log"
	goxmpp "github.com/tsuibin/goxmpp2/xmpp"
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

	var ID goxmpp.JID
	var partnerJid goxmpp.JID
	var password string

	aliceJid := goxmpp.JID("alice@nullcat.de/laptop")
	bobJid := goxmpp.JID("bob@nullcat.de/desktop")

	if *sendFlag {
		ID, partnerJid, password = aliceJid, bobJid, "ThiuJ9wesh"
	} else {
		ID, partnerJid, password = bobJid, aliceJid, "eecot3oXan"
	}

	client, err := xmpp.NewClient(&xmpp.Config{
		Jid:             ID,
		Password:        password,
		TLSConfig:       tls.Config{ServerName: ID.Domain()},
		KeyPath:         "/tmp/otr.key." + password,
		FingerprintPath: "/tmp/otr.buddies." + password,
	})

	if err != nil {
		log.Fatalf("Could not create client: %v", err)
		return
	}

	defer client.Close()

	log.Infof("Partner is Online: %v", client.IsOnline(partnerJid))

	if *sendFlag {
		cnv, err := client.Dial(partnerJid)
		if err != nil {
			log.Errorf("Dial: %v", err)
			return
		}

		for i := 0; !cnv.Ended() && i < 10; i++ {
			log.Infof("Alice: PING %d", i)
			if _, err := cnv.Write([]byte(fmt.Sprintf("PING %d", i))); err != nil {
				log.Warningf("Alice: Write failed: %v", err)
				break
			}

			msg, err := cnv.ReadMessage()
			if err != nil {
				log.Warningf("Alice: ReadMessage failed: %v", err)
				break
			}

			log.Infof("Alice: RECV %d: %s/%v", i, msg, err)
			time.Sleep(2 * time.Second)
		}

		if err := cnv.Close(); err != nil {
			log.Warningf("Alice: Close failed: %v", err)
		}
	} else {
		for {
			cnv := client.Listen()

			log.Println("Dial to", cnv.Jid)
			go func() {
				for i := 0; !cnv.Ended() && i < 10; i++ {
					msg, err := cnv.ReadMessage()
					log.Infof("Bob: RECV %d: %s/%v", i, msg, err)
					log.Infof("Bob: PONG %d", i)
					if _, err := cnv.Write([]byte(fmt.Sprintf("PONG %d", i))); err != nil {
						log.Warningf("Bob: write failed: %v", err)
						break
					}
				}
			}()
		}
	}
}
