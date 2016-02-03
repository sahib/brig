package ipfsutil

import (
	core "github.com/ipfs/go-ipfs/core"
	"golang.org/x/net/context"
)

// Node remembers the settings needed for accessing the ipfs daemon.
type Node struct {
	IpfsNode *core.IpfsNode
	Path     string

	Context context.Context
	Cancel  context.CancelFunc
}
