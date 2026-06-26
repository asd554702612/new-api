package controller

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestGetTopUpInfoHidesLegacyEpayMethodsWhenEpayDisabled(t *testing.T) {
	confirmPaymentComplianceForTest(t)
	resetTopUpInfoPaymentSettings(t)

	operation_setting.PayMethods = []map[string]string{
		{"name": "支付宝", "type": "alipay"},
		{"name": "微信", "type": "wxpay"},
		{"name": "自定义1", "type": "custom1"},
	}
	operation_setting.PayAddress = ""
	operation_setting.EpayId = ""
	operation_setting.EpayKey = ""

	setting.WechatPayEnabled = true
	setting.WechatPayAppID = "wx_app"
	setting.WechatPayMchID = "mch"
	setting.WechatPayAPIv3Key = "12345678901234567890123456789012"
	setting.WechatPayPrivateKey = "private"
	setting.WechatPayMerchantSerialNo = "serial"

	data := performGetTopUpInfoForTest(t)
	require.False(t, data.EnableOnlineTopUp)
	require.True(t, data.EnableWechatPayTopUp)
	require.Contains(t, topUpInfoPaymentTypes(data), "wechat_pay")
	require.Equal(t, "微信支付", topUpInfoPaymentName(data, "wechat_pay"))
	require.NotContains(t, topUpInfoPaymentTypes(data), "wxpay")
	require.NotContains(t, topUpInfoPaymentTypes(data), "alipay")
	require.NotContains(t, topUpInfoPaymentTypes(data), "custom1")
}

func TestGetTopUpInfoKeepsLegacyEpayMethodsWhenEpayEnabled(t *testing.T) {
	confirmPaymentComplianceForTest(t)
	resetTopUpInfoPaymentSettings(t)

	operation_setting.PayMethods = []map[string]string{
		{"name": "支付宝", "type": "alipay"},
		{"name": "微信", "type": "wxpay"},
	}
	operation_setting.PayAddress = "https://pay.example.com"
	operation_setting.EpayId = "epay_id"
	operation_setting.EpayKey = "epay_key"

	data := performGetTopUpInfoForTest(t)
	require.True(t, data.EnableOnlineTopUp)
	require.Contains(t, topUpInfoPaymentTypes(data), "alipay")
	require.Contains(t, topUpInfoPaymentTypes(data), "wxpay")
}

func resetTopUpInfoPaymentSettings(t *testing.T) {
	t.Helper()

	originalPayMethods := operation_setting.PayMethods
	originalPayAddress := operation_setting.PayAddress
	originalEpayID := operation_setting.EpayId
	originalEpayKey := operation_setting.EpayKey
	originalStripeSecret := setting.StripeApiSecret
	originalStripeWebhook := setting.StripeWebhookSecret
	originalStripePrice := setting.StripePriceId
	originalWechatEnabled := setting.WechatPayEnabled
	originalWechatAppID := setting.WechatPayAppID
	originalWechatMchID := setting.WechatPayMchID
	originalWechatAPIv3Key := setting.WechatPayAPIv3Key
	originalWechatPrivateKey := setting.WechatPayPrivateKey
	originalWechatSerialNo := setting.WechatPayMerchantSerialNo
	originalAlipayEnabled := setting.AlipayEnabled
	originalWaffoEnabled := setting.WaffoEnabled
	originalWaffoPancakeMerchantID := setting.WaffoPancakeMerchantID
	originalCreemAPIKey := setting.CreemApiKey
	originalCreemProducts := setting.CreemProducts
	originalCasdoorEnabled := setting.CasdoorPaymentEnabled
	originalCasdoorBaseURL := setting.CasdoorBaseURL
	originalCasdoorClientID := setting.CasdoorClientID
	originalCasdoorClientSecret := setting.CasdoorClientSecret
	originalCasdoorApp := setting.CasdoorApplicationName
	originalCasdoorProduct := setting.CasdoorPaymentProduct
	originalCasdoorProvider := setting.CasdoorPaymentProvider
	originalCasdoorCurrency := setting.CasdoorPaymentCurrency
	originalCasdoorMinTopUp := setting.CasdoorPaymentMinTopUp

	t.Cleanup(func() {
		operation_setting.PayMethods = originalPayMethods
		operation_setting.PayAddress = originalPayAddress
		operation_setting.EpayId = originalEpayID
		operation_setting.EpayKey = originalEpayKey
		setting.StripeApiSecret = originalStripeSecret
		setting.StripeWebhookSecret = originalStripeWebhook
		setting.StripePriceId = originalStripePrice
		setting.WechatPayEnabled = originalWechatEnabled
		setting.WechatPayAppID = originalWechatAppID
		setting.WechatPayMchID = originalWechatMchID
		setting.WechatPayAPIv3Key = originalWechatAPIv3Key
		setting.WechatPayPrivateKey = originalWechatPrivateKey
		setting.WechatPayMerchantSerialNo = originalWechatSerialNo
		setting.AlipayEnabled = originalAlipayEnabled
		setting.WaffoEnabled = originalWaffoEnabled
		setting.WaffoPancakeMerchantID = originalWaffoPancakeMerchantID
		setting.CreemApiKey = originalCreemAPIKey
		setting.CreemProducts = originalCreemProducts
		setting.CasdoorPaymentEnabled = originalCasdoorEnabled
		setting.CasdoorBaseURL = originalCasdoorBaseURL
		setting.CasdoorClientID = originalCasdoorClientID
		setting.CasdoorClientSecret = originalCasdoorClientSecret
		setting.CasdoorApplicationName = originalCasdoorApp
		setting.CasdoorPaymentProduct = originalCasdoorProduct
		setting.CasdoorPaymentProvider = originalCasdoorProvider
		setting.CasdoorPaymentCurrency = originalCasdoorCurrency
		setting.CasdoorPaymentMinTopUp = originalCasdoorMinTopUp
	})

	operation_setting.PayMethods = nil
	operation_setting.PayAddress = ""
	operation_setting.EpayId = ""
	operation_setting.EpayKey = ""
	setting.StripeApiSecret = ""
	setting.StripeWebhookSecret = ""
	setting.StripePriceId = ""
	setting.WechatPayEnabled = false
	setting.WechatPayAppID = ""
	setting.WechatPayMchID = ""
	setting.WechatPayAPIv3Key = ""
	setting.WechatPayPrivateKey = ""
	setting.WechatPayMerchantSerialNo = ""
	setting.AlipayEnabled = false
	setting.WaffoEnabled = false
	setting.WaffoPancakeMerchantID = ""
	setting.CreemApiKey = ""
	setting.CreemProducts = "[]"
	setting.CasdoorPaymentEnabled = false
	setting.CasdoorBaseURL = ""
	setting.CasdoorClientID = ""
	setting.CasdoorClientSecret = ""
	setting.CasdoorApplicationName = ""
	setting.CasdoorPaymentProduct = ""
	setting.CasdoorPaymentProvider = ""
	setting.CasdoorPaymentCurrency = ""
	setting.CasdoorPaymentMinTopUp = 1
}

