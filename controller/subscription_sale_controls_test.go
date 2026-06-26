package controller

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupSubscriptionSaleControllerTestDB(t *testing.T) {
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
	require.NoError(t, db.AutoMigrate(&model.User{}, &model.SubscriptionPlan{}, &model.SubscriptionOrder{}, &model.UserSubscription{}))
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})
}

func TestGetSubscriptionPlansReturnsSaleAvailability(t *testing.T) {
	setupSubscriptionSaleControllerTestDB(t)
	confirmPaymentComplianceForTest(t)

	now := time.Now()
	require.NoError(t, model.DB.Create(&model.SubscriptionPlan{
		Title:              "Limited",
		PriceAmount:        9.99,
		Currency:           "USD",
		DurationUnit:       model.SubscriptionDurationMonth,
		DurationValue:      1,
		Enabled:            true,
		DailyPurchaseLimit: 1,
		DailySaleStartsAt:  "09:00",
		DailySaleEndsAt:    "23:59",
		WeeklySaleDays:     "[]",
		TotalAmount:        1000,
	}).Error)
	var plan model.SubscriptionPlan
	require.NoError(t, model.DB.First(&plan).Error)
	require.NoError(t, model.DB.Create(&model.SubscriptionOrder{
		UserId:          10,
		PlanId:          plan.Id,
		Money:           plan.PriceAmount,
		TradeNo:         "controller-sale-success",
		PaymentMethod:   model.PaymentMethodBalance,
		PaymentProvider: model.PaymentProviderBalance,
		Status:          common.TopUpStatusSuccess,
		CreateTime:      now.Add(-time.Hour).Unix(),
		CompleteTime:    now.Add(-time.Minute).Unix(),
	}).Error)

	router := gin.New()
	router.GET("/api/subscription/plans", func(c *gin.Context) {
		c.Set("id", 10)
		GetSubscriptionPlans(c)
	})
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/subscription/plans", nil))

	var response struct {
		Success bool `json:"success"`
		Data    []struct {
			SaleAvailability model.SubscriptionPlanSaleAvailability `json:"sale_availability"`
		} `json:"data"`
	}
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &response))
	require.True(t, response.Success)
	require.Len(t, response.Data, 1)
	require.False(t, response.Data[0].SaleAvailability.Available)
	require.Equal(t, model.SubscriptionPlanSaleBlockSoldOut, response.Data[0].SaleAvailability.BlockReason)
}

func TestAdminCreateSubscriptionPlanRejectsInvalidSaleWindow(t *testing.T) {
	setupSubscriptionSaleControllerTestDB(t)
	confirmPaymentComplianceForTest(t)
	_ = operation_setting.GetPaymentSetting()

	router := gin.New()
	router.POST("/api/subscription/admin/plans", AdminCreateSubscriptionPlan)
	body := []byte(`{"plan":{"title":"Bad","price_amount":1,"currency":"USD","duration_unit":"month","duration_value":1,"enabled":true,"total_amount":100,"daily_sale_starts_at":"9:00","daily_sale_ends_at":"18:00"}}`)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/subscription/admin/plans", bytes.NewReader(body)))

	var response struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &response))
	require.False(t, response.Success)
	require.Contains(t, response.Message, "HH:mm")
}

func TestAdminCreateSubscriptionPlanStoresUSDRegardlessOfDisplayCurrency(t *testing.T) {
	setupSubscriptionSaleControllerTestDB(t)
	confirmPaymentComplianceForTest(t)
	_ = operation_setting.GetPaymentSetting()
	oldDisplayType := operation_setting.GetGeneralSetting().QuotaDisplayType
	operation_setting.GetGeneralSetting().QuotaDisplayType = operation_setting.QuotaDisplayTypeCNY
	t.Cleanup(func() {
		operation_setting.GetGeneralSetting().QuotaDisplayType = oldDisplayType
	})

	router := gin.New()
	router.POST("/api/subscription/admin/plans", AdminCreateSubscriptionPlan)
	body := []byte(`{"plan":{"title":"USD Plan","price_amount":14.62,"currency":"CNY","duration_unit":"month","duration_value":1,"enabled":true,"total_amount":100}}`)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/subscription/admin/plans", bytes.NewReader(body)))

	var response struct {
		Success bool `json:"success"`
		Data    struct {
			Currency string  `json:"currency"`
			Price    float64 `json:"price_amount"`
		} `json:"data"`
	}
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &response))
	require.True(t, response.Success)
	require.Equal(t, "USD", response.Data.Currency)
	require.InDelta(t, 14.62, response.Data.Price, 0.000001)

	var plan model.SubscriptionPlan
	require.NoError(t, model.DB.Where("title = ?", "USD Plan").First(&plan).Error)
	require.Equal(t, "USD", plan.Currency)
	require.InDelta(t, 14.62, plan.PriceAmount, 0.000001)
}

