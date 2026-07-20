package controller

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/middleware"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/oauth"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type oauthIdentityAPIResponse struct {
	Success bool           `json:"success"`
	Message string         `json:"message"`
	Data    map[string]any `json:"data"`
}

type fakeOAuthProvider struct{}

func (p fakeOAuthProvider) GetName() string { return "Fake" }
func (p fakeOAuthProvider) IsEnabled() bool { return true }
func (p fakeOAuthProvider) ExchangeToken(ctx context.Context, code string, c *gin.Context) (*oauth.OAuthToken, error) {
	return nil, nil
}
func (p fakeOAuthProvider) GetUserInfo(ctx context.Context, token *oauth.OAuthToken) (*oauth.OAuthUser, error) {
	return nil, nil
}
func (p fakeOAuthProvider) IsUserIDTaken(providerUserID string) bool { return false }
func (p fakeOAuthProvider) FillUserByProviderID(user *model.User, providerUserID string) error {
	return nil
}
func (p fakeOAuthProvider) SetProviderUserID(user *model.User, providerUserID string) {}
func (p fakeOAuthProvider) GetProviderPrefix() string                                 { return "fake_" }

func setupOAuthIdentityControllerTest(t *testing.T, identityVerified *atomic.Bool) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)
	common.SetDatabaseTypes(common.DatabaseTypeSQLite, common.DatabaseTypeSQLite)
	common.RedisEnabled = false
	common.RegisterEnabled = false
	common.PasswordRegisterEnabled = false
	common.PasswordLoginEnabled = true
	common.EmailVerificationEnabled = false
	common.PhoneVerificationEnabled = false
	constant.GenerateDefaultToken = false

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	model.DB = db
	model.LOG_DB = db
	require.NoError(t, db.AutoMigrate(&model.User{}, &model.Log{}, &model.Option{}))
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})

	userInfoServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "Bearer oidc-access-token", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"sub":"casdoor-sub","email":"alice@example.com","preferred_username":"alice","name":"Alice"}`))
	}))
	t.Cleanup(userInfoServer.Close)

	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseForm())
		require.Equal(t, "oidc-client", r.Form.Get("client_id"))
		require.Equal(t, "oidc-secret", r.Form.Get("client_secret"))
		require.Equal(t, "authorization-code", r.Form.Get("code"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"oidc-access-token","token_type":"Bearer","expires_in":3600}`))
	}))
	t.Cleanup(tokenServer.Close)

	casdoorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/external/user/sync", r.URL.Path)
		require.Equal(t, "oidc-client", r.Header.Get("X-Casdoor-App-Id"))
		w.Header().Set("Content-Type", "application/json")
		if identityVerified.Load() {
			_, _ = w.Write([]byte(`{"status":"ok","data":{"userId":"casdoor-sub","owner":"gepin","name":"alice","displayName":"Alice","email":"alice@example.com","phone":"13800000000","isVerified":true,"ageChecked":true,"isOver16":true}}`))
			return
		}
		_, _ = w.Write([]byte(`{"status":"ok","data":{"userId":"casdoor-sub","owner":"gepin","name":"alice","displayName":"Alice","email":"alice@example.com","phone":"13800000000","isVerified":false,"ageChecked":false,"isOver16":false}}`))
	}))
	t.Cleanup(casdoorServer.Close)

	originalOIDC := *system_setting.GetOIDCSettings()
	originalServerAddress := system_setting.ServerAddress
	originalCasdoorBaseURL := setting.CasdoorBaseURL
	originalCasdoorClientID := setting.CasdoorClientID
	originalCasdoorClientSecret := setting.CasdoorClientSecret
	originalCasdoorIdentityEnabled := setting.CasdoorIdentityEnabled
	originalCasdoorIdentityCallbackURL := setting.CasdoorIdentityCallbackURL
	t.Cleanup(func() {
		*system_setting.GetOIDCSettings() = originalOIDC
		system_setting.ServerAddress = originalServerAddress
		setting.CasdoorBaseURL = originalCasdoorBaseURL
		setting.CasdoorClientID = originalCasdoorClientID
		setting.CasdoorClientSecret = originalCasdoorClientSecret
		setting.CasdoorIdentityEnabled = originalCasdoorIdentityEnabled
		setting.CasdoorIdentityCallbackURL = originalCasdoorIdentityCallbackURL
	})

	system_setting.ServerAddress = "https://token.gepinkeji.com"
	*system_setting.GetOIDCSettings() = system_setting.OIDCSettings{
		Enabled:               true,
		ClientId:              "oidc-client",
		ClientSecret:          "oidc-secret",
		AuthorizationEndpoint: "https://login.gepinkeji.com/login/oauth/authorize",
		TokenEndpoint:         tokenServer.URL,
		UserInfoEndpoint:      userInfoServer.URL,
	}
	setting.CasdoorBaseURL = casdoorServer.URL
	setting.CasdoorClientID = "oidc-client"
	setting.CasdoorClientSecret = "oidc-secret"
	setting.CasdoorIdentityEnabled = true
	setting.CasdoorIdentityCallbackURL = "https://token.gepinkeji.com/identity/callback"

	router := gin.New()
	router.Use(sessions.Sessions("session", cookie.NewStore([]byte("oauth-identity-test"))))
	router.GET("/seed", func(c *gin.Context) {
		session := sessions.Default(c)
		session.Set("oauth_state", "oauth-state")
		require.NoError(t, session.Save())
		c.Status(http.StatusNoContent)
	})
	router.GET("/session", func(c *gin.Context) {
		session := sessions.Default(c)
		c.JSON(http.StatusOK, gin.H{
			"username":             session.Get("username"),
			"identity_state":       session.Get("casdoor_identity_state"),
			"identity_return_path": session.Get("casdoor_identity_return_path"),
		})
	})
	router.GET("/login-as", func(c *gin.Context) {
		userID, err := strconv.Atoi(c.Query("id"))
		require.NoError(t, err)
		user, err := model.GetUserById(userID, true)
		require.NoError(t, err)
		session := sessions.Default(c)
		session.Set("id", user.Id)
		session.Set("username", user.Username)
		session.Set("role", user.Role)
		session.Set("status", user.Status)
		session.Set("group", user.Group)
		require.NoError(t, session.Save())
		c.Status(http.StatusNoContent)
	})
	router.GET("/api/oauth/:provider", HandleOAuth)
	userRoute := router.Group("/api/user")
	userRoute.Use(middleware.UserAuth())
	{
		userRoute.GET("/self", GetSelf)
		userRoute.POST("/identity/sync", HandleUserIdentitySync)
		userRoute.POST("/identity/verification", HandleUserIdentityVerification)
	}
	router.GET("/identity/callback", HandleCasdoorIdentityCallback)
	return router
}

