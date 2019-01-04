package endpoints

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sahib/brig/catfs"
	"github.com/sahib/config"
	"github.com/stretchr/testify/require"
)

type loginResponse struct {
	Success bool `json:"success"`
}

func TestLoginEndpointSuccess(t *testing.T) {
	withFsAndCfg(t, func(cfg *config.Config, fs *catfs.FS) {
		cfg.SetString("auth.user", "ali")
		cfg.SetString("auth.pass", "ila")

		req := mustCreateRequest(
			t,
			"POST",
			"http://localhost:5000/api/v0/login",
			&LoginRequest{
				Username: "ali",
				Password: "ila",
			},
		)

		rsw := httptest.NewRecorder()
		hdl := NewLoginHandler(cfg)
		hdl.ServeHTTP(rsw, req)

		resp := rsw.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		loginResp := &loginResponse{}
		mustDecodeBody(t, resp.Body, &loginResp)

		require.Equal(t, true, loginResp.Success)
		cookies := resp.Cookies()
		require.Len(t, cookies, 2)
		require.Equal(t, "session", cookies[0].Name)
		require.Equal(t, "csrf", cookies[1].Name)

	})
}