func TestGetTopUpInfoIncludesCasdoorWhenEnabled(t *testing.T) {
	confirmPaymentComplianceForTest(t)
	resetTopUpInfoPaymentSettings(t)

	setting.CasdoorPaymentEnabled = true
	setting.CasdoorBaseURL = "https://login.gepinkeji.com"
	setting.CasdoorClientID = "client-id"
	setting.CasdoorClientSecret = "client-secret"
	setting.CasdoorApplicationName = "app-token-gptk"
	setting.CasdoorPaymentProduct = "external-pay-template"
	setting.CasdoorPaymentProvider = "provider_payment_wechat_gepinkeji"
	setting.CasdoorPaymentCurrency = "CNY"
	setting.CasdoorPaymentMinTopUp = 3

	data := performGetTopUpInfoForTest(t)
	require.True(t, data.EnableCasdoorTopUp)
	require.Equal(t, 3, data.CasdoorMinTopUp)
	require.Contains(t, topUpInfoPaymentTypes(data), "casdoor")
	require.Equal(t, "Casdoor 统一支付", topUpInfoPaymentName(data, "casdoor"))
}

func TestGetTopUpInfoConvertsCasdoorMinTopUpInTokensDisplayMode(t *testing.T) {
	confirmPaymentComplianceForTest(t)
	resetTopUpInfoPaymentSettings(t)
	originalQuotaDisplayType := operation_setting.GetGeneralSetting().QuotaDisplayType
	t.Cleanup(func() {
		operation_setting.GetGeneralSetting().QuotaDisplayType = originalQuotaDisplayType
	})
	operation_setting.GetGeneralSetting().QuotaDisplayType = operation_setting.QuotaDisplayTypeTokens

	setting.CasdoorPaymentEnabled = true
	setting.CasdoorBaseURL = "https://login.gepinkeji.com"
	setting.CasdoorClientID = "client-id"
	setting.CasdoorClientSecret = "client-secret"
	setting.CasdoorApplicationName = "app-token-gptk"
	setting.CasdoorPaymentProduct = "external-pay-template"
	setting.CasdoorPaymentProvider = "provider_payment_wechat_gepinkeji"
	setting.CasdoorPaymentCurrency = "CNY"
	setting.CasdoorPaymentMinTopUp = 3

	data := performGetTopUpInfoForTest(t)
	expectedMinTopUp := int(3 * common.QuotaPerUnit)
	require.True(t, data.EnableCasdoorTopUp)
	require.Equal(t, expectedMinTopUp, data.CasdoorMinTopUp)
	require.Equal(t, strconv.Itoa(expectedMinTopUp), topUpInfoPaymentMinTopUp(data, "casdoor"))
}

type topUpInfoResponseForTest struct {
	Success bool                         `json:"success"`
	Data    topUpInfoResponseForTestData `json:"data"`
}

func performGetTopUpInfoForTest(t *testing.T) topUpInfoResponseForTestData {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/topup/info", GetTopUpInfo)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/topup/info", nil)
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	var response topUpInfoResponseForTest
	require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &response))
	require.True(t, response.Success)
	return response.Data
}

type topUpInfoResponseForTestData struct {
	EnableOnlineTopUp    bool                `json:"enable_online_topup"`
	EnableWechatPayTopUp bool                `json:"enable_wechat_pay_topup"`
	EnableCasdoorTopUp   bool                `json:"enable_casdoor_topup"`
	CasdoorMinTopUp      int                 `json:"casdoor_min_topup"`
	PayMethods           []map[string]string `json:"pay_methods"`
}

func topUpInfoPaymentTypes(data topUpInfoResponseForTestData) []string {
	types := make([]string, 0, len(data.PayMethods))
	for _, method := range data.PayMethods {
		types = append(types, method["type"])
	}
	return types
}

func topUpInfoPaymentName(data topUpInfoResponseForTestData, paymentType string) string {
	for _, method := range data.PayMethods {
		if method["type"] == paymentType {
			return method["name"]
		}
	}
	return ""
}

func topUpInfoPaymentMinTopUp(data topUpInfoResponseForTestData, paymentType string) string {
	for _, method := range data.PayMethods {
		if method["type"] == paymentType {
			return method["min_topup"]
		}
	}
	return ""
}
