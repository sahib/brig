package websocket

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"testing"
	"testing/iotest"

	utils "gx/ipfs/QmW5LxJm2Yo3S7uVsfLM7NsJn2QnKbvZD7uYsZVYR7YViE/go-libp2p-transport/test"
	insecure "gx/ipfs/QmYMGCwtKbLXfh2MUTbLoP66N9Y24DSg8PFJRG21eE2qwv/go-conn-security/insecure"
	ma "gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
	mplex "gx/ipfs/QmZHiqdRuNXujvSPNu1ZWxxzV6a2WhoZpfYkesdgyaKF9f/go-smux-multiplex"
	tptu "gx/ipfs/QmbzWHWFpwCm7GDgrGyKnqUzh48PnRp86VD2Bb8wq6ZfDP/go-libp2p-transport-upgrader"
)

func TestCanDial(t *testing.T) {
	addrWs, err := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/5555/ws")
	if err != nil {
		t.Fatal(err)
	}

	addrTCP, err := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/5555")
	if err != nil {
		t.Fatal(err)
	}

	d := &WebsocketTransport{}
	matchTrue := d.CanDial(addrWs)
	matchFalse := d.CanDial(addrTCP)

	if !matchTrue {
		t.Fatal("expected to match websocket maddr, but did not")
	}

	if matchFalse {
		t.Fatal("expected to not match tcp maddr, but did")
	}
}

func TestWebsocketTransport(t *testing.T) {
	ta := New(&tptu.Upgrader{
		Secure: insecure.New("peerA"),
		Muxer:  new(mplex.Transport),
	})
	tb := New(&tptu.Upgrader{
		Secure: insecure.New("peerB"),
		Muxer:  new(mplex.Transport),
	})

	zero := "/ip4/127.0.0.1/tcp/0/ws"
	utils.SubtestTransport(t, ta, tb, zero, "peerA")
}

func TestWebsocketListen(t *testing.T) {
	zero, err := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/0/ws")
	if err != nil {
		t.Fatal(err)
	}

	tpt := &WebsocketTransport{}
	l, err := tpt.maListen(zero)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	msg := []byte("HELLO WORLD")

	go func() {
		c, err := tpt.maDial(context.Background(), l.Multiaddr())
		if err != nil {
			t.Error(err)
			return
		}

		c.Write(msg)
		c.Close()
	}()

	c, err := l.Accept()
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	obr := iotest.OneByteReader(c)

	out, err := ioutil.ReadAll(obr)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(out, msg) {
		t.Fatal("got wrong message", out, msg)
	}
}

func TestConcurrentClose(t *testing.T) {
	zero, err := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/0/ws")
	if err != nil {
		t.Fatal(err)
	}

	tpt := &WebsocketTransport{}
	l, err := tpt.maListen(zero)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	msg := []byte("HELLO WORLD")

	go func() {
		for i := 0; i < 100; i++ {
			c, err := tpt.maDial(context.Background(), l.Multiaddr())
			if err != nil {
				t.Error(err)
				return
			}

			go c.Write(msg)
			go c.Close()
		}
	}()

	for i := 0; i < 100; i++ {
		c, err := l.Accept()
		if err != nil {
			t.Fatal(err)
		}
		c.Close()
	}
}

func TestWriteZero(t *testing.T) {
	zero, err := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/0/ws")
	if err != nil {
		t.Fatal(err)
	}

	tpt := &WebsocketTransport{}
	l, err := tpt.maListen(zero)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	msg := []byte(nil)

	go func() {
		c, err := tpt.maDial(context.Background(), l.Multiaddr())
		defer c.Close()
		if err != nil {
			t.Error(err)
			return
		}

		for i := 0; i < 100; i++ {
			n, err := c.Write(msg)
			if n != 0 {
				t.Errorf("expected to write 0 bytes, wrote %d", n)
			}
			if err != nil {
				t.Error(err)
				return
			}
		}
	}()

	c, err := l.Accept()
	defer c.Close()
	if err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 100)
	n, err := c.Read(buf)
	if n != 0 {
		t.Errorf("read %d bytes, expected 0", n)
	}
	if err != io.EOF {
		t.Errorf("expected EOF, got err: %s", err)
	}
}
