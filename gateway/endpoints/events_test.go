package endpoints

import (
	"net/http"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/posener/wstest"
	"github.com/stretchr/testify/require"
)

func TestEvents(t *testing.T) {
	withState(t, func(s *TestState) {
		dialer := wstest.NewDialer(s.evHdl)
		conn, resp, err := dialer.Dial("ws://whatever/ws", nil)
		require.Nil(t, err)

		if got, want := resp.StatusCode, http.StatusSwitchingProtocols; got != want {
			t.Fatalf("resp.StatusCode = %q, want %q", got, want)
		}

		go func() {
			resp := s.mustRun(
				t,
				NewMkdirHandler(s.State),
				"POST",
				"http://localhost:5000/api/v0/mkdir",
				&MkdirRequest{
					Path: "/test",
				},
			)
			require.Equal(t, http.StatusOK, resp.StatusCode)
		}()

		typ, data, err := conn.ReadMessage()
		require.Nil(t, err)
		require.Equal(t, websocket.TextMessage, typ)
		require.Equal(t, []byte("fs"), data)
	})
}