func seedOAuthIdentitySession(t *testing.T, router *gin.Engine) []*http.Cookie {
	t.Helper()
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/seed", nil))
	require.Equal(t, http.StatusNoContent, recorder.Code)
	return recorder.Result().Cookies()
}

func addCookies(req *http.Request, cookies []*http.Cookie) {
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
}

func mergeResponseCookies(cookies []*http.Cookie, recorder *httptest.ResponseRecorder) []*http.Cookie {
	merged := append([]*http.Cookie{}, cookies...)
	for _, cookie := range recorder.Result().Cookies() {
		replaced := false
		for i, existing := range merged {
			if existing.Name == cookie.Name {
				merged[i] = cookie
				replaced = true
				break
			}
		}
		if !replaced {
			merged = append(merged, cookie)
		}
	}
	return merged
}

func performOAuthIdentityGET(t *testing.T, router *gin.Engine, path string, cookies []*http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Host = "token.gepinkeji.com"
	req.Header.Set("X-Forwarded-Proto", "https")
	addCookies(req, cookies)
	router.ServeHTTP(recorder, req)
	return recorder
}

func performOAuthIdentityUserRequest(t *testing.T, router *gin.Engine, method string, path string, cookies []*http.Cookie, userID int) *httptest.ResponseRecorder {
	t.Helper()
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, nil)
	req.Host = "token.gepinkeji.com"
	req.Header.Set("X-Forwarded-Proto", "https")
	req.Header.Set("New-Api-User", strconv.Itoa(userID))
	addCookies(req, cookies)
	router.ServeHTTP(recorder, req)
	return recorder
}

func decodeOAuthIdentityResponse(t *testing.T, recorder *httptest.ResponseRecorder) oauthIdentityAPIResponse {
	t.Helper()
	var response oauthIdentityAPIResponse
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &response))
	return response
}

func createOAuthIdentityUser(t *testing.T, username string, oidcID string, verified bool, ageChecked bool, over16 bool, syncedAt int64) *model.User {
	t.Helper()
	user := &model.User{
		Username:           username,
		Password:           "password",
		DisplayName:        username,
		Role:               common.RoleCommonUser,
		Status:             common.UserStatusEnabled,
		Group:              "default",
		OidcId:             oidcID,
		IdentityVerified:   verified,
		IdentityAgeChecked: ageChecked,
		IdentityOver16:     over16,
		IdentitySyncedAt:   syncedAt,
	}
	require.NoError(t, model.DB.Create(user).Error)
	return user
}

