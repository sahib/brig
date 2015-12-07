package trie

import "testing"

func TestPathriciaInsertTrieLinux(t *testing.T) {
	tests := []struct {
		input  string
		name   string
		path   string
		length int64
	}{
		//Insert path | expected node name | expected path | trie.Len()
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
			t.Errorf("Node is nil: %v", test)
			continue
		}

		nodeLen := node.Root().Len()
		if nodeLen != test.length {
			t.Errorf("Length differs, got: %d != expected: %d", nodeLen, test.length)
		}

		if node.Name != test.name {
			t.Errorf("Name differs, got: %s != expected: %s", node.Name, test.name)
		}

		if node.Path() != test.path {
			t.Errorf("Path differs, got: %s != expected: %s", node.Path(), test.path)
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
		//Insert path | expected node name | expected path | trie.Len()
		{"", "", "/", 0},
		{"/", "", "/", 1},
		{"a", "a", "/a", 2},
		{"b", "b", "/a/b", 3},
		{"c", "c", "/a/b/c", 4},
		{"c/de/fe", "fe", "/a/b/c/c/de/fe", 5},
		{"c/de/fe/333", "333", "/a/b/c/c/de/fe/c/de/fe/333", 6},
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

func TestPathriciaRemoveLinux(t *testing.T) {
	paths := []string{
		"home/qitta",
		"/sahib",
		"/eule",
		"home/eule",
		"katze/eule",
		"elch/eule",
		"elch/eule/meow",
	}

	trie := NewTrie()
	for _, path := range paths {
		trie.Insert(path)
	}

	tests := []struct {
		path   string
		length int64
		name   string
	}{
		{"/home", 5, ""},
		{"/katze/Eule", 5, ""},
		{"/katze/eule", 4, "katze"},
		{"/elch/eule/meow", 3, "eule"},
		{"/", 0, ""},
		{"/", 0, ""},
	}

	for _, test := range tests {
		node := trie.Lookup(test.path).Remove()
		if node == nil {
			continue
		}

		if node.Name != test.name {
			t.Errorf("\nRemoving: [%s]\nName differs, got: %s != expected: %s\n", test.path, node.Name, test.name)
		}

		if trie.Length != test.length {
			t.Errorf("Length differs, got: %d != expected: %d\n", trie.Length, test.length)
		}
	}
}
