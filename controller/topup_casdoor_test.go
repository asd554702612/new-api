package controller

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupCasdoorControllerTestDB(t *testing.T) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	common.SetDatabaseTypes(common.DatabaseTypeSQLite, common.DatabaseTypeSQLite)
	common.RedisEnabled = false
	common.AffiliateEnabled = false

	dsn := "file:" + strings.ReplaceAll(t.Name(), "/", "_") + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	model.DB = db
	model.LOG_DB = db
	require.NoError(t, db.AutoMigrate(
		&model.User{},
		&model.Log{},
		&model.TopUp{},
		&model.SubscriptionPlan{},
		&model.SubscriptionOrder{},
		&model.UserSubscription{},
		&model.AffiliateLedger{},
		&model.PaymentActivityConfig{},
		&model.LuckyWheelSession{},
		&model.LuckyWheelDrawRecord{},
		&model.LuckyWheelGoldenWindowClaim{},
		&model.LuckyWheelInviteBonusEvent{},
		&model.RechargeActivityChance{},
		&model.RechargeActivityDrawRecord{},
	))
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})
}

func configureCasdoorPaymentForTest(t *testing.T) {
	t.Helper()
	confirmPaymentComplianceForTest(t)
	oldEnabled := setting.CasdoorPaymentEnabled
	oldBaseURL := setting.CasdoorBaseURL
	oldClientID := setting.CasdoorClientID
	oldSecret := setting.CasdoorClientSecret
	oldApp := setting.CasdoorApplicationName
	oldProduct := setting.CasdoorPaymentProduct
	oldProvider := setting.CasdoorPaymentProvider
	oldCurrency := setting.CasdoorPaymentCurrency
	oldUnitPrice := setting.CasdoorPaymentUnitPrice
	oldMinTopUp := setting.CasdoorPaymentMinTopUp
	t.Cleanup(func() {
		setting.CasdoorPaymentEnabled = oldEnabled
		setting.CasdoorBaseURL = oldBaseURL
		setting.CasdoorClientID = oldClientID
		setting.CasdoorClientSecret = oldSecret
		setting.CasdoorApplicationName = oldApp
		setting.CasdoorPaymentProduct = oldProduct
		setting.CasdoorPaymentProvider = oldProvider
		setting.CasdoorPaymentCurrency = oldCurrency
		setting.CasdoorPaymentUnitPrice = oldUnitPrice
		setting.CasdoorPaymentMinTopUp = oldMinTopUp
	})
	setting.CasdoorPaymentEnabled = true
	setting.CasdoorBaseURL = "https://login.gepinkeji.com"
	setting.CasdoorClientID = "client-id"
	setting.CasdoorClientSecret = "client-secret"
	setting.CasdoorApplicationName = "app-token-gptk"
	setting.CasdoorPaymentProduct = "external-pay-template"
	setting.CasdoorPaymentProvider = "provider_payment_wechat_gepinkeji"
	setting.CasdoorPaymentCurrency = "CNY"
	setting.CasdoorPaymentUnitPrice = 7.3
	setting.CasdoorPaymentMinTopUp = 1
}

func performCasdoorWebhookForTest(t *testing.T, body string) *httptest.ResponseRecorder {
	t.Helper()
	router := gin.New()
	router.POST("/api/casdoor/payment/webhook", CasdoorPaymentWebhook)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/casdoor/payment/webhook", strings.NewReader(body))
	request.Header.Set("X-Casdoor-Webhook-Event", "payment.paid")
	request.Header.Set("X-Casdoor-Webhook-Signature", service.BuildCasdoorWebhookSignature("client-secret", []byte(body)))
	router.ServeHTTP(recorder, request)
	return recorder
}

func performCasdoorAmountForTest(t *testing.T, userID int, body string) *httptest.ResponseRecorder {
	t.Helper()
	router := gin.New()
	router.POST("/api/user/casdoor/amount", func(c *gin.Context) {
		c.Set("id", userID)
		RequestCasdoorAmount(c)
	})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/user/casdoor/amount", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)
	return recorder
}

func performCasdoorPayForTest(t *testing.T, userID int, body string) *httptest.ResponseRecorder {
	t.Helper()
	router := gin.New()
	router.POST("/api/user/casdoor/pay", func(c *gin.Context) {
		c.Set("id", userID)
		RequestCasdoorPay(c)
	})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/user/casdoor/pay", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)
	return recorder
}

