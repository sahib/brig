package server

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParsePath(t *testing.T) {
	tcs := []struct {
		Path   string
		Expect *URL
	}{
		{Path: "/", Expect: &URL{User: "", Path: "/"}},
		{Path: "/a/b/c", Expect: &URL{User: "", Path: "/a/b/c"}},
		{Path: "a/b/c", Expect: &URL{User: "", Path: "/a/b/c"}},
		{Path: "a:/b/c", Expect: &URL{User: "a", Path: "/b/c"}},
		{Path: "a:b/c", Expect: &URL{User: "a", Path: "/b/c"}},
		{Path: "a:b/c/..", Expect: &URL{User: "a", Path: "/b"}},
		{Path: "a::b", Expect: &URL{User: "a", Path: "/:b"}},
		{Path: "a::", Expect: &URL{User: "a", Path: "/:"}},
		{Path: "a:", Expect: &URL{User: "", Path: "/a:"}},
		{Path: ":a", Expect: &URL{User: "", Path: "/:a"}},
	}

	for _, tc := range tcs {
		t.Run(tc.Path, func(t *testing.T) {
			got, err := parsePath(tc.Path)
			require.Nil(t, err)
			require.Equal(t, tc.Expect, got)
		})
	}
}
