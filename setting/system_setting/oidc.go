package system_setting

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/QuantumNous/new-api/setting/config"
)

type OIDCSettings struct {
	Enabled               bool                        `json:"enabled"`
	ClientId              string                      `json:"client_id"`
	ClientSecret          string                      `json:"client_secret"`
	WellKnown             string                      `json:"well_known"`
	AuthorizationEndpoint string                      `json:"authorization_endpoint"`
	TokenEndpoint         string                      `json:"token_endpoint"`
	UserInfoEndpoint      string                      `json:"user_info_endpoint"`
	HostClients           map[string]OIDCClientConfig `json:"host_clients"`
}

type OIDCClientConfig struct {
	ClientId              string `json:"client_id"`
	ClientSecret          string `json:"client_secret"`
	WellKnown             string `json:"well_known"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	UserInfoEndpoint      string `json:"user_info_endpoint"`
}

type OIDCResolvedClient struct {
	OIDCClientConfig
	HostMatched bool
	Host        string
	Origin      string
	RedirectURI string
}

// 默认配置
var defaultOIDCSettings = OIDCSettings{}

func init() {
	// 注册到全局配置管理器
	config.GlobalConfig.Register("oidc", &defaultOIDCSettings)
}

func GetOIDCSettings() *OIDCSettings {
	return &defaultOIDCSettings
}

func (settings *OIDCSettings) ResolveClientForRequest(r *http.Request) (OIDCResolvedClient, bool) {
	if settings == nil || !settings.Enabled {
		return OIDCResolvedClient{}, false
	}

	origin := OIDCRequestOrigin(r)
	host := NormalizeOIDCHost(OIDCRequestHost(r))
	if host != "" {
		for configuredHost, client := range settings.HostClients {
			if NormalizeOIDCHost(configuredHost) != host {
				continue
			}
			if !client.IsComplete() {
				return OIDCResolvedClient{}, false
			}
			return OIDCResolvedClient{
				OIDCClientConfig: client,
				HostMatched:      true,
				Host:             host,
				Origin:           origin,
				RedirectURI:      strings.TrimRight(origin, "/") + "/oauth/oidc",
			}, true
		}
	}

	client := settings.GlobalClient()
	if !client.IsComplete() {
		return OIDCResolvedClient{}, false
	}
	return OIDCResolvedClient{
		OIDCClientConfig: client,
		HostMatched:      false,
		Host:             host,
		Origin:           origin,
		RedirectURI:      strings.TrimRight(ServerAddress, "/") + "/oauth/oidc",
	}, true
}

func (settings *OIDCSettings) GlobalClient() OIDCClientConfig {
	if settings == nil {
		return OIDCClientConfig{}
	}
	return OIDCClientConfig{
		ClientId:              settings.ClientId,
		ClientSecret:          settings.ClientSecret,
		WellKnown:             settings.WellKnown,
		AuthorizationEndpoint: settings.AuthorizationEndpoint,
		TokenEndpoint:         settings.TokenEndpoint,
		UserInfoEndpoint:      settings.UserInfoEndpoint,
	}
}

func (settings *OIDCSettings) HasUsableClient() bool {
	if settings == nil {
		return false
	}
	if settings.GlobalClient().IsComplete() {
		return true
	}
	for _, client := range settings.HostClients {
		if client.IsComplete() {
			return true
		}
	}
	return false
}

func (client OIDCClientConfig) IsComplete() bool {
	return strings.TrimSpace(client.ClientId) != "" &&
		strings.TrimSpace(client.ClientSecret) != "" &&
		strings.TrimSpace(client.AuthorizationEndpoint) != "" &&
		strings.TrimSpace(client.TokenEndpoint) != "" &&
		strings.TrimSpace(client.UserInfoEndpoint) != ""
}

func OIDCRequestHost(r *http.Request) string {
	if r == nil {
		return serverAddressHost()
	}
	host := firstHeaderValue(r.Header.Get("X-Forwarded-Host"))
	if host == "" {
		host = r.Host
	}
	if host == "" {
		host = serverAddressHost()
	}
	return strings.TrimSpace(host)
}

func OIDCRequestOrigin(r *http.Request) string {
	host := OIDCRequestHost(r)
	if host == "" {
		return strings.TrimRight(ServerAddress, "/")
	}
	proto := oidcRequestProto(r)
	return fmt.Sprintf("%s://%s", proto, host)
}

func NormalizeOIDCHost(host string) string {
	host = strings.TrimSpace(firstHeaderValue(host))
	if host == "" {
		return ""
	}
	if parsed, err := url.Parse(host); err == nil && parsed.Host != "" {
		host = parsed.Host
	}
	host = strings.Trim(host, "[]")
	host = strings.TrimSuffix(host, ".")
	if withoutPort, _, err := net.SplitHostPort(host); err == nil {
		host = withoutPort
	} else if strings.Count(host, ":") == 1 {
		if idx := strings.LastIndex(host, ":"); idx > 0 {
			host = host[:idx]
		}
	}
	return strings.ToLower(strings.TrimSpace(host))
}

func oidcRequestProto(r *http.Request) string {
	if r != nil {
		if proto := firstHeaderValue(r.Header.Get("X-Forwarded-Proto")); proto != "" {
			return strings.ToLower(proto)
		}
		if r.TLS != nil {
			return "https"
		}
		if r.URL != nil && r.URL.Scheme != "" {
			return strings.ToLower(r.URL.Scheme)
		}
	}
	if parsed, err := url.Parse(ServerAddress); err == nil && parsed.Scheme != "" {
		return strings.ToLower(parsed.Scheme)
	}
	return "https"
}

func firstHeaderValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if idx := strings.Index(value, ","); idx >= 0 {
		value = value[:idx]
	}
	return strings.TrimSpace(value)
}

func serverAddressHost() string {
	if parsed, err := url.Parse(ServerAddress); err == nil && parsed.Host != "" {
		return parsed.Host
	}
	return ""
}
