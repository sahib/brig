package endpoints

import (
	"context"
	"net/http"
	"runtime"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/posener/wstest"
	"github.com/stretchr/testify/require"
)

func TestEvents(t *testing.T) {
	withState(t, func(s *testState) {
		// This is stupid. I couldn't get DialContext()
		// to pass the user value to the actual handler.
		// Pretty sure it was a problem on my side though...
		s.evHdl.testing = true

		// This call evHdl.ServeHTTP when sending something on conn.
		dialer := wstest.NewDialer(s.evHdl)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		conn, resp, err := dialer.DialContext(ctx, "ws://whatever/ws", nil)
		require.Nil(t, err)

		if got, want := resp.StatusCode, http.StatusSwitchingProtocols; got != want {
			t.Fatalf("resp.StatusCode = %q, want %q", got, want)
		}

		go func() {
			// give it a little time so ServeHTTP() of the events handler
			// can reach the "please notify me now" stage.
			time.Sleep(100 * time.Millisecond)

			// trigger an event:
			resp := s.mustRun(
				t,
				NewMkdirHandler(s.State),
				"POST",
				"http://localhost:5000/api/v0/events",
				&MkdirRequest{
					Path: "/test",
				},
			)
			require.Equal(t, http.StatusOK, resp.StatusCode)
		}()

		done := make(chan bool)

		go func() {
			typ, data, err := conn.ReadMessage()
			require.Nil(t, err)
			require.Equal(t, websocket.TextMessage, typ)
			require.Equal(t, []byte("fs"), data)
			done <- true
		}()

		select {
		case <-done:
		case <-time.After(10 * time.Second):
			buf := make([]byte, 1<<20)
			stacklen := runtime.Stack(buf, true)
			t.Logf(string(buf[:stacklen]))
			t.Fatalf("test took too long")
		}
	})
}
