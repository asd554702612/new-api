package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func setupTokenAuthIdentityTest(t *testing.T) {
	t.Helper()

	gin.SetMode(gin.TestMode)
	originalDB := model.DB
	originalLogDB := model.LOG_DB
	originalRedisEnabled := common.RedisEnabled
	originalSQLitePath := common.SQLitePath
	originalIsMasterNode := common.IsMasterNode
	originalCasdoorIdentityEnabled := setting.CasdoorIdentityEnabled
	originalCasdoorIdentityApiRequired := setting.CasdoorIdentityApiRequired
	originalSQLDSN, hadSQLDSN := os.LookupEnv("SQL_DSN")
	t.Cleanup(func() {
		if hadSQLDSN {
			require.NoError(t, os.Setenv("SQL_DSN", originalSQLDSN))
		} else {
			require.NoError(t, os.Unsetenv("SQL_DSN"))
		}
		model.DB = originalDB
		model.LOG_DB = originalLogDB
		common.RedisEnabled = originalRedisEnabled
		common.SQLitePath = originalSQLitePath
		common.IsMasterNode = originalIsMasterNode
		setting.CasdoorIdentityEnabled = originalCasdoorIdentityEnabled
		setting.CasdoorIdentityApiRequired = originalCasdoorIdentityApiRequired
	})

	common.SetDatabaseTypes(common.DatabaseTypeSQLite, common.DatabaseTypeSQLite)
	common.RedisEnabled = false
	common.IsMasterNode = false

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	common.SQLitePath = dsn
	require.NoError(t, os.Unsetenv("SQL_DSN"))
	require.NoError(t, model.InitDB())
	model.LOG_DB = model.DB
	require.NoError(t, model.DB.AutoMigrate(&model.User{}, &model.Token{}))
}

func createTokenAuthIdentityUser(t *testing.T, verified bool, ageChecked bool, over16 bool) (int, string) {
	t.Helper()
	user := &model.User{
		Username:           "identity_api_user",
		Role:               common.RoleCommonUser,
		Status:             common.UserStatusEnabled,
		Group:              "default",
		IdentityVerified:   verified,
		IdentityAgeChecked: ageChecked,
		IdentityOver16:     over16,
	}
	require.NoError(t, model.DB.Create(user).Error)

	token := &model.Token{
		UserId:         user.Id,
		Key:            fmt.Sprintf("identitytoken%d", user.Id),
		Status:         common.TokenStatusEnabled,
		UnlimitedQuota: true,
	}
	require.NoError(t, model.DB.Create(token).Error)
	return user.Id, token.Key
}

func performTokenAuthIdentityRequest(tokenKey string, auth gin.HandlerFunc) *httptest.ResponseRecorder {
	router := gin.New()
	router.GET("/v1/chat/completions", auth, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true, "user_id": c.GetInt("id")})
	})

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/chat/completions", nil)
	req.Header.Set("Authorization", "Bearer sk-"+tokenKey)
	router.ServeHTTP(recorder, req)
	return recorder
}

func TestTokenAuthRejectsUnverifiedUserWhenCasdoorIdentityAPIRequired(t *testing.T) {
	setupTokenAuthIdentityTest(t)
	setting.CasdoorIdentityEnabled = true
	setting.CasdoorIdentityApiRequired = true
	_, tokenKey := createTokenAuthIdentityUser(t, false, false, false)

	recorder := performTokenAuthIdentityRequest(tokenKey, TokenAuth())

	require.Equal(t, http.StatusForbidden, recorder.Code)
	require.Contains(t, recorder.Body.String(), "请先完成实名认证后再调用 API")
	require.Contains(t, recorder.Body.String(), "access_denied")
}

func TestTokenAuthRejectsUserWhenIdentityAgeIsUnknown(t *testing.T) {
	setupTokenAuthIdentityTest(t)
	setting.CasdoorIdentityEnabled = true
	setting.CasdoorIdentityApiRequired = true
	_, tokenKey := createTokenAuthIdentityUser(t, true, false, true)

	recorder := performTokenAuthIdentityRequest(tokenKey, TokenAuth())

	require.Equal(t, http.StatusForbidden, recorder.Code)
	require.Contains(t, recorder.Body.String(), "请先完成实名认证后再调用 API")
}

func TestTokenAuthAllowsVerifiedUserWhenCasdoorIdentityAPIRequired(t *testing.T) {
	setupTokenAuthIdentityTest(t)
	setting.CasdoorIdentityEnabled = true
	setting.CasdoorIdentityApiRequired = true
	userID, tokenKey := createTokenAuthIdentityUser(t, true, true, true)

	recorder := performTokenAuthIdentityRequest(tokenKey, TokenAuth())

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), fmt.Sprintf(`"user_id":%d`, userID))
}

func TestTokenAuthAllowsUnverifiedUserWhenCasdoorIdentityAPIRequirementDisabled(t *testing.T) {
	setupTokenAuthIdentityTest(t)
	setting.CasdoorIdentityEnabled = true
	setting.CasdoorIdentityApiRequired = false
	userID, tokenKey := createTokenAuthIdentityUser(t, false, false, false)

	recorder := performTokenAuthIdentityRequest(tokenKey, TokenAuth())

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), fmt.Sprintf(`"user_id":%d`, userID))
}

func TestTokenAuthReadOnlyDoesNotRequireCasdoorIdentity(t *testing.T) {
	setupTokenAuthIdentityTest(t)
	setting.CasdoorIdentityEnabled = true
	setting.CasdoorIdentityApiRequired = true
	userID, tokenKey := createTokenAuthIdentityUser(t, false, false, false)

	recorder := performTokenAuthIdentityRequest(tokenKey, TokenAuthReadOnly())

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), fmt.Sprintf(`"user_id":%d`, userID))
}