func seedLoggedInOAuthIdentitySession(t *testing.T, router *gin.Engine, userID int) []*http.Cookie {
	t.Helper()
	recorder := performOAuthIdentityGET(t, router, "/login-as?id="+strconv.Itoa(userID), nil)
	require.Equal(t, http.StatusNoContent, recorder.Code)
	return recorder.Result().Cookies()
}

func TestGetSelfIncludesIdentitySnapshot(t *testing.T) {
	var verified atomic.Bool
	router := setupOAuthIdentityControllerTest(t, &verified)
	user := createOAuthIdentityUser(t, "self_identity_user", "casdoor-sub", true, true, true, 12345)
	cookies := seedLoggedInOAuthIdentitySession(t, router, user.Id)

	recorder := performOAuthIdentityUserRequest(t, router, http.MethodGet, "/api/user/self", cookies, user.Id)
	response := decodeOAuthIdentityResponse(t, recorder)

	require.True(t, response.Success, response.Message)
	require.Equal(t, true, response.Data["identity_verified"])
	require.Equal(t, true, response.Data["identity_age_checked"])
	require.Equal(t, true, response.Data["identity_over16"])
	require.EqualValues(t, 12345, response.Data["identity_synced_at"])
}

func TestUserIdentitySyncUpdatesSnapshot(t *testing.T) {
	var verified atomic.Bool
	verified.Store(true)
	router := setupOAuthIdentityControllerTest(t, &verified)
	user := createOAuthIdentityUser(t, "sync_identity_user", "casdoor-sub", false, false, false, 0)
	cookies := seedLoggedInOAuthIdentitySession(t, router, user.Id)

	recorder := performOAuthIdentityUserRequest(t, router, http.MethodPost, "/api/user/identity/sync", cookies, user.Id)
	response := decodeOAuthIdentityResponse(t, recorder)

	require.True(t, response.Success, response.Message)
	require.Equal(t, true, response.Data["identity_verified"])
	require.Equal(t, true, response.Data["identity_age_checked"])
	require.Equal(t, true, response.Data["identity_over16"])
	require.NotZero(t, response.Data["identity_synced_at"])

	var updated model.User
	require.NoError(t, model.DB.First(&updated, user.Id).Error)
	require.True(t, updated.IdentityVerified)
	require.True(t, updated.IdentityAgeChecked)
	require.True(t, updated.IdentityOver16)
	require.NotZero(t, updated.IdentitySyncedAt)
}

func TestUserIdentitySyncFallsBackToOIDCClientCredentials(t *testing.T) {
	var verified atomic.Bool
	verified.Store(true)
	router := setupOAuthIdentityControllerTest(t, &verified)
	setting.CasdoorClientID = ""
	setting.CasdoorClientSecret = ""
	user := createOAuthIdentityUser(t, "sync_identity_oidc_fallback_user", "casdoor-sub", false, false, false, 0)
	cookies := seedLoggedInOAuthIdentitySession(t, router, user.Id)

	recorder := performOAuthIdentityUserRequest(t, router, http.MethodPost, "/api/user/identity/sync", cookies, user.Id)
	response := decodeOAuthIdentityResponse(t, recorder)

	require.True(t, response.Success, response.Message)
	require.Equal(t, true, response.Data["identity_verified"])
	require.Equal(t, true, response.Data["identity_age_checked"])
	require.Equal(t, true, response.Data["identity_over16"])
}

func TestUserIdentitySyncFailsWhenOIDCNotBound(t *testing.T) {
	var verified atomic.Bool
	router := setupOAuthIdentityControllerTest(t, &verified)
	user := createOAuthIdentityUser(t, "sync_unbound_user", "", false, false, false, 0)
	cookies := seedLoggedInOAuthIdentitySession(t, router, user.Id)

	recorder := performOAuthIdentityUserRequest(t, router, http.MethodPost, "/api/user/identity/sync", cookies, user.Id)
	response := decodeOAuthIdentityResponse(t, recorder)

	require.False(t, response.Success)
	require.NotEmpty(t, response.Message)
}