func TestTopUpMinForCurrentDisplayUsesConfiguredMinTopUp(t *testing.T) {
	originalQuotaDisplayType := operation_setting.GetGeneralSetting().QuotaDisplayType
	t.Cleanup(func() {
		operation_setting.GetGeneralSetting().QuotaDisplayType = originalQuotaDisplayType
	})

	operation_setting.GetGeneralSetting().QuotaDisplayType = operation_setting.QuotaDisplayTypeUSD
	require.Equal(t, int64(3), topUpMinForCurrentDisplay(3))

	operation_setting.GetGeneralSetting().QuotaDisplayType = operation_setting.QuotaDisplayTypeTokens
	require.Equal(t, int64(3*common.QuotaPerUnit), topUpMinForCurrentDisplay(3))
}

func TestCasdoorAmountAndPayRejectBelowCasdoorMinTopUp(t *testing.T) {
	setupCasdoorControllerTestDB(t)
	configureCasdoorPaymentForTest(t)
	originalMinTopUp := operation_setting.MinTopUp
	originalQuotaDisplayType := operation_setting.GetGeneralSetting().QuotaDisplayType
	t.Cleanup(func() {
		operation_setting.MinTopUp = originalMinTopUp
		operation_setting.GetGeneralSetting().QuotaDisplayType = originalQuotaDisplayType
	})
	operation_setting.MinTopUp = 3
	operation_setting.GetGeneralSetting().QuotaDisplayType = operation_setting.QuotaDisplayTypeUSD
	setting.CasdoorPaymentMinTopUp = 10
	require.NoError(t, model.DB.Create(&model.User{Id: 705, Username: "casdoor-reject", Status: common.UserStatusEnabled, Group: "default"}).Error)

	amountRecorder := performCasdoorAmountForTest(t, 705, `{"amount":3}`)
	payRecorder := performCasdoorPayForTest(t, 705, `{"amount":3}`)

	for _, recorder := range []*httptest.ResponseRecorder{amountRecorder, payRecorder} {
		require.Equal(t, http.StatusOK, recorder.Code)
		var response struct {
			Message string `json:"message"`
			Data    string `json:"data"`
		}
		require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &response))
		require.Equal(t, "error", response.Message)
		require.Equal(t, "充值数量不能小于 10", response.Data)
	}
}

func TestCasdoorAmountConvertsMinTopUpInTokensDisplayMode(t *testing.T) {
	setupCasdoorControllerTestDB(t)
	configureCasdoorPaymentForTest(t)
	originalMinTopUp := operation_setting.MinTopUp
	originalQuotaDisplayType := operation_setting.GetGeneralSetting().QuotaDisplayType
	t.Cleanup(func() {
		operation_setting.MinTopUp = originalMinTopUp
		operation_setting.GetGeneralSetting().QuotaDisplayType = originalQuotaDisplayType
	})
	operation_setting.MinTopUp = 1
	operation_setting.GetGeneralSetting().QuotaDisplayType = operation_setting.QuotaDisplayTypeTokens
	setting.CasdoorPaymentMinTopUp = 3
	require.NoError(t, model.DB.Create(&model.User{Id: 706, Username: "casdoor-tokens", Status: common.UserStatusEnabled, Group: "default"}).Error)
	expectedMinTopUp := int64(3 * common.QuotaPerUnit)

	tooLow := performCasdoorAmountForTest(t, 706, `{"amount":`+strconv.FormatInt(expectedMinTopUp-1, 10)+`}`)

	var lowResponse struct {
		Message string `json:"message"`
		Data    string `json:"data"`
	}
	require.NoError(t, common.Unmarshal(tooLow.Body.Bytes(), &lowResponse))
	require.Equal(t, "error", lowResponse.Message)
	require.Equal(t, "充值数量不能小于 "+strconv.FormatInt(expectedMinTopUp, 10), lowResponse.Data)
}

