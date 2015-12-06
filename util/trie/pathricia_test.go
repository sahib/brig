package trie

import "testing"

func TestPathriciaInsertTrieLinux(t *testing.T) {
	tests := []struct {
		input  string
		name   string
		path   string
		length int64
	}{
		//Insert path | expected node name | expected path | path len
		{"", "", "/", 0},
		{"\\", "\\", "/\\", 1},
		{"a", "a", "/a", 2},
		{"a/b", "b", "/a/b", 3},
		{"home", "home", "/home", 4},
		{"sahib", "sahib", "/sahib", 5},
		{"home/qitta", "qitta", "/home/qitta", 6},
		{"   ", "   ", "/   ", 7},
	}

	trie := NewTrie()
	for _, test := range tests {

		// Inserting at the root node.
		node := trie.Insert(test.input)
		if node == nil {
			t.Errorf("Node is nil.", test)
			continue
		}

		nodeLen := node.Root().Len()
		if nodeLen != test.length {
			t.Errorf("Length differs, got: %d != expected: %d\n", nodeLen, test.length)
		}

		if node.Name != test.name {
			t.Errorf("Name differs, got: %s != expected: %s\n", node.Name, test.name)
		}

		if node.Path() != test.path {
			t.Errorf("Path differs, got: %s != expected: %s\n", node.Path(), test.path)
		}

	}
}

func TestPathriciaInsertRelativeLinux(t *testing.T) {
	tests := []struct {
		input  string
		name   string
		path   string
		length int64
	}{
		//Insert path | expected node name | expected path | path len
		{"", "", "/", 0},
		{"/", "", "/", 1},
		{"a", "a", "/a", 2},
		{"b", "b", "/a/b", 3},
		{"c", "c", "/a/b/c", 4},
	}

	trie := NewTrie()
	node := trie.Root()
	for _, test := range tests {

		// Inserting at always at the returned node.
		node = node.Insert(test.input)
		if node == nil {
			t.Errorf("Node is nil.", test)
			continue
		}

		// Check the explicitly added paths.
		nodeLen := trie.Length
		if nodeLen != test.length {
			t.Errorf("Length differs, got: %d != expected: %d\n", nodeLen, test.length)
		}

		if node.Name != test.name {
			t.Errorf("Name differs, got: %s != expected: %s\n", node.Name, test.name)
		}

		if node.Path() != test.path {
			t.Errorf("Path differs, got: %s != expected: %s\n", node.Path(), test.path)
		}
	}

}
