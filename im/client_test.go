package im

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
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

var (
	aliceJid   = xmpp.JID("alice@jabber.nullcat.de/laptop")
	bobJid     = xmpp.JID("bob@jabber.nullcat.de/desktop")
	alicePwd   = "ThiuJ9wesh"
	bobPwd     = "eecot3oXan"
	aliceKey   = filepath.Join(os.TempDir(), "otr.key.alice")
	bobKey     = filepath.Join(os.TempDir(), "otr.key.bob")
	buddyPathA = filepath.Join(os.TempDir(), "otr.test-buddies.alice")
	buddyPathB = filepath.Join(os.TempDir(), "otr.test-buddies.bob")
)

type Run struct {
	alice *Client
	bob   *Client
}

func writeDummyBuddies(t *testing.T, r *Run) {
	fa := r.alice.Fingerprint()
	fb := r.bob.Fingerprint()

	t.Log("FB: ", fb)

	aliceBuddies := fmt.Sprintf("%s: %s\n", r.bob.C.Jid, fb)
	bobBuddies := fmt.Sprintf("%s: %s\n", r.alice.C.Jid, fa)

	if err := ioutil.WriteFile(buddyPathA, []byte(aliceBuddies), 0644); err != nil {
		t.Errorf("Could not create %v: %v", buddyPathA)
	}

	if err := ioutil.WriteFile(buddyPathB, []byte(bobBuddies), 0644); err != nil {
		t.Errorf("Could not create %v: %v", buddyPathB)
	}
}

func clientPingPong(t *testing.T) {
	r := Run{}

	defer func() {
		for _, path := range []string{aliceKey, bobKey, buddyPathA, buddyPathB} {
			if err := os.Remove(path); err != nil {
				t.Logf("Note: could not remove %v", path)
			}
		}
	}()

	client, err := NewClient(&Config{
		Jid:                  aliceJid,
		Password:             alicePwd,
		TLSConfig:            tls.Config{ServerName: aliceJid.Domain()},
		KeyPath:              aliceKey,
		FingerprintStorePath: buddyPathA,
	})

	if err != nil {
		t.Errorf("Could not create alice client: %v", err)
		return
	}

	r.alice = client

	client, err = NewClient(&Config{
		Jid:                  bobJid,
		Password:             bobPwd,
		TLSConfig:            tls.Config{ServerName: bobJid.Domain()},
		KeyPath:              bobKey,
		FingerprintStorePath: buddyPathB,
	})

	if err != nil {
		t.Errorf("Could not create bob client: %v", err)
		return
	}

	r.bob = client

	writeDummyBuddies(t, &r)
	done := make(chan bool)

	go func() {
		cnv, err := r.alice.Dial(bobJid)
		if err != nil {
			t.Errorf("Dial: %v", err)
			return
		}

		for i := 0; !cnv.Ended() && i < 10; i++ {
			t.Logf("Alice: PING %d", i)
			cnv.Write([]byte(fmt.Sprintf("PING %d", i)))

			msg, err := cnv.ReadMessage()
			t.Logf("Alice: RECV %d: %s/%v", i, msg, err)
			time.Sleep(1 * time.Millisecond)
		}

		done <- true
	}()

	cnv := r.bob.Listen()
	t.Logf("Talking to %v", cnv.Jid)

	for i := 0; !cnv.Ended() && i < 10; i++ {
		msg, err := cnv.ReadMessage()
		t.Logf("Bob: RECV %d: %s/%v", i, msg, err)
		t.Logf("Bob: PONG %d", i)
		cnv.Write([]byte(fmt.Sprintf("PONG %d", i)))
	}

	<-done
}

func TestClientPingPong(t *testing.T) {
	clientPingPong(t)
}
