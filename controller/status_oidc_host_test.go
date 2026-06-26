package controller

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupStatusOIDCHostTest(t *testing.T) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	common.UsingSQLite = true
	common.UsingMySQL = false
	common.UsingPostgreSQL = false
	common.RedisEnabled = false

	dsn := "file:" + strings.ReplaceAll(t.Name(), "/", "_") + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	model.DB = db
	model.LOG_DB = db
	require.NoError(t, db.AutoMigrate(&model.PaymentActivityConfig{}))
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})

	originalOIDC := *system_setting.GetOIDCSettings()
	originalServerAddress := system_setting.ServerAddress
	t.Cleanup(func() {
		*system_setting.GetOIDCSettings() = originalOIDC
		system_setting.ServerAddress = originalServerAddress
	})
}

func TestGetStatusReturnsOIDCClientForForwardedHost(t *testing.T) {
	setupStatusOIDCHostTest(t)
	system_setting.ServerAddress = "https://global.example"
	*system_setting.GetOIDCSettings() = system_setting.OIDCSettings{
		Enabled:               true,
		ClientId:              "global-client",
		ClientSecret:          "global-secret",
		AuthorizationEndpoint: "https://login.example/global/auth",
		TokenEndpoint:         "https://login.example/global/token",
		UserInfoEndpoint:      "https://login.example/global/userinfo",
		HostClients: map[string]system_setting.OIDCClientConfig{
			"token.gepinkeji.com": {
				ClientId:              "gepin-client",
				ClientSecret:          "gepin-secret",
				AuthorizationEndpoint: "https://login.example/gepin/auth",
				TokenEndpoint:         "https://login.example/gepin/token",
				UserInfoEndpoint:      "https://login.example/gepin/userinfo",
			},
			"token.gptk.cc.cd": {
				ClientId:              "gptk-client",
				ClientSecret:          "gptk-secret",
				AuthorizationEndpoint: "https://login.example/gptk/auth",
				TokenEndpoint:         "https://login.example/gptk/token",
				UserInfoEndpoint:      "https://login.example/gptk/userinfo",
			},
		},
	}

	router := gin.New()
	router.GET("/api/status", GetStatus)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	req.Host = "internal.local"
	req.Header.Set("X-Forwarded-Host", "token.gptk.cc.cd")
	req.Header.Set("X-Forwarded-Proto", "https")
	router.ServeHTTP(recorder, req)

	var response struct {
		Success bool           `json:"success"`
		Data    map[string]any `json:"data"`
	}
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &response))
	require.True(t, response.Success)
	require.Equal(t, true, response.Data["oidc_enabled"])
	require.Equal(t, "gptk-client", response.Data["oidc_client_id"])
	require.Equal(t, "https://login.example/gptk/auth", response.Data["oidc_authorization_endpoint"])
}

func TestGetStatusFallsBackToGlobalOIDCClientForUnknownHost(t *testing.T) {
	setupStatusOIDCHostTest(t)
	system_setting.ServerAddress = "https://global.example"
	*system_setting.GetOIDCSettings() = system_setting.OIDCSettings{
		Enabled:               true,
		ClientId:              "global-client",
		ClientSecret:          "global-secret",
		AuthorizationEndpoint: "https://login.example/global/auth",
		TokenEndpoint:         "https://login.example/global/token",
		UserInfoEndpoint:      "https://login.example/global/userinfo",
		HostClients: map[string]system_setting.OIDCClientConfig{
			"token.gepinkeji.com": {
				ClientId:              "gepin-client",
				ClientSecret:          "gepin-secret",
				AuthorizationEndpoint: "https://login.example/gepin/auth",
				TokenEndpoint:         "https://login.example/gepin/token",
				UserInfoEndpoint:      "https://login.example/gepin/userinfo",
			},
		},
	}

	router := gin.New()
	router.GET("/api/status", GetStatus)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	req.Host = "unknown.example"
	router.ServeHTTP(recorder, req)

	var response struct {
		Success bool           `json:"success"`
		Data    map[string]any `json:"data"`
	}
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &response))
	require.True(t, response.Success)
	require.Equal(t, true, response.Data["oidc_enabled"])
	require.Equal(t, "global-client", response.Data["oidc_client_id"])
	require.Equal(t, "https://login.example/global/auth", response.Data["oidc_authorization_endpoint"])
}