func TestUserIdentitySyncFailsClosedWhenCasdoorSyncFails(t *testing.T) {
	var verified atomic.Bool
	verified.Store(true)
	router := setupOAuthIdentityControllerTest(t, &verified)
	user := createOAuthIdentityUser(t, "sync_failure_user", "casdoor-sub", false, false, false, 0)
	cookies := seedLoggedInOAuthIdentitySession(t, router, user.Id)

	setting.CasdoorClientSecret = ""
	oidcSettings := system_setting.GetOIDCSettings()
	oidcSettings.ClientSecret = ""
	recorder := performOAuthIdentityUserRequest(t, router, http.MethodPost, "/api/user/identity/sync", cookies, user.Id)
	response := decodeOAuthIdentityResponse(t, recorder)

	require.False(t, response.Success)

	var updated model.User
	require.NoError(t, model.DB.First(&updated, user.Id).Error)
	require.False(t, updated.IdentityVerified)
	require.False(t, updated.IdentityAgeChecked)
	require.False(t, updated.IdentityOver16)
	require.Zero(t, updated.IdentitySyncedAt)
}

func TestUserIdentityVerificationRedirectsAndCallbackReturnsToPersonal(t *testing.T) {
	var verified atomic.Bool
	verified.Store(false)
	router := setupOAuthIdentityControllerTest(t, &verified)
	user := createOAuthIdentityUser(t, "verification_user", "casdoor-sub", false, false, false, 0)
	cookies := seedLoggedInOAuthIdentitySession(t, router, user.Id)

	recorder := performOAuthIdentityUserRequest(t, router, http.MethodPost, "/api/user/identity/verification", cookies, user.Id)
	response := decodeOAuthIdentityResponse(t, recorder)

	require.True(t, response.Success, response.Message)
	require.Equal(t, "identity_required", response.Data["action"])
	redirectURL, ok := response.Data["redirect_url"].(string)
	require.True(t, ok)
	require.NotEmpty(t, redirectURL)

	cookies = mergeResponseCookies(cookies, recorder)
	sessionRecorder := performOAuthIdentityGET(t, router, "/session", cookies)
	var sessionData map[string]any
	require.NoError(t, common.Unmarshal(sessionRecorder.Body.Bytes(), &sessionData))
	require.Equal(t, "/console/personal", sessionData["identity_return_path"])
	require.NotEmpty(t, sessionData["identity_state"])

	parsed, err := url.Parse(redirectURL)
	require.NoError(t, err)
	state := parsed.Query().Get("state")
	require.NotEmpty(t, state)

	verified.Store(true)
	callbackRecorder := performOAuthIdentityGET(t, router, "/identity/callback?state="+url.QueryEscape(state), cookies)
	require.Equal(t, http.StatusFound, callbackRecorder.Code)
	require.Equal(t, "/console/personal", callbackRecorder.Header().Get("Location"))
}

func TestOIDCIdentityLoginCreatesUserWhenRegistrationClosed(t *testing.T) {
	var verified atomic.Bool
	verified.Store(true)
	router := setupOAuthIdentityControllerTest(t, &verified)
	cookies := seedOAuthIdentitySession(t, router)

	recorder := performOAuthIdentityGET(t, router, "/api/oauth/oidc?code=authorization-code&state=oauth-state", cookies)
	response := decodeOAuthIdentityResponse(t, recorder)

	require.True(t, response.Success, response.Message)
	require.Equal(t, "alice", response.Data["username"])

	var user model.User
	require.NoError(t, model.DB.Where("oidc_id = ?", "casdoor-sub").First(&user).Error)
	require.Equal(t, "alice", user.Username)
	require.True(t, user.IdentityVerified)
	require.True(t, user.IdentityAgeChecked)
	require.True(t, user.IdentityOver16)
	require.NotZero(t, user.IdentitySyncedAt)
}

func TestOIDCLoginCreatesUserWhenRegistrationClosedAndIdentityDisabled(t *testing.T) {
	var verified atomic.Bool
	router := setupOAuthIdentityControllerTest(t, &verified)
	setting.CasdoorIdentityEnabled = false
	cookies := seedOAuthIdentitySession(t, router)

	recorder := performOAuthIdentityGET(t, router, "/api/oauth/oidc?code=authorization-code&state=oauth-state", cookies)
	response := decodeOAuthIdentityResponse(t, recorder)

	require.True(t, response.Success, response.Message)
	require.Equal(t, "alice", response.Data["username"])

	var user model.User
	require.NoError(t, model.DB.Where("oidc_id = ?", "casdoor-sub").First(&user).Error)
	require.Equal(t, "alice", user.Username)
	require.False(t, user.IdentityVerified)
	require.False(t, user.IdentityAgeChecked)
	require.False(t, user.IdentityOver16)
	require.Zero(t, user.IdentitySyncedAt)
}

