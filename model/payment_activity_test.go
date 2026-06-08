package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func defaultTestLuckyWheelConfig() LuckyWheelConfig {
	return LuckyWheelConfig{
		EligibleOrderTypes:  []string{PaymentOrderTypeBalance, PaymentOrderTypeSubscription},
		MultiplierStep:      0.1,
		GlobalMaxMultiplier: 2,
		IntroText:           "test wheel",
		RulesTitle:          "rules",
		RulesItems:          []string{"pay and draw"},
		AmountTiers: []LuckyWheelAmountTier{
			{
				Id:            "tier_1",
				Name:          "Tier 1",
				MinAmount:     1,
				MinMultiplier: 1,
				MaxMultiplier: 1,
				DrawCount:     1,
			},
		},
		InviteBonus: LuckyWheelInviteBonusConfig{
			Enabled:          false,
			QualifyingAmount: 20,
			BonusPerInvitee:  0.2,
			MaxBonus:         1,
			ConsumePolicy:    LuckyWheelInviteBonusConsumeNextSessionOnce,
		},
		GoldenWindow: LuckyWheelGoldenWindowConfig{
			Enabled:    false,
			Timezone:   "Asia/Shanghai",
			StartTime:  "20:00",
			EndTime:    "22:00",
			MinAmount:  50,
			ExtraDraws: 1,
			DailyQuota: 5,
		},
	}
}

func defaultTestRechargeActivityConfig() RechargeActivityConfig {
	return RechargeActivityConfig{
		EligibleOrderTypes: []string{PaymentOrderTypeBalance, PaymentOrderTypeSubscription},
		IntroText:          "test recharge activity",
		RulesTitle:         "rules",
		RulesItems:         []string{"pay and draw"},
		Prizes: []RechargeActivityPrize{
			{
				Id:                "small",
				Name:              "Small prize",
				RewardAmount:      0,
				RewardDescription: "Manual small prize",
				Probability:       100,
				MinPayAmount:      1,
				Enabled:           true,
				SortOrder:         1,
			},
		},
	}
}

func insertActivityTestTopUp(t *testing.T, tradeNo string, userID int, amount int64, money float64, provider string) {
	t.Helper()
	require.NoError(t, DB.Create(&TopUp{
		UserId:          userID,
		Amount:          amount,
		Money:           money,
		TradeNo:         tradeNo,
		PaymentMethod:   provider,
		PaymentProvider: provider,
		Status:          common.TopUpStatusPending,
		CreateTime:      common.GetTimestamp(),
	}).Error)
}

func TestPaymentActivitiesAreGrantedWhenTopUpCompletes(t *testing.T) {
	truncateTables(t)
	insertUserForPaymentGuardTest(t, 901, 0)
	insertActivityTestTopUp(t, "activity-grant-topup", 901, 2, 20, PaymentProviderAlipay)
	require.NoError(t, UpdateLuckyWheelConfig(true, defaultTestLuckyWheelConfig()))
	require.NoError(t, UpdateRechargeActivityConfig(true, defaultTestRechargeActivityConfig()))

	require.NoError(t, RechargeAlipay("activity-grant-topup", "127.0.0.1"))
	require.NoError(t, RechargeAlipay("activity-grant-topup", "127.0.0.1"))

	var sessions int64
	require.NoError(t, DB.Model(&LuckyWheelSession{}).Where("source_order_trade_no = ?", "activity-grant-topup").Count(&sessions).Error)
	assert.Equal(t, int64(1), sessions)

	var chances int64
	require.NoError(t, DB.Model(&RechargeActivityChance{}).Where("source_order_trade_no = ?", "activity-grant-topup").Count(&chances).Error)
	assert.Equal(t, int64(1), chances)

	var session LuckyWheelSession
	require.NoError(t, DB.Where("source_order_trade_no = ?", "activity-grant-topup").First(&session).Error)
	assert.Equal(t, float64(20), session.SourcePayAmount)
	assert.Equal(t, int64(2*common.QuotaPerUnit), session.RewardBaseQuota)
}

func TestDrawLuckyWheelSettlesBonusQuota(t *testing.T) {
	truncateTables(t)
	insertUserForPaymentGuardTest(t, 902, 0)
	insertActivityTestTopUp(t, "activity-wheel-draw", 902, 2, 20, PaymentProviderAlipay)
	require.NoError(t, UpdateLuckyWheelConfig(true, defaultTestLuckyWheelConfig()))

	require.NoError(t, RechargeAlipay("activity-wheel-draw", "127.0.0.1"))
	before := getUserQuotaForPaymentGuardTest(t, 902)

	summary, err := GetLuckyWheelSummary(902)
	require.NoError(t, err)
	require.NotNil(t, summary.ActiveSession)

	result, err := DrawLuckyWheel(902, summary.ActiveSession.Id)
	require.NoError(t, err)

	assert.True(t, result.Settled)
	assert.Equal(t, int64(2*common.QuotaPerUnit), result.SettledBonusQuota)
	assert.Equal(t, before+int(2*common.QuotaPerUnit), getUserQuotaForPaymentGuardTest(t, 902))
}

func TestDrawRechargeActivityCreatesPendingManualFulfillment(t *testing.T) {
	truncateTables(t)
	insertUserForPaymentGuardTest(t, 903, 0)
	insertActivityTestTopUp(t, "activity-recharge-draw", 903, 2, 20, PaymentProviderAlipay)
	require.NoError(t, UpdateRechargeActivityConfig(true, defaultTestRechargeActivityConfig()))

	require.NoError(t, RechargeAlipay("activity-recharge-draw", "127.0.0.1"))
	before := getUserQuotaForPaymentGuardTest(t, 903)

	summary, err := GetRechargeActivitySummary(903)
	require.NoError(t, err)
	require.Len(t, summary.PendingChances, 1)

	result, err := DrawRechargeActivity(903, summary.PendingChances[0].Id)
	require.NoError(t, err)

	assert.Equal(t, RechargeActivityFulfillmentPending, result.Record.FulfillmentStatus)
	assert.Equal(t, int64(0), result.Record.RewardAmount)
	assert.Equal(t, "Manual small prize", result.Record.RewardDescription)
	assert.Equal(t, []string{"small"}, result.Record.EligiblePrizeList)
	assert.Equal(t, before, getUserQuotaForPaymentGuardTest(t, 903))
}

func TestRechargeActivityConfigRejectsInvalidPrizeProbability(t *testing.T) {
	truncateTables(t)
	cfg := defaultTestRechargeActivityConfig()
	cfg.Prizes[0].Probability = 50

	err := UpdateRechargeActivityConfig(true, cfg)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "概率")
}
