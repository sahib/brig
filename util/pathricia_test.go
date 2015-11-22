package util

import (
	"fmt"
	"testing"
)

func TestPathricia(t *testing.T) {
	trie := NewTrie()
	fmt.Println(trie.Size())

	n := trie.Add("/home/qitta")
	fmt.Println(n, n.Root(), n.Parent)
	n = trie.Add("/home/sahib/gat/georg")
	fmt.Println(n, n.Root(), n.Parent)
	n = trie.Remove("/home/sahib/gat/georg")
	fmt.Println(n, n.Root(), n.Parent)
	n = trie.Remove("/home")
	fmt.Println(n, n.Root(), n.Parent)
	n = trie.Remove("/")
	fmt.Println(n, n.Root(), n)
}
