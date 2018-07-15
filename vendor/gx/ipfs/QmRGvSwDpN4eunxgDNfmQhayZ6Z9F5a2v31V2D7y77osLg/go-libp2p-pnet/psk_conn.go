package pnet

import (
	"crypto/cipher"
	"crypto/rand"
	"io"
	"net"

	salsa20 "gx/ipfs/QmNi5J1mEQKAKWbPRBEMKKYVNok9EN4MsGM4YUqPvraPEX/go-crypto-dav/salsa20"
	ipnet "gx/ipfs/QmW7Ump7YyBMr712Ta3iEVh3ZYcfVvJaPryfbCnyE826b4/go-libp2p-interface-pnet"
	mpool "gx/ipfs/QmWBug6eBS7AxRdCDVuSY5CnSit7cS2XnPFYJWqWDumhCG/go-msgio/mpool"
)

// we are using buffer pool as user needs their slice back
// so we can't do XOR cripter in place
var (
	errShortNonce  = ipnet.NewError("could not read full nonce")
	errInsecureNil = ipnet.NewError("insecure is nil")
	errPSKNil      = ipnet.NewError("pre-shread key is nil")
)

type pskConn struct {
	net.Conn
	psk *[32]byte

	writeS20 cipher.Stream
	readS20  cipher.Stream
}

func (c *pskConn) Read(out []byte) (int, error) {
	if c.readS20 == nil {
		nonce := make([]byte, 24)
		_, err := io.ReadFull(c.Conn, nonce)
		if err != nil {
			return 0, errShortNonce
		}
		c.readS20 = salsa20.New(c.psk, nonce)
	}

	maxn := uint32(len(out))
	in := mpool.ByteSlicePool.Get(maxn).([]byte) // get buffer
	defer mpool.ByteSlicePool.Put(maxn, in)      // put the buffer back

	in = in[:maxn]            // truncate to required length
	n, err := c.Conn.Read(in) // read to in
	if n > 0 {
		c.readS20.XORKeyStream(out[:n], in[:n]) // decrypt to out buffer
	}
	return n, err
}

func (c *pskConn) Write(in []byte) (int, error) {
	if c.writeS20 == nil {
		nonce := make([]byte, 24)
		_, err := rand.Read(nonce)
		if err != nil {
			return 0, err
		}
		_, err = c.Conn.Write(nonce)
		if err != nil {
			return 0, err
		}

		c.writeS20 = salsa20.New(c.psk, nonce)
	}
	n := uint32(len(in))
	out := mpool.ByteSlicePool.Get(n).([]byte) // get buffer
	defer mpool.ByteSlicePool.Put(n, out)      // put the buffer back

	out = out[:n]                    // truncate to required length
	c.writeS20.XORKeyStream(out, in) // encrypt

	return c.Conn.Write(out) // send
}

var _ net.Conn = (*pskConn)(nil)

func newPSKConn(psk *[32]byte, insecure net.Conn) (net.Conn, error) {
	if insecure == nil {
		return nil, errInsecureNil
	}
	if psk == nil {
		return nil, errPSKNil
	}
	return &pskConn{
		Conn: insecure,
		psk:  psk,
	}, nil
}
