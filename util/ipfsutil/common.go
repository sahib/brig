package ipfsutil

import (
	core "github.com/ipfs/go-ipfs/core"
	"golang.org/x/net/context"
)

// Node remembers the settings needed for accessing the ipfs daemon.
type Node struct {
	ipfsNode *core.IpfsNode
	Path     string
	// TODO: Not needed.
	APIPort   int
	SwarmPort int

	Context context.Context
	Cancel  context.CancelFunc
}

// TODO
// const MQTTProtocolID = "/brig/mqtt"
//
// type Conn struct {
// }
//
// type Listener struct {
// 	nd *Node
// }
//
// // Accept waits for and returns the next connection to the listener.
// func (ls *Listener) Accept() (Conn, error) {
// 	if !ls.nd.IsOnline() {
// 		return ErrOffline
// 	}
//
// 	ipnd := ls.nd.ipfsNode
// 	ipnd.PeerHost.SetStreamHandler(MQTTProtocolID, func(st inet.Stream) {
//
// 	})
// }
//
// // Close closes the listener.
// // Any blocked Accept operations will be unblocked and return errors.
// func (ls *Listener) Close() error {
//
// }
//
// // Addr returns the listener's network address.
// func (ls *Listener) Addr() Addr {
// 	return &net.TCPAddr{
// 		IP:   net.IPv4(127, 0, 0, 1),
// 		Port: ls.nd.SwarmPort,
// 	}
// }
//
// func (nd *Node) RegMQTT() {
// 	fmt.Println("Im", nd.ipfsNode.Identity.Pretty())
//
// 	nd.ipfsNode.PeerHost.SetStreamHandler(MQTTProtocolID, func(st inet.Stream) {
// 		defer st.Close()
//
// 		data, err := ioutil.ReadAll(st)
// 		if err != nil {
// 			fmt.Println("Failed to read stream data", err)
// 			return
// 		}
//
// 		fmt.Println("Stream data was", string(data))
//
// 		if _, err := st.Write(data); err != nil {
// 			fmt.Println("Write back failed", err)
// 			return
// 		}
// 	})
// }
//
// func (nd *Node) SendMQTT(id string) {
// 	pid, err := peer.IDB58Decode(id)
// 	if err != nil {
// 		fmt.Println("Bad ID", err)
// 		return
// 	}
//
// 	st, err := nd.ipfsNode.PeerHost.NewStream(nd.Context, MQTTProtocolID, pid)
// 	if err != nil {
// 		fmt.Println("New stream failed:", err)
// 		return
// 	}
//
// 	defer st.Close()
//
// 	data := testutil.CreateDummyBuf(64*1024 + 1)
// 	if _, err := st.Write(data); err != nil {
// 		fmt.Println("Send write failed", err)
// 		return
// 	}
// }
