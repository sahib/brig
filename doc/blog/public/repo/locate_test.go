package repo

import (
	"os"
	"path/filepath"
	"testing"
)

var (
	TestPath      = filepath.Join(os.TempDir(), "brig-test")
	TestPathEmpty = filepath.Join(TestPath, "a", "b", "c", "d")
	TestPathRepo  = filepath.Join(TestPath, "a", ".brig")
)

func createTestDir() {
	for _, dir := range []string{TestPathEmpty, TestPathRepo} {
		if err := os.MkdirAll(dir, 0777); err != nil {
			panic(err)
		}
	}
}

func purgeTestDir() {
	err := os.RemoveAll(TestPath)
	if err != nil {
		panic(err)
	}
}

func TestFindRepo(t *testing.T) {
	createTestDir()
	defer purgeTestDir()

	tests := []struct {
		input string
		want  string
	}{
		{TestPath, ""},
		{TestPathEmpty, filepath.Dir(TestPathRepo)},
		{TestPathRepo, filepath.Dir(TestPathRepo)},
		{filepath.Dir(TestPathRepo), filepath.Dir(TestPathRepo)},
	}

	for _, test := range tests {
		got := FindRepo(test.input)
		if got != test.want {
			t.Errorf("\nFindRepo(%q) == %q\nexpected: %q",
				test.input, got, test.want)
		}
	}
}
