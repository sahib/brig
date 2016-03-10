// Package trie implements a general purpose Path-*Node.
package trie

import (
	"fmt"
	"os"
	"strings"
)

// Node represents a single node in a *Node, but it can be used as a whole
// (sub-)*Node through the *Node interface. A node value of `nil` is a perfectly
// valid trie. Node is suitable for embedding it into other structs.
type Node struct {
	// Pointer to parent node or nil
	Parent *Node

	// Basename to child-nodes
	Children map[string]*Node

	// Basename of the node's Path
	Name string

	// Number of explicitly added children of this node.
	// (1 for leaf nodes)
	Length int64

	// Depth of the node. The root is at depth 0.
	Depth uint16

	// Arbitrary data pointer
	Data interface{}
}

// SplitPath splits the path according to os.PathSeparator,
// but omits a leading empty name on /unix/paths
func SplitPath(path string) []string {
	names := strings.Split(path, string(os.PathSeparator))
	if len(names) > 0 && names[0] == "" {
		return names[1:]
	}

	return names
}

// NewNode returns a trie with the root element pre-inserted.
// Note that `nil` is a perfectly valid, but empty trie.
func NewNode() *Node {
	return &Node{}
}

// NewNodeWithData works like NewNode but populates the .Data field.
func NewNodeWithData(data interface{}) *Node {
	return &Node{Data: data}
}

// Root returns the root node of the trie.
func (n *Node) Root() *Node {
	if n != nil && n.Parent != nil {
		return n.Parent.Root()
	}
	return n
}

// Insert adds a node into the trie at `path`
func (n *Node) Insert(path string) *Node {
	return n.InsertWithData(path, nil)
}

// InsertWithData adds a node into the trie at `path`, storing `data`
// in the Node.Data field.
func (n *Node) InsertWithData(path string, data interface{}) *Node {
	curr := n

	// Empty node, create new one implicitly:
	if curr == nil {
		curr = NewNode()
	}

	wasAdded := false

	for _, name := range SplitPath(path) {
		if curr.Children == nil {
			curr.Children = make(map[string]*Node)
		}
		child, ok := curr.Children[name]
		if !ok {
			child = &Node{
				Parent: curr,
				Name:   name,
				Depth:  uint16(curr.Depth + 1),
				Data:   data,
			}

			curr.Children[name] = child
			wasAdded = true
		}

		curr = child
	}

	if wasAdded && curr != nil {
		curr.up(func(parent *Node) {
			parent.Length++
		})
	}

	if curr != nil {
		curr.Data = data
	}

	return curr
}

// Lookup searches a Node by it's absolute path.
// Returns nil if Node does not exist.
func (n *Node) Lookup(path string) *Node {
	curr := n
	if n == nil {
		return nil
	}

	if path == "/" {
		return n.Root()
	}

	for _, name := range SplitPath(path) {
		child, ok := curr.Children[name]
		if !ok {
			return nil
		}
		curr = child
	}
	return curr
}

// Remove removes the receiver and all of it's children.
// The removed node's parent is returned.
func (n *Node) Remove() *Node {
	if n == nil {
		return nil
	}

	// Adjusts the parent's length:
	length := n.Length
	n.up(func(parent *Node) {
		parent.Length -= length
	})

	// Removes link to self:
	if n.Parent != nil {
		delete(n.Parent.Children, n.Name)
	}

	// Make children garbage collectable:
	parent := n.Parent
	n.Walk(true, func(child *Node) bool {
		child.Children = nil
		child.Parent = nil
		return true
	})
	return parent
}

// Walk iterates over all (including intermediate )nodes in the trie.
// Depending on dfs the nodes are visited in depth-first or breadth-first.
// The supplied callback is called once for each visited node.
func (n *Node) Walk(dfs bool, visit func(*Node) bool) {
	if n == nil {
		return
	}

	if !dfs {
		if !visit(n) {
			return
		}
	}

	if n.Children != nil {
		for _, child := range n.Children {
			child.Walk(dfs, visit)
		}
	}

	if dfs {
		if !visit(n) {
			return
		}
	}
}

// Print dumps a debugging representation of the trie on stdout.
func (n *Node) Print() {
	n.Walk(false, func(child *Node) bool {
		fmt.Println(strings.Repeat(" ", int(child.Depth)*4), child.Name)
		return true
	})
}

// Up walks from the receiving node to the root node,
// calling `visit` on each node on it's way.
func (n *Node) Up(visit func(*Node)) {
	n.up(func(n *Node) {
		visit(n)
	})
}

// up is the same as Up, but works on the native *Node.
func (n *Node) up(visit func(*Node)) {
	if n != nil {
		visit(n)
		n.Parent.up(visit)
	}
}

// Len returns the number of explicitly inserted elements in the trie.
func (n *Node) Len() int64 {
	if n == nil {
		return 0
	}
	return n.Length
}

// Path returns a full absolute path from the receiver
// to the root node of the trie.
func (n *Node) Path() string {
	if n == nil {
		return ""
	}

	s := make([]string, n.Depth+2)
	i := len(s) - 1

	n.up(func(parent *Node) {
		s[i] = parent.Name
		i--
	})

	return buildPath(s)
}

// String returns the absolute path of the node.
func (n *Node) String() string {
	if n == nil {
		return "<nil>"
	}
	return n.Path()
}
