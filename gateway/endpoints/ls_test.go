package endpoints

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLsEndpoint(t *testing.T) {
	withState(t, func(s *testState) {
		exampleData := bytes.NewReader([]byte("Hello world"))
		require.Nil(t, s.fs.Stage("/hello/world.png", exampleData))

		resp := s.mustRun(
			t,
			NewLsHandler(s.State),
			"POST",
			"http://localhost:5000/api/v0/ls",
			&LsRequest{
				Root: "/",
			},
		)

		require.Equal(t, http.StatusOK, resp.StatusCode)

		lsResp := &LsResponse{}
		mustDecodeBody(t, resp.Body, &lsResp)

		require.Len(t, lsResp.Files, 1)
		require.Equal(t, lsResp.Files[0].Path, "/hello")
	})
}
