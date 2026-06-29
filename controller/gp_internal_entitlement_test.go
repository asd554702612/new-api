package controller

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupGpInternalEntitlementTestDB(t *testing.T) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	common.SetDatabaseTypes(common.DatabaseTypeSQLite, common.DatabaseTypeSQLite)
	common.RedisEnabled = false

	dsn := "file:" + strings.ReplaceAll(t.Name(), "/", "_") + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	model.DB = db
	model.LOG_DB = db
	require.NoError(t, db.AutoMigrate(&model.User{}, &model.SubscriptionPlan{}, &model.UserSubscription{}))
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})
}

func buildGpInternalEntitlementRouter() *gin.Engine {
	router := gin.New()
	router.GET("/api/gp/internal/entitlement/:user_id", GetGpInternalEntitlement)
	return router
}

func TestGetGpInternalEntitlementRejectsMissingSecret(t *testing.T) {
	setupGpInternalEntitlementTestDB(t)
	t.Setenv("GP_INTERNAL_SHARED_SECRET", "test-secret")

	recorder := httptest.NewRecorder()
	buildGpInternalEntitlementRouter().ServeHTTP(
		recorder,
		httptest.NewRequest(http.MethodGet, "/api/gp/internal/entitlement/42", nil),
	)

	require.Equal(t, http.StatusUnauthorized, recorder.Code)
	var response struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &response))
	require.False(t, response.Success)
	require.Contains(t, response.Message, "GP internal secret")
}

func TestGetGpInternalEntitlementReturnsWalletPlansAndSubscriptions(t *testing.T) {
	setupGpInternalEntitlementTestDB(t)
	t.Setenv("GP_INTERNAL_SHARED_SECRET", "test-secret")

	user := model.User{
		Id:        42,
		Username:  "gp-user",
		Quota:     880000,
		UsedQuota: 120000,
		Status:    common.UserStatusEnabled,
	}
	user.SetSetting(dto.UserSetting{BillingPreference: "wallet_first"})
	require.NoError(t, model.DB.Create(&user).Error)

	now := time.Now().Unix()
	plan := model.SubscriptionPlan{
		Id:            77,
		Title:         "GP Pro",
		PriceAmount:   19.9,
		Currency:      "USD",
		DurationUnit:  model.SubscriptionDurationMonth,
		DurationValue: 1,
		Enabled:       true,
		TotalAmount:   5000000,
		UpgradeGroup:  "gp-pro",
	}
	require.NoError(t, model.DB.Create(&plan).Error)
	require.NoError(t, model.DB.Create(&model.UserSubscription{
		UserId:      user.Id,
		PlanId:      plan.Id,
		Status:      "active",
		AmountTotal: plan.TotalAmount,
		AmountUsed:  1000000,
		StartTime:   now - 60,
		EndTime:     now + 3600,
		Source:      "balance",
	}).Error)

	request := httptest.NewRequest(http.MethodGet, "/api/gp/internal/entitlement/42", nil)
	request.Header.Set("X-GP-Internal-Secret", "test-secret")
	recorder := httptest.NewRecorder()
	buildGpInternalEntitlementRouter().ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	var response struct {
		Success bool `json:"success"`
		Data    struct {
			User struct {
				ID        int    `json:"id"`
				Username  string `json:"username"`
				Quota     int    `json:"quota"`
				UsedQuota int    `json:"used_quota"`
			} `json:"user"`
			BillingPreference string `json:"billing_preference"`
			Subscriptions     struct {
				Active []model.SubscriptionSummary `json:"active"`
				All    []model.SubscriptionSummary `json:"all"`
			} `json:"subscriptions"`
			Plans []SubscriptionPlanDTO `json:"plans"`
		} `json:"data"`
	}
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &response))
	require.True(t, response.Success)
	require.Equal(t, 42, response.Data.User.ID)
	require.Equal(t, "gp-user", response.Data.User.Username)
	require.Equal(t, 880000, response.Data.User.Quota)
	require.Equal(t, 120000, response.Data.User.UsedQuota)
	require.Equal(t, "wallet_first", response.Data.BillingPreference)
	require.Len(t, response.Data.Subscriptions.Active, 1)
	require.Len(t, response.Data.Subscriptions.All, 1)
	require.Len(t, response.Data.Plans, 1)
	require.Equal(t, "GP Pro", response.Data.Plans[0].Plan.Title)
}