func TestOIDCIdentityUnverifiedRedirectsAndCallbackRechecksBeforeLogin(t *testing.T) {
	var verified atomic.Bool
	verified.Store(false)
	router := setupOAuthIdentityControllerTest(t, &verified)
	cookies := seedOAuthIdentitySession(t, router)

	recorder := performOAuthIdentityGET(t, router, "/api/oauth/oidc?code=authorization-code&state=oauth-state", cookies)
	response := decodeOAuthIdentityResponse(t, recorder)

	require.True(t, response.Success, response.Message)
	require.Equal(t, "identity_required", response.Data["action"])
	redirectURL, ok := response.Data["redirect_url"].(string)
	require.True(t, ok)
	require.NotEmpty(t, redirectURL)

	cookies = mergeResponseCookies(cookies, recorder)
	sessionRecorder := performOAuthIdentityGET(t, router, "/session", cookies)
	var sessionData map[string]any
	require.NoError(t, common.Unmarshal(sessionRecorder.Body.Bytes(), &sessionData))
	require.Nil(t, sessionData["username"])
	require.NotEmpty(t, sessionData["identity_state"])

	parsed, err := url.Parse(redirectURL)
	require.NoError(t, err)
	state := parsed.Query().Get("state")
	require.NotEmpty(t, state)

	verified.Store(true)
	callbackRecorder := performOAuthIdentityGET(t, router, "/identity/callback?state="+url.QueryEscape(state), cookies)
	require.Equal(t, http.StatusFound, callbackRecorder.Code)
	require.Equal(t, "/console/token", callbackRecorder.Header().Get("Location"))

	cookies = mergeResponseCookies(cookies, callbackRecorder)
	sessionRecorder = performOAuthIdentityGET(t, router, "/session", cookies)
	require.NoError(t, common.Unmarshal(sessionRecorder.Body.Bytes(), &sessionData))
	require.Equal(t, "alice", sessionData["username"])

	var user model.User
	require.NoError(t, model.DB.Where("oidc_id = ?", "casdoor-sub").First(&user).Error)
	require.True(t, user.IdentityVerified)
	require.True(t, user.IdentityAgeChecked)
	require.True(t, user.IdentityOver16)
}

func TestOIDCIdentityCallbackRejectsDisabledLocalUser(t *testing.T) {
	var verified atomic.Bool
	verified.Store(false)
	router := setupOAuthIdentityControllerTest(t, &verified)
	cookies := seedOAuthIdentitySession(t, router)

	recorder := performOAuthIdentityGET(t, router, "/api/oauth/oidc?code=authorization-code&state=oauth-state", cookies)
	response := decodeOAuthIdentityResponse(t, recorder)
	require.True(t, response.Success, response.Message)

	redirectURL, ok := response.Data["redirect_url"].(string)
	require.True(t, ok)
	parsed, err := url.Parse(redirectURL)
	require.NoError(t, err)
	state := parsed.Query().Get("state")
	require.NotEmpty(t, state)

	require.NoError(t, model.DB.Model(&model.User{}).Where("oidc_id = ?", "casdoor-sub").Update("status", common.UserStatusDisabled).Error)

	cookies = mergeResponseCookies(cookies, recorder)
	verified.Store(true)
	callbackRecorder := performOAuthIdentityGET(t, router, "/identity/callback?state="+url.QueryEscape(state), cookies)
	require.Equal(t, http.StatusOK, callbackRecorder.Code)

	response = decodeOAuthIdentityResponse(t, callbackRecorder)
	require.False(t, response.Success)

	sessionRecorder := performOAuthIdentityGET(t, router, "/session", cookies)
	var sessionData map[string]any
	require.NoError(t, common.Unmarshal(sessionRecorder.Body.Bytes(), &sessionData))
	require.Nil(t, sessionData["username"])
}

func TestOAuthRegistrationDisabledStillBlocksNonOIDCProvider(t *testing.T) {
	common.RegisterEnabled = false
	user, err := findOrCreateOAuthUser(
		nil,
		fakeOAuthProvider{},
		&oauth.OAuthUser{ProviderUserID: "provider-user", Username: "provider-user"},
		nil,
		false,
	)

	require.Nil(t, user)
	require.IsType(t, &OAuthRegistrationDisabledError{}, err)
}
