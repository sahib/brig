package cmd

import (
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/fatih/color"
	"github.com/sahib/brig/client"
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
	entry    client.StatInfo
}

type treeCfg struct {
	showPin bool
	format  func(n *treeNode) string
}

// This is a very stripped down version of util.Trie.Insert()
// but with support for ordering the elements.
func (n *treeNode) Insert(entry client.StatInfo) {
	parts := strings.Split(entry.Path, "/")
	if len(parts) > 0 && parts[0] == "" {
		parts = parts[1:]
	}

	curr := n
	currParts := []string{""}

	for depth, name := range parts {
		if curr.children == nil {
			curr.children = make(map[string]*treeNode)
		}

		currParts = append(currParts, name)
		currPath := strings.Join(currParts, "/")

		child, ok := curr.children[name]
		if !ok {
			childEntry := entry
			if currPath != entry.Path {
				childEntry = client.StatInfo{
					IsPinned: false,
					Path:     currPath,
				}
			}

			child = &treeNode{
				name:  name,
				depth: depth + 1,
				entry: childEntry,
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
	// Sort case insensitive.
	iRunes := []rune(n.order[i].name)
	jRunes := []rune(n.order[j].name)

	for idx := 0; idx < len(iRunes) && idx < len(jRunes); idx++ {
		ir := iRunes[idx]
		jr := jRunes[idx]

		lir := unicode.ToLower(ir)
		ljr := unicode.ToLower(jr)

		if lir != ljr {
			return lir < ljr
		}

		if ir != jr {
			return ir < jr
		}
	}

	return false
}

func (n *treeNode) Print(cfg *treeCfg) {
	parents := make([]*treeNode, n.depth)
	curr := n

	sort.Sort(n)

	// You could do probably go upwards and print to
	// a string buffer for performance, but this is probably
	// not necessary/critical here.
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

	// Default to an auto-formatter:
	format := cfg.format
	if format == nil {
		format = func(n *treeNode) string {
			switch {
			case n.name == "/":
				return color.MagentaString("•")
			case n.entry.IsDir:
				return color.GreenString(n.name)
			}

			return " " + n.name
		}
	}

	prefix := treeRuneBar
	if n.name == "/" {
		prefix = ""
	}

	formatted := format(n)
	pinState := ""
	if cfg.showPin && n.entry.IsPinned {
		pinState += " " + color.CyanString("🖈")
	}

	fmt.Printf("%s%s%s\n", prefix, formatted, pinState)
	for _, child := range n.order {
		child.Print(cfg)
	}
}

func showTree(entries []client.StatInfo, cfg *treeCfg) {
	root := &treeNode{name: "/"}
	nfiles, ndirs := 0, 0

	hasRoot := false
	for _, entry := range entries {
		if entry.Path == "/" {
			root.entry = entry
			hasRoot = true
		} else {
			root.Insert(entry)
		}

		if entry.IsDir {
			ndirs++
		} else {
			nfiles++
		}
	}

	if !hasRoot {
		root.entry = client.StatInfo{
			Path: "/",
		}
	}

	root.Print(cfg)

	// Speak understandable english:
	dirLabel := "directories"
	if ndirs == 1 {
		dirLabel = "directory"
	}

	fileLabel := "files"
	if nfiles == 1 {
		fileLabel = "file"
	}

	fmt.Printf("\n%d %s, %d %s\n", ndirs, dirLabel, nfiles, fileLabel)
}
