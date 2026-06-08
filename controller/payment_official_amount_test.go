package controller

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/stretchr/testify/require"
)

func TestGetOfficialPayMoneyUsesChannelUnitPrice(t *testing.T) {
	originalPrice := operation_setting.Price
	originalWechatPayUnitPrice := setting.WechatPayUnitPrice
	originalAlipayUnitPrice := setting.AlipayUnitPrice
	originalQuotaDisplayType := operation_setting.GetGeneralSetting().QuotaDisplayType
	originalDiscounts := make(map[int]float64, len(operation_setting.GetPaymentSetting().AmountDiscount))
	for k, v := range operation_setting.GetPaymentSetting().AmountDiscount {
		originalDiscounts[k] = v
	}
	originalTopupGroupRatio := common.TopupGroupRatio2JSONString()

	t.Cleanup(func() {
		operation_setting.Price = originalPrice
		setting.WechatPayUnitPrice = originalWechatPayUnitPrice
		setting.AlipayUnitPrice = originalAlipayUnitPrice
		operation_setting.GetGeneralSetting().QuotaDisplayType = originalQuotaDisplayType
		operation_setting.GetPaymentSetting().AmountDiscount = originalDiscounts
		require.NoError(t, common.UpdateTopupGroupRatioByJSONString(originalTopupGroupRatio))
	})

	operation_setting.Price = 7.3
	setting.WechatPayUnitPrice = 6.8
	setting.AlipayUnitPrice = 7.1
	operation_setting.GetGeneralSetting().QuotaDisplayType = operation_setting.QuotaDisplayTypeUSD
	operation_setting.GetPaymentSetting().AmountDiscount = map[int]float64{
		10: 0.8,
	}
	require.NoError(t, common.UpdateTopupGroupRatioByJSONString(`{"default":1,"vip":1.2}`))

	require.InDelta(t, 65.28, getOfficialPayMoney(10, "vip", officialTopUpUnitPrice(setting.WechatPayUnitPrice)), 0.000001)
	require.InDelta(t, 68.16, getOfficialPayMoney(10, "vip", officialTopUpUnitPrice(setting.AlipayUnitPrice)), 0.000001)
}

func TestOfficialTopUpUnitPriceFallsBackToEpayPrice(t *testing.T) {
	originalPrice := operation_setting.Price
	originalQuotaDisplayType := operation_setting.GetGeneralSetting().QuotaDisplayType
	originalDiscounts := make(map[int]float64, len(operation_setting.GetPaymentSetting().AmountDiscount))
	for k, v := range operation_setting.GetPaymentSetting().AmountDiscount {
		originalDiscounts[k] = v
	}
	originalTopupGroupRatio := common.TopupGroupRatio2JSONString()

	t.Cleanup(func() {
		operation_setting.Price = originalPrice
		operation_setting.GetGeneralSetting().QuotaDisplayType = originalQuotaDisplayType
		operation_setting.GetPaymentSetting().AmountDiscount = originalDiscounts
		require.NoError(t, common.UpdateTopupGroupRatioByJSONString(originalTopupGroupRatio))
	})

	operation_setting.Price = 7.3
	operation_setting.GetGeneralSetting().QuotaDisplayType = operation_setting.QuotaDisplayTypeTokens
	operation_setting.GetPaymentSetting().AmountDiscount = map[int]float64{
		int(common.QuotaPerUnit * 3): 0.5,
	}
	require.NoError(t, common.UpdateTopupGroupRatioByJSONString(`{"default":1,"vip":1.2}`))

	unitPrice := officialTopUpUnitPrice(0)
	require.InDelta(t, 13.14, getOfficialPayMoney(int64(common.QuotaPerUnit*3), "vip", unitPrice), 0.000001)

	unitPrice = officialTopUpUnitPrice(-1)
	require.InDelta(t, 13.14, getOfficialPayMoney(int64(common.QuotaPerUnit*3), "vip", unitPrice), 0.000001)
}
