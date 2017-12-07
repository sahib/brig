package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/sahib/brig/client"
	"github.com/sahib/brig/util/colors"
)

var (
	treeRunePipe   = "│"
	treeRuneTri    = "├"
	treeRuneBar    = "──"
	treeRuneCorner = "└"
)

type treeNode struct {
	name     string
	order    []*treeNode
	children map[string]*treeNode
	isLast   bool
	parent   *treeNode
	depth    int
	entry    *client.StatInfo
}

// This is a very stripped down version util.Trie.Insert()
// but with support for ordering the elements.
func (n *treeNode) Insert(entry *client.StatInfo) {
	parts := strings.Split(entry.Path, "/")
	if len(parts) > 0 && parts[0] == "" {
		parts = parts[1:]
	}

	curr := n

	for depth, name := range parts {
		if curr.children == nil {
			curr.children = make(map[string]*treeNode)
		}

		child, ok := curr.children[name]
		if !ok {
			child = &treeNode{
				name:  name,
				depth: depth + 1,
				entry: entry,
			}

			child.isLast = true
			if len(curr.order) > 0 {
				curr.order[len(curr.order)-1].isLast = false
			}

			child.parent = curr
			curr.children[name] = child
			curr.order = append(curr.order, child)
		}

		curr = child
	}
}

func (n *treeNode) Len() int {
	return len(n.order)
}

func (n *treeNode) Swap(i, j int) {
	n.order[i], n.order[j] = n.order[j], n.order[i]

	// This is not very clever, but works and is obvious:
	n.order[i].isLast = i == len(n.order)-1
	n.order[j].isLast = j == len(n.order)-1
}

func (n *treeNode) Less(i, j int) bool {
	return n.order[i].name < n.order[j].name
}

func (n *treeNode) Print() {
	parents := make([]*treeNode, n.depth)
	curr := n

	sort.Sort(n)

	// You could do probably go upwards and print to
	// a string buffer for performance, but this is "fun" code...
	for i := 0; i < n.depth; i++ {
		parents[n.depth-i-1] = curr
		curr = curr.parent
	}

	for i := 0; i < n.depth; i++ {
		if i == n.depth-1 {
			if n.isLast {
				fmt.Printf("%s", treeRuneCorner)
			} else {
				fmt.Printf("%s", treeRuneTri)
			}
		} else {
			if parents[i].isLast {
				fmt.Printf("%s  ", " ")
			} else {
				fmt.Printf("%s  ", treeRunePipe)
			}
		}
	}

	name, prefix := n.name, treeRuneBar

	switch {
	case n.name == "/":
		name, prefix = colors.Colorize(n.name, colors.Magenta), ""
	case n.entry.IsDir:
		name = colors.Colorize(name, colors.Green)
	}

	fmt.Printf("%s%s\n", prefix, name)

	for _, child := range n.order {
		child.Print()
	}
}

func showTree(entries []*client.StatInfo, maxDepth int) error {
	root := &treeNode{name: "/"}
	nfiles, ndirs := 0, 0

	for _, entry := range entries {
		if entry.Path == "/" {
			root.entry = entry
		} else {
			root.Insert(entry)
		}

		if entry.IsDir {
			ndirs++
		} else {
			nfiles++
		}
	}

	root.Print()

	fmt.Printf("\n%d directories, %d files\n", ndirs, nfiles)
	return nil
}
