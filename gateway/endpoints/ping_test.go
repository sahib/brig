package endpoints

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPingEndpointSuccess(t *testing.T) {
	withState(t, func(s *testState) {
		resp := s.mustRun(
			t,
			NewPingHandler(s.State),
			"POST",
			"http://localhost:5000/api/v0/ping",
			nil,
		)

		require.Equal(t, http.StatusOK, resp.StatusCode)
		pingResp := &PingResponse{}
		mustDecodeBody(t, resp.Body, &pingResp)
		require.Equal(t, true, pingResp.IsOnline)
	})
}
