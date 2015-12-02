// This package implements a general purpose Path-Trie.
package trie

import (
	"os"
	"strings"
)

// Node represents a single node in a Trie, but it can be used as a whole
// (sub-)Trie through the Trie interface. A node value of `nil` is a perfectly
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
}

// Trie represents the required methods for accessing a directory structure.
type Trie interface {
	// Root returns the uppermost node of the trie.
	Root() *Node

	// Insert adds a new node in the trie at string. If the node already exists,
	// nothing changes. This operation costs O(log(n)). The newly created or
	// existant node is returned.
	Insert(path string) *Node

	// Lookup searches for a node references by a path.
	Lookup(path string) *Node

	// Remove removes the node at path and all of it's children.
	// The parent of the removed node is returned, which might be nil.
	Remove() *Node

	// Len returns the current number of elements in the trie.
	// This counts only explicitly inserted Nodes.
	Len() int64
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

// NewTrie returns a trie with the root element pre-inserted.
// Note that `nil` is a perfectly valid, but empty trie.
func NewTrie() *Node {
	return &Node{}
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
	curr := n
	if curr == nil {
		curr = NewTrie()
	}

	wasAdded := false

	for depth, name := range SplitPath(path) {
		if curr.Children == nil {
			curr.Children = make(map[string]*Node)
		}
		child, ok := curr.Children[name]
		if !ok {
			child = &Node{
				Parent: curr,
				Name:   name,
				Depth:  uint16(depth + 1),
			}
			curr.Children[name] = child
			wasAdded = true
		}
		curr = child
	}

	if wasAdded && curr != nil {
		curr.Up(func(parent *Node) {
			parent.Length++
		})
	}
	return curr
}

// Lookup searches a Node by it's absolute path.
func (n *Node) Lookup(path string) *Node {
	curr := n
	if n == nil {
		return nil
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
	n.Up(func(parent *Node) {
		parent.Length -= length
	})

	// Removes link to self:
	if n.Parent != nil {
		delete(n.Parent.Children, n.Name)
	}

	// Make children garbage collectable:
	parent := n.Parent
	n.Walk(true, func(child *Node) {
		child.Children = nil
		child.Parent = nil
	})
	return parent
}

// Walk iterates over all (including intermediate )nodes in the trie.
// Depending on dfs the nodes are visited in depth-first or breadth-first.
// The supplied callback is called for each visited node.
func (n *Node) Walk(dfs bool, visit func(*Node)) {
	if n == nil {
		return
	}

	if !dfs {
		visit(n)
	}

	if n.Children != nil {
		for _, child := range n.Children {
			child.Walk(dfs, visit)
		}
	}

	if dfs {
		visit(n)
	}
}

// Up walks from the receiving node to the root node,
// calling `visit` on each node on it's way.
func (n *Node) Up(visit func(*Node)) {
	if n != nil {
		visit(n)
		n.Parent.Up(visit)
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

	n.Up(func(parent *Node) {
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
