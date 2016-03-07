package im

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	log "github.com/Sirupsen/logrus"
	colorlog "github.com/disorganizer/brig/util/log"
	"github.com/disorganizer/brig/util/testutil"
)

func init() {
	log.SetOutput(os.Stderr)

	// Only log the warning severity or above.
	log.SetLevel(log.DebugLevel)

	// Log pretty text
	log.SetFormatter(&colorlog.ColorfulLogFormatter{})
}

func TestClientPingPong(t *testing.T) {
	return // TODO

	clientAlice, err := NewDummyClient(AliceJid, AlicePwd)
	if err != nil {
		t.Errorf("Creating Alice' client failed: %v", err)
		return
	}

	clientBob, err := NewDummyClient(BobJid, BobPwd)
	if err != nil {
		t.Errorf("Creating Alice' client failed: %v", err)
		return
	}

	paths, err := MakeBuddies(clientAlice, clientBob)
	if err != nil {
		t.Errorf("Alice could not be a friend of bob: %v", err)
		return
	}

	defer testutil.Remover(t, paths...)

	done := make(chan bool)

	suffix := testutil.CreateDummyBuf(8 * 1024 * 1024)

	go func() {
		cnv, err := clientAlice.Dial(BobJid)
		if err != nil {
			t.Errorf("Dial: %v", err)
			return
		}

		for i := 0; !cnv.Ended() && i < 10; i++ {
			t.Logf("Alice: PING %d", i)
			data := []byte(fmt.Sprintf("PING %d - %v", i, suffix))

			if _, err := cnv.Write(data); err != nil {
				t.Errorf("alice: write failed: %v", err)
				return
			}

			msg, err := cnv.ReadMessage()
			t.Logf("Alice: RECV %d: %s/%v", i, msg, err)
			if err != nil {
				t.Errorf("alice: read failed: %v", err)
				return
			}

			if !bytes.Equal(msg, []byte(fmt.Sprintf("PONG %d - %v", i, suffix))) {
				t.Errorf("PING %d does not match PONG %d", i, i)
				return
			}
		}

		done <- true
	}()

	cnv := clientBob.Listen()
	t.Logf("Talking to %v", cnv.Jid)

	for i := 0; !cnv.Ended() && i < 10; i++ {
		msg, err := cnv.ReadMessage()
		t.Logf("Bob: RECV %d: %s/%v", i, msg, err)
		if err != nil {
			t.Errorf("bob: read failed: %v", err)
			return
		}

		if !bytes.Equal(msg, []byte(fmt.Sprintf("PING %d - %v", i, suffix))) {
			t.Errorf("PING %d does not match PONG %d", i, i)
			return
		}

		t.Logf("Bob: PONG %d", i)
		if _, err = cnv.Write([]byte(fmt.Sprintf("PONG %d - %v", i, suffix))); err != nil {
			t.Errorf("bob: write failed: %v", err)
			return
		}
	}

	<-done
	if err := cnv.Close(); err != nil {
		t.Errorf("bob: Close failed: %v", err)
	}
}

func TestLargeMessage(t *testing.T) {
	// return // TODO
	clientAlice, err := NewDummyClient(AliceJid, AlicePwd)
	if err != nil {
		t.Errorf("Creating Alice' client failed: %v", err)
		return
	}

	defer clientAlice.Close()

	clientBob, err := NewDummyClient(BobJid, BobPwd)
	if err != nil {
		t.Errorf("Creating Alice' client failed: %v", err)
		return
	}

	defer clientBob.Close()

	paths, err := MakeBuddies(clientAlice, clientBob)
	if err != nil {
		t.Errorf("Alice could not be a friend of bob: %v", err)
		return
	}

	defer testutil.Remover(t, paths...)

	N := 1 * 1024 * 1024
	sendBuf := testutil.CreateDummyBuf(int64(N))

	go func() {
		cnv, err := clientAlice.Dial(BobJid)
		if err != nil {
			t.Errorf("Dial: %v", err)
			return
		}

		n, err := cnv.Write(sendBuf)
		if err != nil {
			t.Errorf("Write failed: %v", err)
			return
		}

		if n != N {
			t.Errorf("Written size (%d) differs from source (%d)", n, N)
			return
		}
	}()

	cnv := clientBob.Listen()
	if cnv == nil {
		t.Errorf("Bob finished listening too early.")
		return
	}

	recvBuf, err := cnv.ReadMessage()
	if err != nil {
		fmt.Println("test recv err", err)
		t.Errorf("Receiving from alice failed: %v", err)
		return
	}

	if len(recvBuf) != N {
		t.Errorf("Received weird amount of bytes from alice: %d; should be %d", len(recvBuf), N)
		return
	}

	if !bytes.Equal(sendBuf, recvBuf) {
		t.Errorf("Big messages differ.")
		return
	}
}
