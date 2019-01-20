package endpoints

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

type loginResponse struct {
	Success bool `json:"success"`
}

func TestLoginEndpointSuccess(t *testing.T) {
	withState(t, func(s *testState) {
		resp := s.mustRun(
			t,
			NewLoginHandler(s.State),
			"POST",
			"http://localhost:5000/api/v0/login",
			&LoginRequest{
				Username: "ali",
				Password: "ila",
			},
		)

		require.Equal(t, http.StatusOK, resp.StatusCode)

		loginResp := &loginResponse{}
		mustDecodeBody(t, resp.Body, &loginResp)

		require.Equal(t, true, loginResp.Success)
		cookies := resp.Cookies()
		require.Equal(t, "sess", cookies[0].Name)
	})
}
