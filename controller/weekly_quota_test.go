package controller

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type weeklyQuotaAPIResponse struct {
	Success bool                   `json:"success"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
}

func setupWeeklyQuotaControllerTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	gin.SetMode(gin.TestMode)
	originalPhoneVerificationEnabled := common.PhoneVerificationEnabled
	common.UsingSQLite = true
	common.UsingMySQL = false
	common.UsingPostgreSQL = false
	common.RedisEnabled = false
	common.BatchUpdateEnabled = false
	common.PhoneVerificationEnabled = false

	setting := operation_setting.GetWeeklyQuotaSetting()
	originalEnabled := setting.Enabled
	originalAmount := setting.Amount
	t.Cleanup(func() {
		setting.Enabled = originalEnabled
		setting.Amount = originalAmount
		common.PhoneVerificationEnabled = originalPhoneVerificationEnabled
	})
	setting.Enabled = true
	setting.Amount = 500

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	model.DB = db
	model.LOG_DB = db
	require.NoError(t, db.AutoMigrate(&model.User{}, &model.Log{}, &model.WeeklyQuotaClaim{}))

	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})

	return db
}

func createWeeklyQuotaControllerUser(t *testing.T, phoneNumber string) *model.User {
	t.Helper()
	user := &model.User{
		Username:    "weeklyapi",
		Password:    "hashed",
		Status:      common.UserStatusEnabled,
		Role:        common.RoleCommonUser,
		Quota:       100,
		PhoneNumber: phoneNumber,
		CreatedAt:   1_700_000_000,
	}
	require.NoError(t, model.DB.Create(user).Error)
	return user
}

func newWeeklyQuotaRouter(t *testing.T, userId int) *gin.Engine {
	t.Helper()
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("id", userId)
		c.Set("role", common.RoleCommonUser)
		c.Next()
	})
	router.GET("/api/user/weekly_quota", GetWeeklyQuotaStatus)
	router.POST("/api/user/weekly_quota", ClaimWeeklyQuota)
	return router
}

func decodeWeeklyQuotaAPIResponse(t *testing.T, recorder *httptest.ResponseRecorder) weeklyQuotaAPIResponse {
	t.Helper()
	var response weeklyQuotaAPIResponse
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &response))
	return response
}

func TestGetWeeklyQuotaStatusAPI(t *testing.T) {
	setupWeeklyQuotaControllerTestDB(t)
	user := createWeeklyQuotaControllerUser(t, "")
	router := newWeeklyQuotaRouter(t, user.Id)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/user/weekly_quota", nil))

	response := decodeWeeklyQuotaAPIResponse(t, recorder)
	require.True(t, response.Success, response.Message)
	require.Equal(t, "claimable", response.Data["status"])
	require.EqualValues(t, 500, response.Data["amount"])
}

func TestClaimWeeklyQuotaAPIIncrementsQuotaAndRejectsDuplicate(t *testing.T) {
	setupWeeklyQuotaControllerTestDB(t)
	user := createWeeklyQuotaControllerUser(t, "")
	router := newWeeklyQuotaRouter(t, user.Id)

	first := httptest.NewRecorder()
	router.ServeHTTP(first, httptest.NewRequest(http.MethodPost, "/api/user/weekly_quota", nil))
	firstResponse := decodeWeeklyQuotaAPIResponse(t, first)
	require.True(t, firstResponse.Success, firstResponse.Message)
	require.EqualValues(t, 500, firstResponse.Data["quota_awarded"])

	var updated model.User
	require.NoError(t, model.DB.Select("quota").Where("id = ?", user.Id).First(&updated).Error)
	require.Equal(t, 600, updated.Quota)

	second := httptest.NewRecorder()
	router.ServeHTTP(second, httptest.NewRequest(http.MethodPost, "/api/user/weekly_quota", nil))
	secondResponse := decodeWeeklyQuotaAPIResponse(t, second)
	require.False(t, secondResponse.Success)
	require.Contains(t, secondResponse.Message, "已领取")
}

func TestClaimWeeklyQuotaAPIRequiresPhoneWhenEnabled(t *testing.T) {
	setupWeeklyQuotaControllerTestDB(t)
	common.PhoneVerificationEnabled = true
	user := createWeeklyQuotaControllerUser(t, "")
	router := newWeeklyQuotaRouter(t, user.Id)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/user/weekly_quota", nil))

	response := decodeWeeklyQuotaAPIResponse(t, recorder)
	require.False(t, response.Success)
	require.Contains(t, response.Message, "绑定手机号")
}
