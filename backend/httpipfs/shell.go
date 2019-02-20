package httpipfs

import (
	"fmt"
	"net/http"

	shell "github.com/ipfs/go-ipfs-api"
)

type Node struct {
	sh *shell.Shell
}

func NewNode(port int) (*Node, error) {
	client := &http.Client{}
	addr := fmt.Sprintf("localhost:%d", port)
	sh := shell.NewShellWithClient(addr, client)
	if !sh.IsUp() {
		return nil, fmt.Errorf("could not reach daemon api")
	}

	return &Node{
		sh: sh,
	}, nil
}

func (nd *Node) IsOnline() bool {
	return nd.sh.IsUp()
}

func (nd *Node) Connect() error {
	// TODO
	return nil
}

func (nd *Node) Disconnect() error {
	// TODO
	return nil
}

func (nd *Node) Close() error {
	return nil
}

// Name returns "ipfs" as name of the backend.
func (nd *Node) Name() string {
	return "httpipfs"
}
