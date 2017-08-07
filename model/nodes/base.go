package nodes

import (
	"strings"
	"time"

	capnp_model "github.com/disorganizer/brig/model/nodes/capnp"
	h "github.com/disorganizer/brig/util/hashlib"
)

// Base is a place that holds all common attributes of all Nodes.
// It also defines some utility function that will be mixed into real nodes.
type Base struct {
	// Basename of this node
	name string

	// Hash of this node (might be empty)
	hash h.Hash

	// Last modification time of this node.
	modTime time.Time

	// Type of this node
	nodeType NodeType

	// Unique identifier for this node
	uid uint64
}

// Name returns the name of this node (e.g. /a/b/c -> c)
// The root directory will have the name empty string.
func (b *Base) Name() string {
	return b.name
}

// Hash returns the hash of this node.
func (b *Base) Hash() h.Hash {
	return b.hash
}

// Type returns the type of this node.
func (b *Base) Type() NodeType {
	return b.nodeType
}

// ModTime will return the last time this node's content
// was modified. Metadata changes are not recorded.
func (b *Base) ModTime() time.Time {
	return b.modTime
}

// Inode will return a unique ID that is different for each node.
func (b *Base) Inode() uint64 {
	return b.uid
}

/////// UTILS /////////

func (b *Base) setBaseAttrsToNode(capnode capnp_model.Node) error {
	modTimeBin, err := b.modTime.MarshalText()
	if err != nil {
		return err
	}

	if err := capnode.SetModTime(string(modTimeBin)); err != nil {
		return err
	}
	if err := capnode.SetHash(b.hash); err != nil {
		return err
	}
	return capnode.SetName(b.name)
}

func (b *Base) parseBaseAttrsFromNode(capnode capnp_model.Node) error {
	var err error
	b.name, err = capnode.Name()
	if err != nil {
		return err
	}

	b.hash, err = capnode.Hash()
	if err != nil {
		return err
	}

	unparsedModTime, err := capnode.ModTime()
	if err != nil {
		return err
	}

	if err := b.modTime.UnmarshalText([]byte(unparsedModTime)); err != nil {
		return err
	}

	return nil
}

func prefixSlash(s string) string {
	if !strings.HasPrefix(s, "/") {
		return "/" + s
	}

	return s
}
