package endpoints

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sahib/brig/catfs"
	"github.com/sahib/config"
	"github.com/stretchr/testify/require"
)

type mkdirResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func TestMkdirEndpointSuccess(t *testing.T) {
	withFsAndCfg(t, func(cfg *config.Config, fs *catfs.FS) {
		cfg.SetString("auth.user", "ali")
		cfg.SetString("auth.pass", "ila")

		req := mustCreateRequest(
			t,
			"POST",
			"http://localhost:5000/api/v0/mkdir",
			&MkdirRequest{
				Path: "/test",
			},
		)

		rsw := httptest.NewRecorder()
		hdl := NewMkdirHandler(cfg, fs)
		hdl.ServeHTTP(rsw, req)

		resp := rsw.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		mkdirResp := &mkdirResponse{}
		mustDecodeBody(t, resp.Body, &mkdirResp)
		require.Equal(t, true, mkdirResp.Success)

		info, err := fs.Stat("/test")
		require.Nil(t, err)
		require.Equal(t, "/test", info.Path)
	})
}
