package endpoints

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLogEndpointSuccess(t *testing.T) {
	withState(t, func(s *testState) {
		require.Nil(t, s.fs.Stage("/x", bytes.NewReader([]byte("hello"))))
		require.Nil(t, s.fs.MakeCommit("hello"))
		require.Nil(t, s.fs.Stage("/x", bytes.NewReader([]byte("world"))))
		require.Nil(t, s.fs.MakeCommit("world"))
		require.Nil(t, s.fs.Remove("/x"))
		require.Nil(t, s.fs.MakeCommit("remove"))

		resp := s.mustRun(
			t,
			NewLogHandler(s.State),
			"POST",
			"http://localhost:5000/api/v0/log",
			nil,
		)

		require.Equal(t, http.StatusOK, resp.StatusCode)

		data := &LogResponse{}
		mustDecodeBody(t, resp.Body, &data)
		require.Equal(t, true, data.Success)
		require.Equal(t, 4, len(data.Commits))

		require.Equal(t, "", data.Commits[0].Msg)
		require.Equal(t, []string{"curr"}, data.Commits[0].Tags)

		require.Equal(t, "remove", data.Commits[1].Msg)
		require.Equal(t, []string{"head"}, data.Commits[1].Tags)

		require.Equal(t, "world", data.Commits[2].Msg)
		require.Equal(t, []string{}, data.Commits[2].Tags)

		require.Equal(t, "hello", data.Commits[3].Msg)
		require.Equal(t, []string{"init"}, data.Commits[3].Tags)
	})
}
