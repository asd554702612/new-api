package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/stretchr/testify/require"
)

func insertBalancePurchasePlanForTest(t *testing.T, id int, price float64, currency string) *SubscriptionPlan {
	t.Helper()
	plan := &SubscriptionPlan{
		Id:            id,
		Title:         "Balance Purchase Plan",
		PriceAmount:   price,
		Currency:      currency,
		DurationUnit:  SubscriptionDurationMonth,
		DurationValue: 1,
		Enabled:       true,
		TotalAmount:   1234,
	}
	require.NoError(t, DB.Create(plan).Error)
	return plan
}

func TestPurchaseSubscriptionWithBalanceChargesUSDPriceAmount(t *testing.T) {
	truncateTables(t)
	oldQuotaPerUnit := common.QuotaPerUnit
	common.QuotaPerUnit = 100
	t.Cleanup(func() { common.QuotaPerUnit = oldQuotaPerUnit })

	insertUserForPaymentGuardTest(t, 701, 2000)
	plan := insertBalancePurchasePlanForTest(t, 701, 10, "USD")

	require.NoError(t, PurchaseSubscriptionWithBalance(701, plan.Id))

	var user User
	require.NoError(t, DB.Where("id = ?", 701).First(&user).Error)
	require.Equal(t, 1000, user.Quota)

	var order SubscriptionOrder
	require.NoError(t, DB.Where("user_id = ? AND plan_id = ?", 701, plan.Id).First(&order).Error)
	require.Equal(t, PaymentProviderBalance, order.PaymentProvider)
	require.Equal(t, common.TopUpStatusSuccess, order.Status)
	require.Equal(t, "charged_quota=1000", order.ProviderPayload)

	var sub UserSubscription
	require.NoError(t, DB.Where("user_id = ? AND plan_id = ?", 701, plan.Id).First(&sub).Error)
	require.Equal(t, plan.TotalAmount, sub.AmountTotal)
}

func TestPurchaseSubscriptionWithBalanceConvertsLegacyCNYPlanPriceToUSD(t *testing.T) {
	truncateTables(t)
	oldQuotaPerUnit := common.QuotaPerUnit
	oldExchangeRate := operation_setting.USDExchangeRate
	common.QuotaPerUnit = 100
	operation_setting.USDExchangeRate = 7.3
	t.Cleanup(func() {
		common.QuotaPerUnit = oldQuotaPerUnit
		operation_setting.USDExchangeRate = oldExchangeRate
	})

	insertUserForPaymentGuardTest(t, 702, 2000)
	plan := insertBalancePurchasePlanForTest(t, 702, 73, "CNY")

	require.NoError(t, PurchaseSubscriptionWithBalance(702, plan.Id))

	var user User
	require.NoError(t, DB.Where("id = ?", 702).First(&user).Error)
	require.Equal(t, 1000, user.Quota)

	var order SubscriptionOrder
	require.NoError(t, DB.Where("user_id = ? AND plan_id = ?", 702, plan.Id).First(&order).Error)
	require.Equal(t, "charged_quota=1000", order.ProviderPayload)
}
