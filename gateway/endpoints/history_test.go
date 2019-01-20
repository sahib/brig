package endpoints

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHistoryEndpointSuccess(t *testing.T) {
	withState(t, func(s *testState) {
		require.Nil(t, s.fs.Stage("/x", bytes.NewReader([]byte("hello"))))
		require.Nil(t, s.fs.MakeCommit("hello"))
		require.Nil(t, s.fs.Stage("/x", bytes.NewReader([]byte("world"))))
		require.Nil(t, s.fs.MakeCommit("world"))
		require.Nil(t, s.fs.Remove("/x"))
		require.Nil(t, s.fs.MakeCommit("remove"))

		resp := s.mustRun(
			t,
			NewHistoryHandler(s.State),
			"POST",
			"http://localhost:5000/api/v0/history",
			&HistoryRequest{
				Path: "/x",
			},
		)

		require.Equal(t, http.StatusOK, resp.StatusCode)

		data := &HistoryResponse{}
		mustDecodeBody(t, resp.Body, &data)
		require.Equal(t, true, data.Success)

		ents := data.Entries
		require.Len(t, ents, 3)

		require.Equal(t, "removed", ents[0].Change)
		require.Equal(t, "/x", ents[0].Path)
		require.Equal(t, "remove", ents[0].Head.Msg)
		require.Equal(t, []string{"head"}, ents[0].Head.Tags)

		require.Equal(t, "modified", ents[1].Change)
		require.Equal(t, "/x", ents[1].Path)
		require.Equal(t, "world", ents[1].Head.Msg)
		require.Equal(t, []string{}, ents[1].Head.Tags)

		require.Equal(t, "added", ents[2].Change)
		require.Equal(t, "/x", ents[2].Path)
		require.Equal(t, "hello", ents[2].Head.Msg)
		require.Equal(t, []string{"init"}, ents[2].Head.Tags)
	})
}

func TestHistoryEndpointForbidden(t *testing.T) {
	withState(t, func(s *testState) {
		s.mustChangeFolders(t, "/public")

		resp := s.mustRun(
			t,
			NewHistoryHandler(s.State),
			"POST",
			"http://localhost:5000/api/v0/history",
			&HistoryRequest{
				Path: "/x",
			},
		)

		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}