func TestCasdoorWebhookCompletesTopUpAndIsIdempotent(t *testing.T) {
	setupCasdoorControllerTestDB(t)
	configureCasdoorPaymentForTest(t)
	require.NoError(t, model.DB.Create(&model.User{Id: 701, Username: "casdoor-topup", Status: common.UserStatusEnabled, Quota: 0}).Error)
	require.NoError(t, model.DB.Create(&model.TopUp{
		UserId:          701,
		Amount:          3,
		Money:           88.66,
		TradeNo:         "casdoor-topup-1",
		PaymentMethod:   model.PaymentMethodCasdoor,
		PaymentProvider: model.PaymentProviderCasdoor,
		CreateTime:      time.Now().Unix(),
		Status:          common.TopUpStatusPending,
	}).Error)
	body := `{"event":"payment.paid","application":"app-token-gptk","externalOrderId":"casdoor-topup-1","orderId":"gepin/order_1","paymentId":"gepin/payment_1","userId":"gepin/alice","products":[{"owner":"gepin","name":"external-pay-template","displayName":"GPTK 充值","price":88.66,"currency":"CNY","quantity":1}],"amount":88.66,"currency":"CNY","providerName":"provider_payment_wechat_gepinkeji","paidTime":"2026-06-25T10:00:00+08:00"}`

	first := performCasdoorWebhookForTest(t, body)
	second := performCasdoorWebhookForTest(t, body)

	require.Equal(t, http.StatusOK, first.Code)
	require.Equal(t, http.StatusOK, second.Code)
	var topUp model.TopUp
	require.NoError(t, model.DB.Where("trade_no = ?", "casdoor-topup-1").First(&topUp).Error)
	require.Equal(t, common.TopUpStatusSuccess, topUp.Status)
	require.Contains(t, topUp.ProviderPayload, "gepin/payment_1")
	var user model.User
	require.NoError(t, model.DB.Select("quota").Where("id = ?", 701).First(&user).Error)
	require.Equal(t, int(3*common.QuotaPerUnit), user.Quota)
}

func TestCasdoorWebhookRejectsApplicationMismatch(t *testing.T) {
	setupCasdoorControllerTestDB(t)
	configureCasdoorPaymentForTest(t)
	require.NoError(t, model.DB.Create(&model.User{Id: 702, Username: "casdoor-mismatch", Status: common.UserStatusEnabled, Quota: 0}).Error)
	require.NoError(t, model.DB.Create(&model.TopUp{
		UserId:          702,
		Amount:          2,
		Money:           12.34,
		TradeNo:         "casdoor-app-mismatch",
		PaymentMethod:   model.PaymentMethodCasdoor,
		PaymentProvider: model.PaymentProviderCasdoor,
		CreateTime:      time.Now().Unix(),
		Status:          common.TopUpStatusPending,
	}).Error)
	body := `{"event":"payment.paid","application":"app-token-other","externalOrderId":"casdoor-app-mismatch","orderId":"gepin/order_2","paymentId":"gepin/payment_2","amount":12.34,"currency":"CNY","providerName":"provider_payment_wechat_gepinkeji","paidTime":"2026-06-25T10:00:00+08:00"}`

	recorder := performCasdoorWebhookForTest(t, body)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	topUp := model.GetTopUpByTradeNo("casdoor-app-mismatch")
	require.NotNil(t, topUp)
	require.Equal(t, common.TopUpStatusPending, topUp.Status)
}

func TestCasdoorWebhookCompletesSubscriptionOrder(t *testing.T) {
	setupCasdoorControllerTestDB(t)
	configureCasdoorPaymentForTest(t)
	require.NoError(t, model.DB.Create(&model.User{Id: 703, Username: "casdoor-sub", Status: common.UserStatusEnabled, Quota: 0}).Error)
	plan := &model.SubscriptionPlan{
		Title:         "Casdoor Plan",
		PriceAmount:   19.99,
		Currency:      "CNY",
		DurationUnit:  model.SubscriptionDurationMonth,
		DurationValue: 1,
		Enabled:       true,
		TotalAmount:   1000,
	}
	require.NoError(t, model.DB.Create(plan).Error)
	require.NoError(t, model.DB.Create(&model.SubscriptionOrder{
		UserId:          703,
		PlanId:          plan.Id,
		Money:           19.99,
		TradeNo:         "casdoor-sub-1",
		PaymentMethod:   model.PaymentMethodCasdoor,
		PaymentProvider: model.PaymentProviderCasdoor,
		Status:          common.TopUpStatusPending,
		CreateTime:      time.Now().Unix(),
	}).Error)
	body := `{"event":"payment.paid","application":"app-token-gptk","externalOrderId":"casdoor-sub-1","orderId":"gepin/order_3","paymentId":"gepin/payment_3","amount":19.99,"currency":"CNY","providerName":"provider_payment_wechat_gepinkeji","paidTime":"2026-06-25T10:00:00+08:00"}`

	recorder := performCasdoorWebhookForTest(t, body)

	require.Equal(t, http.StatusOK, recorder.Code)
	order := model.GetSubscriptionOrderByTradeNo("casdoor-sub-1")
	require.NotNil(t, order)
	require.Equal(t, common.TopUpStatusSuccess, order.Status)
	require.Contains(t, order.ProviderPayload, "gepin/payment_3")
	var count int64
	require.NoError(t, model.DB.Model(&model.UserSubscription{}).Where("user_id = ?", 703).Count(&count).Error)
	require.Equal(t, int64(1), count)
}
