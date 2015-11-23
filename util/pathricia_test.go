package util

import (
	"fmt"
	"testing"
)

func TestPathriciaLinux(t *testing.T) {
	trie := NewTrie()
	fmt.Println(trie.Len())

	n := trie.Insert("/home/qitta")
	fmt.Println(n, n.Root(), n.Parent)
	n = trie.Insert("/home/sahib/gat/georg")
	fmt.Println(n, n.Root(), n.Parent)
	fmt.Println("Whole trie:")

	trie.Walk(false, func(child *Node) {
		fmt.Println("  :", child)
	})

	fmt.Println("----", SplitPath("/home/qitta"))

	n = trie.Lookup("/home/sahib/gat/georg").Remove()
	fmt.Println(n, n.Root(), n.Parent)
	n = trie.Lookup("/home").Remove()
	fmt.Println(n, n.Root(), n.Parent)
	n = trie.Lookup("/").Remove()
	fmt.Println(n, n.Root(), n)

	var x *Node = nil
	fmt.Println(x.Path())
}

func TestPathriciaWindows(t *testing.T) {
	// TODO: This fails:
	trie := NewTrie()
	n := trie.Insert("C:/Albert")
	fmt.Println(n)
	trie.Walk(false, func(child *Node) {
		fmt.Println("  :", child)
	})
}