func TestAdminUpdateSubscriptionPlanStoresUSDRegardlessOfDisplayCurrency(t *testing.T) {
	setupSubscriptionSaleControllerTestDB(t)
	confirmPaymentComplianceForTest(t)
	_ = operation_setting.GetPaymentSetting()
	oldDisplayType := operation_setting.GetGeneralSetting().QuotaDisplayType
	operation_setting.GetGeneralSetting().QuotaDisplayType = operation_setting.QuotaDisplayTypeCNY
	t.Cleanup(func() {
		operation_setting.GetGeneralSetting().QuotaDisplayType = oldDisplayType
	})
	require.NoError(t, model.DB.Create(&model.SubscriptionPlan{
		Id:            66,
		Title:         "Existing Plan",
		PriceAmount:   9.99,
		Currency:      "CNY",
		DurationUnit:  model.SubscriptionDurationMonth,
		DurationValue: 1,
		Enabled:       true,
		TotalAmount:   100,
	}).Error)

	router := gin.New()
	router.PUT("/api/subscription/admin/plans/:id", AdminUpdateSubscriptionPlan)
	body := []byte(`{"plan":{"title":"Updated Plan","price_amount":14.62,"currency":"CNY","duration_unit":"month","duration_value":1,"enabled":true,"total_amount":100}}`)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPut, "/api/subscription/admin/plans/66", bytes.NewReader(body)))

	var response struct {
		Success bool `json:"success"`
	}
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &response))
	require.True(t, response.Success)

	var plan model.SubscriptionPlan
	require.NoError(t, model.DB.Where("id = ?", 66).First(&plan).Error)
	require.Equal(t, "USD", plan.Currency)
	require.InDelta(t, 14.62, plan.PriceAmount, 0.000001)
}

func TestAdminListPlanUserSubscriptionsReturnsPageInfo(t *testing.T) {
	setupSubscriptionSaleControllerTestDB(t)

	require.NoError(t, model.DB.Create(&model.User{
		Id:          20,
		Username:    "buyer",
		DisplayName: "Buyer Display",
		Email:       "buyer@example.com",
		Status:      common.UserStatusEnabled,
		Group:       "default",
		AffCode:     "buyer_aff",
		Password:    "secret",
	}).Error)
	require.NoError(t, model.DB.Create(&model.SubscriptionPlan{
		Id:            20,
		Title:         "Controller Plan",
		PriceAmount:   9.99,
		Currency:      "USD",
		DurationUnit:  model.SubscriptionDurationMonth,
		DurationValue: 1,
		Enabled:       true,
		TotalAmount:   1000,
	}).Error)
	require.NoError(t, model.DB.Create(&model.UserSubscription{
		Id:          20,
		UserId:      20,
		PlanId:      20,
		AmountTotal: 1000,
		AmountUsed:  200,
		Status:      "active",
		StartTime:   time.Now().Add(-time.Hour).Unix(),
		EndTime:     time.Now().Add(time.Hour).Unix(),
		Source:      "order",
	}).Error)

	router := gin.New()
	router.GET("/api/subscription/admin/plans/:id/subscriptions", AdminListPlanUserSubscriptions)
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/subscription/admin/plans/20/subscriptions?p=1&page_size=10&status=active&keyword=buyer", nil)
	router.ServeHTTP(recorder, req)

	var response struct {
		Success bool `json:"success"`
		Data    struct {
			Page     int `json:"page"`
			PageSize int `json:"page_size"`
			Total    int `json:"total"`
			Items    []struct {
				Subscription model.UserSubscription `json:"subscription"`
				User         model.User             `json:"user"`
			} `json:"items"`
		} `json:"data"`
	}
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &response))
	require.True(t, response.Success)
	require.Equal(t, 1, response.Data.Page)
	require.Equal(t, 10, response.Data.PageSize)
	require.Equal(t, 1, response.Data.Total)
	require.Len(t, response.Data.Items, 1)
	require.Equal(t, 20, response.Data.Items[0].Subscription.Id)
	require.Equal(t, 20, response.Data.Items[0].User.Id)
	require.Equal(t, "buyer", response.Data.Items[0].User.Username)
	require.Empty(t, response.Data.Items[0].User.Password)
	require.Nil(t, response.Data.Items[0].User.AccessToken)
}

func TestAdminListPlanUserSubscriptionsRejectsInvalidPlanID(t *testing.T) {
	setupSubscriptionSaleControllerTestDB(t)

	router := gin.New()
	router.GET("/api/subscription/admin/plans/:id/subscriptions", AdminListPlanUserSubscriptions)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/subscription/admin/plans/0/subscriptions", nil))

	var response struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &response))
	require.False(t, response.Success)
	require.Contains(t, response.Message, "套餐ID")
}

