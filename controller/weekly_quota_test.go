package controller

import (
	"bytes"
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
	originalPlanId := setting.PlanId
	originalPeriodDays := setting.PeriodDays
	t.Cleanup(func() {
		setting.Enabled = originalEnabled
		setting.Amount = originalAmount
		setting.PlanId = originalPlanId
		setting.PeriodDays = originalPeriodDays
		common.PhoneVerificationEnabled = originalPhoneVerificationEnabled
	})
	setting.Enabled = true
	setting.Amount = 0
	setting.PlanId = 8101
	setting.PeriodDays = 7

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	model.DB = db
	model.LOG_DB = db
	require.NoError(t, db.AutoMigrate(&model.User{}, &model.Log{}, &model.WeeklyQuotaClaim{}, &model.SubscriptionPlan{}, &model.UserSubscription{}))
	require.NoError(t, db.Create(&model.SubscriptionPlan{
		Id:            8101,
		Title:         "Controller Gift Plan",
		Subtitle:      "Gift",
		PriceAmount:   0,
		Currency:      "USD",
		DurationUnit:  model.SubscriptionDurationDay,
		DurationValue: 30,
		Enabled:       false,
		TotalAmount:   67890,
	}).Error)
	require.NoError(t, db.Create(&model.SubscriptionPlan{
		Id:            8102,
		Title:         "Forged Plan",
		PriceAmount:   0,
		Currency:      "USD",
		DurationUnit:  model.SubscriptionDurationDay,
		DurationValue: 30,
		Enabled:       true,
		TotalAmount:   99999,
	}).Error)

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

func TestGetWeeklyQuotaStatusAPIRequiresPhoneWhenUnbound(t *testing.T) {
	setupWeeklyQuotaControllerTestDB(t)
	user := createWeeklyQuotaControllerUser(t, "")
	router := newWeeklyQuotaRouter(t, user.Id)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/user/weekly_quota", nil))

	response := decodeWeeklyQuotaAPIResponse(t, recorder)
	require.True(t, response.Success, response.Message)
	require.Equal(t, "phone_required", response.Data["status"])
	require.EqualValues(t, 8101, response.Data["plan_id"])
	require.EqualValues(t, 7, response.Data["period_days"])
}

func TestGetWeeklyQuotaStatusAPIReturnsGiftPlanWhenPhoneBound(t *testing.T) {
	setupWeeklyQuotaControllerTestDB(t)
	user := createWeeklyQuotaControllerUser(t, "+8613800138000")
	router := newWeeklyQuotaRouter(t, user.Id)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/user/weekly_quota", nil))

	response := decodeWeeklyQuotaAPIResponse(t, recorder)
	require.True(t, response.Success, response.Message)
	require.Equal(t, "claimable", response.Data["status"])
	require.EqualValues(t, 8101, response.Data["plan_id"])
	require.EqualValues(t, 7, response.Data["period_days"])
	plan := response.Data["plan"].(map[string]interface{})
	require.EqualValues(t, 8101, plan["id"])
	require.Equal(t, "Controller Gift Plan", plan["title"])
	require.EqualValues(t, 67890, plan["total_amount"])
}

func TestClaimWeeklyQuotaAPICreatesConfiguredGiftSubscriptionAndRejectsForgedPlan(t *testing.T) {
	setupWeeklyQuotaControllerTestDB(t)
	user := createWeeklyQuotaControllerUser(t, "+8613800138000")
	router := newWeeklyQuotaRouter(t, user.Id)

	first := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/user/weekly_quota", bytes.NewBufferString(`{"plan_id":8102}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(first, req)
	firstResponse := decodeWeeklyQuotaAPIResponse(t, first)
	require.True(t, firstResponse.Success, firstResponse.Message)
	require.EqualValues(t, 8101, firstResponse.Data["plan_id"])
	require.NotZero(t, firstResponse.Data["subscription_id"])

	var updated model.User
	require.NoError(t, model.DB.Select("quota").Where("id = ?", user.Id).First(&updated).Error)
	require.Equal(t, 100, updated.Quota)

	var sub model.UserSubscription
	require.NoError(t, model.DB.Where("user_id = ?", user.Id).First(&sub).Error)
	require.Equal(t, 8101, sub.PlanId)
	require.Equal(t, model.SubscriptionSourceGiftClaim, sub.Source)

	second := httptest.NewRecorder()
	router.ServeHTTP(second, httptest.NewRequest(http.MethodPost, "/api/user/weekly_quota", nil))
	secondResponse := decodeWeeklyQuotaAPIResponse(t, second)
	require.False(t, secondResponse.Success)
	require.Contains(t, secondResponse.Message, "已领取")
}

func TestClaimWeeklyQuotaAPIRequiresPhoneWhenUnbound(t *testing.T) {
	setupWeeklyQuotaControllerTestDB(t)
	user := createWeeklyQuotaControllerUser(t, "")
	router := newWeeklyQuotaRouter(t, user.Id)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/user/weekly_quota", nil))

	response := decodeWeeklyQuotaAPIResponse(t, recorder)
	require.False(t, response.Success)
	require.Contains(t, response.Message, "绑定手机号")
}
