package util

import (
	"fmt"
	"os"
	"strings"
)

type Node struct {
	Parent   *Node
	Children map[string]*Node
	Name     string
	NodeSize int64
}

type Trie interface {
	Root() *Node

	Add(string) *Node

	Lookup(string) *Node

	Remove(string) *Node

	Size() int64
}

func NewTrie() *Node {
	return &Node{NodeSize: 0}
}

func (n *Node) Root() *Node {
	if n != nil && n.Parent != nil {
		return n.Parent.Root()
	}
	return n
}

func (n *Node) Add(path string) *Node {
	names := strings.Split(path, string(os.PathSeparator))
	curr := n
	if curr == nil {
		curr = &Node{NodeSize: 0}
	}

	for _, name := range names {
		if curr.Children == nil {
			curr.Children = make(map[string]*Node)
		}
		child, ok := curr.Children[name]
		if !ok {
			child = &Node{
				Parent:   curr,
				Children: nil,
				Name:     name,
				NodeSize: 1,
			}
			curr.Children[name] = child
			curr.Up(func(parent *Node) {
				parent.NodeSize++
			})
		}
		curr = child
	}

	return curr
}

func (n *Node) Lookup(path string) *Node {
	if n == nil {
		return nil
	}
	names := strings.Split(path, string(os.PathSeparator))
	curr := n
	for _, name := range names {
		if child, ok := curr.Children[name]; !ok {
			return nil
		} else {
			curr = child
		}
	}
	return curr
}

func (n *Node) Remove(path string) *Node {
	node := n.Lookup(path)
	if node == nil {
		return nil
	}

	// Adjusts the parent's size.
	size := node.NodeSize
	node.Up(func(parent *Node) {
		parent.NodeSize -= size
	})

	// Removes link to self.
	if node.Parent != nil {
		delete(node.Parent.Children, node.Name)
	}

	// Make children garbage collectable.
	parent := node.Parent
	node.Iterate(func(child *Node) {
		child.Children = nil
		child.Parent = nil
	})
	return parent
}

func (n *Node) Iterate(visit func(*Node)) {
	if n == nil {
		return
	}

	if n.Children != nil {
		for _, child := range n.Children {
			child.Iterate(visit)
		}
	}
	visit(n)
}

func (n *Node) Up(visit func(*Node)) {
	if n != nil {
		visit(n)
		n.Parent.Up(visit)
	}
}

func (n *Node) Size() int64 {
	if n == nil {
		return 0
	}
	return n.NodeSize
}

func (n *Node) String() string {
	if n == nil {
		return "<nil>"
	}
	return fmt.Sprintf("[%p:%s](N: %d)", n, n.Name, n.NodeSize)
}
