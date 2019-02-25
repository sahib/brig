package httpipfs

import (
	"errors"
	"fmt"

	shell "github.com/ipfs/go-ipfs-api"
	log "github.com/sirupsen/logrus"
)

var (
	ErrOffline = errors.New("backend is in offline mode")
)

type Node struct {
	sh          *shell.Shell
	allowNetOps bool
}

func NewNode(port int) (*Node, error) {
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	log.Infof("Connecting to IPFS HTTP API at %s", addr)
	sh := shell.NewShell(addr)

	return &Node{
		sh:          sh,
		allowNetOps: true,
	}, nil
}

func (nd *Node) IsOnline() bool {
	return nd.sh.IsUp() && nd.allowNetOps
}

func (nd *Node) Connect() error {
	nd.allowNetOps = true
	return nil
}

func (nd *Node) Disconnect() error {
	nd.allowNetOps = false
	return nil
}

func (nd *Node) Close() error {
	return nil
}

// Name returns "ipfs" as name of the backend.
func (nd *Node) Name() string {
	return "httpipfs"
}
