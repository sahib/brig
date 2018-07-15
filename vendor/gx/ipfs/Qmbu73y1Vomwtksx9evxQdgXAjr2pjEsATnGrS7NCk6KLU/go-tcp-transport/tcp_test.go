package tcp

import (
	"testing"

	utils "gx/ipfs/QmW5LxJm2Yo3S7uVsfLM7NsJn2QnKbvZD7uYsZVYR7YViE/go-libp2p-transport/test"
	insecure "gx/ipfs/QmYMGCwtKbLXfh2MUTbLoP66N9Y24DSg8PFJRG21eE2qwv/go-conn-security/insecure"
	ma "gx/ipfs/QmYmsdtJ3HsodkePE3eU3TsCaP2YvPZJ4LoXnNkDE5Tpt7/go-multiaddr"
	mplex "gx/ipfs/QmZHiqdRuNXujvSPNu1ZWxxzV6a2WhoZpfYkesdgyaKF9f/go-smux-multiplex"
	tptu "gx/ipfs/QmbzWHWFpwCm7GDgrGyKnqUzh48PnRp86VD2Bb8wq6ZfDP/go-libp2p-transport-upgrader"
)

func TestTcpTransport(t *testing.T) {
	for i := 0; i < 2; i++ {
		ta := NewTCPTransport(&tptu.Upgrader{
			Secure: insecure.New("peerA"),
			Muxer:  new(mplex.Transport),
		})
		tb := NewTCPTransport(&tptu.Upgrader{
			Secure: insecure.New("peerB"),
			Muxer:  new(mplex.Transport),
		})

		zero := "/ip4/127.0.0.1/tcp/0"
		utils.SubtestTransport(t, ta, tb, zero, "peerA")

		envReuseportVal = false
	}
	envReuseportVal = true
}

func TestTcpTransportCantListenUtp(t *testing.T) {
	for i := 0; i < 2; i++ {
		utpa, err := ma.NewMultiaddr("/ip4/127.0.0.1/udp/0/utp")
		if err != nil {
			t.Fatal(err)
		}

		tpt := NewTCPTransport(&tptu.Upgrader{
			Secure: insecure.New("peerB"),
			Muxer:  new(mplex.Transport),
		})

		_, err = tpt.Listen(utpa)
		if err == nil {
			t.Fatal("shouldnt be able to listen on utp addr with tcp transport")
		}

		envReuseportVal = false
	}
	envReuseportVal = true
}
