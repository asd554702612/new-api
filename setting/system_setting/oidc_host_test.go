package system_setting

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOIDCResolveClientForRequestPrefersForwardedHostAndStripsPort(t *testing.T) {
	originalServerAddress := ServerAddress
	t.Cleanup(func() { ServerAddress = originalServerAddress })
	ServerAddress = "https://global.example"

	settings := &OIDCSettings{
		Enabled:               true,
		ClientId:              "global-client",
		ClientSecret:          "global-secret",
		AuthorizationEndpoint: "https://login.example/global/auth",
		TokenEndpoint:         "https://login.example/global/token",
		UserInfoEndpoint:      "https://login.example/global/userinfo",
		HostClients: map[string]OIDCClientConfig{
			"token.gptk.cc.cd": {
				ClientId:              "host-client",
				ClientSecret:          "host-secret",
				AuthorizationEndpoint: "https://login.example/host/auth",
				TokenEndpoint:         "https://login.example/host/token",
				UserInfoEndpoint:      "https://login.example/host/userinfo",
			},
		},
	}

	req, err := http.NewRequest(http.MethodGet, "http://internal.local/api/status", nil)
	require.NoError(t, err)
	req.Host = "internal.local"
	req.Header.Set("X-Forwarded-Host", "token.gptk.cc.cd:443")
	req.Header.Set("X-Forwarded-Proto", "https")

	resolved, ok := settings.ResolveClientForRequest(req)
	require.True(t, ok)
	require.True(t, resolved.HostMatched)
	require.Equal(t, "token.gptk.cc.cd", resolved.Host)
	require.Equal(t, "https://token.gptk.cc.cd:443", resolved.Origin)
	require.Equal(t, "https://token.gptk.cc.cd:443/oauth/oidc", resolved.RedirectURI)
	require.Equal(t, "host-client", resolved.ClientId)
	require.Equal(t, "host-secret", resolved.ClientSecret)
	require.Equal(t, "https://login.example/host/token", resolved.TokenEndpoint)
}

func TestOIDCResolveClientForRequestFallsBackToGlobalConfig(t *testing.T) {
	originalServerAddress := ServerAddress
	t.Cleanup(func() { ServerAddress = originalServerAddress })
	ServerAddress = "https://global.example"

	settings := &OIDCSettings{
		Enabled:               true,
		ClientId:              "global-client",
		ClientSecret:          "global-secret",
		AuthorizationEndpoint: "https://login.example/global/auth",
		TokenEndpoint:         "https://login.example/global/token",
		UserInfoEndpoint:      "https://login.example/global/userinfo",
		HostClients: map[string]OIDCClientConfig{
			"token.gptk.cc.cd": {
				ClientId:              "host-client",
				ClientSecret:          "host-secret",
				AuthorizationEndpoint: "https://login.example/host/auth",
				TokenEndpoint:         "https://login.example/host/token",
				UserInfoEndpoint:      "https://login.example/host/userinfo",
			},
		},
	}

	req, err := http.NewRequest(http.MethodGet, "https://unknown.example/api/status", nil)
	require.NoError(t, err)
	req.Host = "unknown.example"

	resolved, ok := settings.ResolveClientForRequest(req)
	require.True(t, ok)
	require.False(t, resolved.HostMatched)
	require.Equal(t, "unknown.example", resolved.Host)
	require.Equal(t, "https://unknown.example", resolved.Origin)
	require.Equal(t, "https://global.example/oauth/oidc", resolved.RedirectURI)
	require.Equal(t, "global-client", resolved.ClientId)
	require.Equal(t, "https://login.example/global/auth", resolved.AuthorizationEndpoint)
}

func TestOIDCHasUsableClientAcceptsHostOnlyConfig(t *testing.T) {
	settings := &OIDCSettings{
		Enabled: true,
		HostClients: map[string]OIDCClientConfig{
			"token.gepinkeji.com": {
				ClientId:              "host-client",
				ClientSecret:          "host-secret",
				AuthorizationEndpoint: "https://login.example/host/auth",
				TokenEndpoint:         "https://login.example/host/token",
				UserInfoEndpoint:      "https://login.example/host/userinfo",
			},
		},
	}

	require.True(t, settings.HasUsableClient())
}