func setupAdminUpdateUserSubscriptionFixture(t *testing.T) int {
	t.Helper()
	require.NoError(t, model.DB.Create(&model.User{
		Id:          31,
		Username:    "subscription-editor",
		DisplayName: "Subscription Editor",
		Email:       "subscription-editor@example.com",
		Status:      common.UserStatusEnabled,
		Group:       "default",
		AffCode:     "subscription_editor_aff",
		Password:    "secret",
	}).Error)
	require.NoError(t, model.DB.Create(&model.SubscriptionPlan{
		Id:            31,
		Title:         "Old Plan",
		PriceAmount:   9.99,
		Currency:      "USD",
		DurationUnit:  model.SubscriptionDurationMonth,
		DurationValue: 1,
		Enabled:       true,
		TotalAmount:   1000,
	}).Error)
	require.NoError(t, model.DB.Create(&model.SubscriptionPlan{
		Id:            32,
		Title:         "New Plan",
		PriceAmount:   19.99,
		Currency:      "USD",
		DurationUnit:  model.SubscriptionDurationMonth,
		DurationValue: 1,
		Enabled:       true,
		TotalAmount:   2000,
	}).Error)
	require.NoError(t, model.DB.Create(&model.UserSubscription{
		Id:            31,
		UserId:        31,
		PlanId:        31,
		AmountTotal:   1000,
		AmountUsed:    500,
		Status:        "active",
		StartTime:     time.Now().Add(-time.Hour).Unix(),
		EndTime:       time.Now().Add(time.Hour).Unix(),
		Source:        "order",
		LastResetTime: time.Now().Add(-time.Hour).Unix(),
		NextResetTime: time.Now().Add(time.Hour).Unix(),
	}).Error)
	return 31
}

func TestAdminUpdateUserSubscriptionUpdatesEditableFields(t *testing.T) {
	setupSubscriptionSaleControllerTestDB(t)
	subId := setupAdminUpdateUserSubscriptionFixture(t)
	startTime := time.Now().Add(-2 * time.Hour).Unix()
	endTime := time.Now().Add(2 * time.Hour).Unix()

	router := gin.New()
	router.PUT("/api/subscription/admin/user_subscriptions/:id", AdminUpdateUserSubscription)
	body := []byte(`{"plan_id":32,"status":"active","start_time":` + strconv.FormatInt(startTime, 10) + `,"end_time":` + strconv.FormatInt(endTime, 10) + `,"amount_total":0,"amount_used":0,"next_reset_time":0}`)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPut, "/api/subscription/admin/user_subscriptions/31", bytes.NewReader(body)))

	var response struct {
		Success bool `json:"success"`
		Data    struct {
			Subscription model.UserSubscription `json:"subscription"`
		} `json:"data"`
	}
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &response))
	require.True(t, response.Success)
	require.Equal(t, subId, response.Data.Subscription.Id)
	require.Equal(t, 32, response.Data.Subscription.PlanId)
	require.EqualValues(t, 0, response.Data.Subscription.AmountTotal)
	require.EqualValues(t, 0, response.Data.Subscription.AmountUsed)
	require.Equal(t, startTime, response.Data.Subscription.StartTime)
	require.Equal(t, endTime, response.Data.Subscription.EndTime)

	var sub model.UserSubscription
	require.NoError(t, model.DB.Where("id = ?", subId).First(&sub).Error)
	require.Equal(t, 32, sub.PlanId)
	require.EqualValues(t, 0, sub.AmountTotal)
	require.EqualValues(t, 0, sub.AmountUsed)
	require.Equal(t, startTime, sub.StartTime)
	require.Equal(t, endTime, sub.EndTime)
	require.EqualValues(t, 0, sub.NextResetTime)
}

func TestAdminUpdateUserSubscriptionRejectsInvalidInput(t *testing.T) {
	setupSubscriptionSaleControllerTestDB(t)
	setupAdminUpdateUserSubscriptionFixture(t)

	tests := []struct {
		name string
		body string
		want string
	}{
		{
			name: "invalid status",
			body: `{"plan_id":32,"status":"paused","start_time":100,"end_time":200,"amount_total":1000,"amount_used":0,"next_reset_time":0}`,
			want: "状态",
		},
		{
			name: "end before start",
			body: `{"plan_id":32,"status":"active","start_time":200,"end_time":100,"amount_total":1000,"amount_used":0,"next_reset_time":0}`,
			want: "结束",
		},
		{
			name: "used exceeds total",
			body: `{"plan_id":32,"status":"active","start_time":100,"end_time":200,"amount_total":1000,"amount_used":1001,"next_reset_time":0}`,
			want: "已用",
		},
	}

	router := gin.New()
	router.PUT("/api/subscription/admin/user_subscriptions/:id", AdminUpdateUserSubscription)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPut, "/api/subscription/admin/user_subscriptions/31", strings.NewReader(tt.body)))

			var response struct {
				Success bool   `json:"success"`
				Message string `json:"message"`
			}
			require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &response))
			require.False(t, response.Success)
			require.Contains(t, response.Message, tt.want)
		})
	}
}
