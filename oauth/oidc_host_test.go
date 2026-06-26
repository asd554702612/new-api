package oauth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestOIDCExchangeTokenUsesHostClientAndCarriesUserInfoEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userInfoServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "Bearer host-access-token", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"sub":"new-api-123","email":"user@example.com","preferred_username":"alice","name":"Alice"}`))
	}))
	t.Cleanup(userInfoServer.Close)

	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseForm())
		require.Equal(t, "host-client", r.Form.Get("client_id"))
		require.Equal(t, "host-secret", r.Form.Get("client_secret"))
		require.Equal(t, "authorization-code", r.Form.Get("code"))
		require.Equal(t, "authorization_code", r.Form.Get("grant_type"))
		require.Equal(t, "https://token.gptk.cc.cd/oauth/oidc", r.Form.Get("redirect_uri"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"host-access-token","token_type":"Bearer","expires_in":3600}`))
	}))
	t.Cleanup(tokenServer.Close)

	originalOIDC := *system_setting.GetOIDCSettings()
	originalServerAddress := system_setting.ServerAddress
	t.Cleanup(func() {
		*system_setting.GetOIDCSettings() = originalOIDC
		system_setting.ServerAddress = originalServerAddress
	})
	system_setting.ServerAddress = "https://global.example"
	*system_setting.GetOIDCSettings() = system_setting.OIDCSettings{
		Enabled:               true,
		ClientId:              "global-client",
		ClientSecret:          "global-secret",
		AuthorizationEndpoint: "https://login.example/global/auth",
		TokenEndpoint:         "https://login.example/global/token",
		UserInfoEndpoint:      "https://login.example/global/userinfo",
		HostClients: map[string]system_setting.OIDCClientConfig{
			"token.gptk.cc.cd": {
				ClientId:              "host-client",
				ClientSecret:          "host-secret",
				AuthorizationEndpoint: "https://login.example/host/auth",
				TokenEndpoint:         tokenServer.URL,
				UserInfoEndpoint:      userInfoServer.URL,
			},
		},
	}

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	req := httptest.NewRequest(http.MethodGet, "/api/oauth/oidc?code=authorization-code", nil)
	req.Host = "internal.local"
	req.Header.Set("X-Forwarded-Host", "token.gptk.cc.cd")
	req.Header.Set("X-Forwarded-Proto", "https")
	c.Request = req

	provider := &OIDCProvider{}
	token, err := provider.ExchangeToken(c.Request.Context(), "authorization-code", c)
	require.NoError(t, err)
	require.Equal(t, "host-access-token", token.AccessToken)
	require.Equal(t, userInfoServer.URL, token.UserInfoEndpoint)

	user, err := provider.GetUserInfo(c.Request.Context(), token)
	require.NoError(t, err)
	require.Equal(t, "new-api-123", user.ProviderUserID)
	require.Equal(t, "alice", user.Username)
	require.Equal(t, "Alice", user.DisplayName)
	require.Equal(t, "user@example.com", user.Email)
}
