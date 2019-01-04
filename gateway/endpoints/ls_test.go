package endpoints

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sahib/brig/catfs"
	"github.com/sahib/config"
	"github.com/stretchr/testify/require"
)

func TestLsEndpoint(t *testing.T) {
	withFsAndCfg(t, func(cfg *config.Config, fs *catfs.FS) {
		req := mustCreateRequest(
			t,
			"POST",
			"http://localhost:5000/api/v0/ls",
			&LsRequest{
				Root:     "/",
				MaxDepth: -1,
			},
		)

		rsw := httptest.NewRecorder()
		hdl := NewLsHandler(cfg, fs)
		hdl.ServeHTTP(rsw, req)

		resp := rsw.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		lsResp := &LsResponse{}
		mustDecodeBody(t, resp.Body, &lsResp)

		require.Len(t, lsResp.Files, 3)
		require.Equal(t, lsResp.Files[0].Path, "/")
		require.Equal(t, lsResp.Files[1].Path, "/hello")
		require.Equal(t, lsResp.Files[2].Path, "/hello/world.png")
	})
}
