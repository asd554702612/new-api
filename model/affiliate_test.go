package model

import (
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func resetAffiliateSettingsForTest(t *testing.T) {
	t.Helper()
	originalEnabled := common.AffiliateEnabled
	originalRate := common.AffiliateRebateRate
	originalSignupRewardEnabled := common.AffiliateSignupRewardEnabled
	originalSignupRewardQuota := common.AffiliateSignupRewardQuota
	originalIdentityEnabled := common.AffiliateIdentityEnabled
	originalIdentityConfig := common.AffiliateIdentityConfig
	originalWithdrawEnabled := common.AffiliateWithdrawEnabled
	originalWithdrawMinQuota := common.AffiliateWithdrawMinQuota
	originalWithdrawDailyLimit := common.AffiliateWithdrawDailyLimit
	originalComplianceConfirmed := operation_setting.GetPaymentSetting().ComplianceConfirmed
	originalComplianceVersion := operation_setting.GetPaymentSetting().ComplianceTermsVersion
	t.Cleanup(func() {
		common.AffiliateEnabled = originalEnabled
		common.AffiliateRebateRate = originalRate
		common.AffiliateSignupRewardEnabled = originalSignupRewardEnabled
		common.AffiliateSignupRewardQuota = originalSignupRewardQuota
		common.AffiliateIdentityEnabled = originalIdentityEnabled
		common.AffiliateIdentityConfig = originalIdentityConfig
		common.AffiliateWithdrawEnabled = originalWithdrawEnabled
		common.AffiliateWithdrawMinQuota = originalWithdrawMinQuota
		common.AffiliateWithdrawDailyLimit = originalWithdrawDailyLimit
		operation_setting.GetPaymentSetting().ComplianceConfirmed = originalComplianceConfirmed
		operation_setting.GetPaymentSetting().ComplianceTermsVersion = originalComplianceVersion
	})
	common.AffiliateEnabled = true
	common.AffiliateRebateRate = 10
	common.AffiliateSignupRewardEnabled = false
	common.AffiliateSignupRewardQuota = 0
	common.AffiliateIdentityEnabled = false
	common.AffiliateIdentityConfig = DefaultAffiliateIdentityConfigJSON()
	common.AffiliateWithdrawEnabled = true
	common.AffiliateWithdrawMinQuota = int(common.QuotaPerUnit)
	common.AffiliateWithdrawDailyLimit = 3
	operation_setting.GetPaymentSetting().ComplianceConfirmed = true
	operation_setting.GetPaymentSetting().ComplianceTermsVersion = operation_setting.CurrentComplianceTermsVersion
}

func insertAffiliateUserForTest(t *testing.T, id int, username string, inviterID int) {
	t.Helper()
	require.NoError(t, DB.Create(&User{
		Id:        id,
		Username:  username,
		Status:    common.UserStatusEnabled,
		InviterId: inviterID,
		AffCode:   username,
	}).Error)
}

func getAffiliateLedgerCountForTest(t *testing.T, action string, tradeNo string) int64 {
	t.Helper()
	var count int64
	query := DB.Model(&AffiliateLedger{}).Where("action = ?", action)
	if tradeNo != "" {
		query = query.Where("source_order_trade_no = ?", tradeNo)
	}
	require.NoError(t, query.Count(&count).Error)
	return count
}

func TestRechargeWaffoAccruesAffiliateRebateOnce(t *testing.T) {
	truncateTables(t)
	resetAffiliateSettingsForTest(t)

	insertAffiliateUserForTest(t, 501, "affiliate-inviter", 0)
	insertAffiliateUserForTest(t, 502, "affiliate-invitee", 501)
	insertTopUpForPaymentGuardTest(t, "affiliate-waffo-order", 502, PaymentProviderWaffo)

	require.NoError(t, RechargeWaffo("affiliate-waffo-order", "127.0.0.1"))
	require.NoError(t, RechargeWaffo("affiliate-waffo-order", "127.0.0.1"))

	var inviter User
	require.NoError(t, DB.Where("id = ?", 501).First(&inviter).Error)
	expectedRebate := int(float64(2*int64(common.QuotaPerUnit)) * 0.10)
	assert.Equal(t, expectedRebate, inviter.AffQuota)
	assert.Equal(t, expectedRebate, inviter.AffHistoryQuota)
	assert.Equal(t, int64(1), getAffiliateLedgerCountForTest(t, AffiliateLedgerActionAccrue, "affiliate-waffo-order"))
}

func TestRechargeDoesNotAccrueAffiliateRebateWhenDisabled(t *testing.T) {
	truncateTables(t)
	resetAffiliateSettingsForTest(t)
	common.AffiliateEnabled = false

	insertAffiliateUserForTest(t, 511, "affiliate-disabled-inviter", 0)
	insertAffiliateUserForTest(t, 512, "affiliate-disabled-invitee", 511)
	insertTopUpForPaymentGuardTest(t, "affiliate-disabled-order", 512, PaymentProviderWaffo)

	require.NoError(t, RechargeWaffo("affiliate-disabled-order", "127.0.0.1"))

	var inviter User
	require.NoError(t, DB.Where("id = ?", 511).First(&inviter).Error)
	assert.Equal(t, 0, inviter.AffQuota)
	assert.Equal(t, int64(0), getAffiliateLedgerCountForTest(t, AffiliateLedgerActionAccrue, "affiliate-disabled-order"))
}

func TestRechargeDoesNotAccrueAffiliateRebateWithoutInviter(t *testing.T) {
	truncateTables(t)
	resetAffiliateSettingsForTest(t)

	insertAffiliateUserForTest(t, 513, "affiliate-no-inviter", 0)
	insertTopUpForPaymentGuardTest(t, "affiliate-no-inviter-order", 513, PaymentProviderWaffo)

	require.NoError(t, RechargeWaffo("affiliate-no-inviter-order", "127.0.0.1"))

	assert.Equal(t, int64(0), getAffiliateLedgerCountForTest(t, AffiliateLedgerActionAccrue, "affiliate-no-inviter-order"))
}

func TestRechargeDoesNotAccrueAffiliateRebateWhenRateIsZero(t *testing.T) {
	truncateTables(t)
	resetAffiliateSettingsForTest(t)
	common.AffiliateRebateRate = 0

	insertAffiliateUserForTest(t, 514, "affiliate-zero-rate-inviter", 0)
	insertAffiliateUserForTest(t, 515, "affiliate-zero-rate-invitee", 514)
	insertTopUpForPaymentGuardTest(t, "affiliate-zero-rate-order", 515, PaymentProviderWaffo)

	require.NoError(t, RechargeWaffo("affiliate-zero-rate-order", "127.0.0.1"))

	var inviter User
	require.NoError(t, DB.Where("id = ?", 514).First(&inviter).Error)
	assert.Equal(t, 0, inviter.AffQuota)
	assert.Equal(t, int64(0), getAffiliateLedgerCountForTest(t, AffiliateLedgerActionAccrue, "affiliate-zero-rate-order"))
}

func TestCompleteSubscriptionOrderAccruesAffiliateRebateOnce(t *testing.T) {
	truncateTables(t)
	resetAffiliateSettingsForTest(t)

	insertAffiliateUserForTest(t, 516, "affiliate-sub-inviter", 0)
	insertAffiliateUserForTest(t, 517, "affiliate-sub-invitee", 516)
	plan := insertSubscriptionPlanForPaymentGuardTest(t, 518)
	insertSubscriptionOrderForPaymentGuardTest(t, "affiliate-sub-order", 517, plan.Id, PaymentProviderStripe)

	require.NoError(t, CompleteSubscriptionOrder("affiliate-sub-order", `{"provider":"stripe"}`, PaymentProviderStripe, ""))
	require.NoError(t, CompleteSubscriptionOrder("affiliate-sub-order", `{"provider":"stripe"}`, PaymentProviderStripe, ""))

	var inviter User
	require.NoError(t, DB.Where("id = ?", 516).First(&inviter).Error)
	expectedRebate := int(9.99 * common.QuotaPerUnit * 0.10)
	assert.Equal(t, expectedRebate, inviter.AffQuota)
	assert.Equal(t, int64(1), getAffiliateLedgerCountForTest(t, AffiliateLedgerActionAccrue, "affiliate-sub-order"))
}

func TestAffiliateExclusiveRateOverridesGlobalRate(t *testing.T) {
	truncateTables(t)
	resetAffiliateSettingsForTest(t)

	insertAffiliateUserForTest(t, 601, "affiliate-rate-inviter", 0)
	insertAffiliateUserForTest(t, 602, "affiliate-rate-invitee", 601)
	require.NoError(t, AdminSetAffiliateUserSettings(601, "VIP601", floatPtrForTest(25)))

	var rebate int
	err := DB.Transaction(func(tx *gorm.DB) error {
		var applyErr error
		rebate, applyErr = ApplyAffiliateRebateTx(tx, 602, int(common.QuotaPerUnit), "exclusive-rate-order", AffiliateSourceOrderTypeTopUp, PaymentProviderStripe)
		return applyErr
	})
	require.NoError(t, err)
	assert.Equal(t, int(common.QuotaPerUnit*0.25), rebate)

	var inviter User
	require.NoError(t, DB.Where("id = ?", 601).First(&inviter).Error)
	assert.Equal(t, "VIP601", inviter.AffCode)
	assert.True(t, inviter.AffCodeCustom)
	assert.NotNil(t, inviter.AffRebateRatePercent)
	assert.Equal(t, 25.0, *inviter.AffRebateRatePercent)
	assert.Equal(t, int(common.QuotaPerUnit*0.25), inviter.AffQuota)
}

func TestAdminBatchRateAndInviteRelationOverwrite(t *testing.T) {
	truncateTables(t)
	resetAffiliateSettingsForTest(t)

	insertAffiliateUserForTest(t, 611, "affiliate-old-inviter", 0)
	insertAffiliateUserForTest(t, 612, "affiliate-new-inviter", 0)
	insertAffiliateUserForTest(t, 613, "affiliate-manual-invitee", 611)
	require.NoError(t, DB.Model(&User{}).Where("id = ?", 611).Update("aff_count", 1).Error)

	require.NoError(t, AdminBatchSetAffiliateRebateRate([]int{611, 612}, floatPtrForTest(18)))
	_, err := AdminSetInviteRelation(612, 613, false)
	require.Error(t, err)

	relation, err := AdminSetInviteRelation(612, 613, true)
	require.NoError(t, err)
	assert.Equal(t, 612, relation.InviterId)
	assert.Equal(t, 611, relation.PreviousInviterId)

	var oldInviter, newInviter, invitee User
	require.NoError(t, DB.Where("id = ?", 611).First(&oldInviter).Error)
	require.NoError(t, DB.Where("id = ?", 612).First(&newInviter).Error)
	require.NoError(t, DB.Where("id = ?", 613).First(&invitee).Error)
	assert.Equal(t, 0, oldInviter.AffCount)
	assert.Equal(t, 1, newInviter.AffCount)
	assert.Equal(t, 612, invitee.InviterId)
	require.NotNil(t, oldInviter.AffRebateRatePercent)
	require.NotNil(t, newInviter.AffRebateRatePercent)
	assert.Equal(t, 18.0, *oldInviter.AffRebateRatePercent)
	assert.Equal(t, 18.0, *newInviter.AffRebateRatePercent)

	overview, err := GetAffiliateUserOverview(613)
	require.NoError(t, err)
	assert.Equal(t, 612, overview.InviterId)
	assert.Equal(t, "affiliate-new-inviter", overview.InviterUsername)
}

func TestAffiliateSignupBonusIsIdempotent(t *testing.T) {
	truncateTables(t)
	resetAffiliateSettingsForTest(t)
	common.AffiliateSignupRewardEnabled = true
	common.AffiliateSignupRewardQuota = int(common.QuotaPerUnit)

	insertAffiliateUserForTest(t, 621, "affiliate-signup-inviter", 0)
	insertAffiliateUserForTest(t, 622, "affiliate-signup-invitee", 621)

	applied, err := ApplyAffiliateSignupBonus(622)
	require.NoError(t, err)
	assert.True(t, applied)
	applied, err = ApplyAffiliateSignupBonus(622)
	require.NoError(t, err)
	assert.False(t, applied)

	var inviter User
	require.NoError(t, DB.Where("id = ?", 621).First(&inviter).Error)
	assert.Equal(t, int(common.QuotaPerUnit), inviter.AffQuota)
	assert.Equal(t, int(common.QuotaPerUnit), inviter.AffHistoryQuota)
	assert.Equal(t, int64(1), getAffiliateLedgerCountForTest(t, AffiliateLedgerActionSignupBonus, ""))
}

func TestAffiliateIdentitySkipsRiskInviteeAndAppliesMultiplier(t *testing.T) {
	truncateTables(t)
	resetAffiliateSettingsForTest(t)
	common.AffiliateIdentityEnabled = true
	common.AffiliateIdentityConfig = `{"inviter_rate_multiplier":0.6,"invitee_rate_multiplier":0.7,"duration_hours":720,"qualified_invitee_count":1,"qualified_pay_amount":1,"eligible_order_types":["topup","subscription"],"fingerprint_enforcement_enabled":true,"max_accounts_per_fingerprint_hash":1}`

	insertAffiliateUserForTest(t, 631, "affiliate-identity-inviter", 0)
	insertAffiliateUserForTest(t, 632, "affiliate-identity-risk", 631)
	insertAffiliateUserForTest(t, 633, "affiliate-identity-ok", 631)
	require.NoError(t, RecordAffiliateSignupFingerprint(632, AffiliateSignupFingerprintInput{CompositeHash: "same-fp"}))
	require.NoError(t, RecordAffiliateSignupFingerprint(633, AffiliateSignupFingerprintInput{CompositeHash: "same-fp"}))
	require.NoError(t, DB.Model(&AffiliateSignupFingerprint{}).Where("user_id = ?", 632).Updates(map[string]any{"risk_flagged": true, "duplicate_count": 1, "risk_reason": "duplicate_fingerprint"}).Error)
	require.NoError(t, DB.Create(&AffiliateLedger{UserId: 631, RelatedUserId: 632, Action: AffiliateLedgerActionAccrue, Quota: int(common.QuotaPerUnit), SourceOrderType: AffiliateSourceOrderTypeTopUp}).Error)
	require.NoError(t, RefreshAffiliateIdentitiesForInviter(631))
	assert.Nil(t, GetActiveAffiliateIdentity(631))

	require.NoError(t, DB.Model(&AffiliateSignupFingerprint{}).Where("user_id = ?", 633).Updates(map[string]any{"risk_flagged": false, "duplicate_count": 0, "risk_reason": ""}).Error)
	require.NoError(t, DB.Create(&AffiliateLedger{UserId: 631, RelatedUserId: 633, Action: AffiliateLedgerActionAccrue, Quota: int(common.QuotaPerUnit), SourceOrderType: AffiliateSourceOrderTypeTopUp}).Error)
	require.NoError(t, RefreshAffiliateIdentitiesForInviter(631))

	identity := GetActiveAffiliateIdentity(631)
	require.NotNil(t, identity)
	assert.Equal(t, AffiliateIdentityTypeInviter, identity.IdentityType)
	assert.Equal(t, 0.6, ResolveAffiliateIdentityMultiplier(631, 1.0))
}

func TestAffiliateWithdrawalWritesActionLedgersAndStrictStatusFlow(t *testing.T) {
	truncateTables(t)
	resetAffiliateSettingsForTest(t)

	require.NoError(t, DB.Create(&User{
		Id:       641,
		Username: "affiliate-withdraw-actions",
		Status:   common.UserStatusEnabled,
		AffQuota: int(3 * common.QuotaPerUnit),
	}).Error)
	withdrawal, err := CreateAffiliateWithdrawal(641, int(common.QuotaPerUnit), "wechat_manual", "wechat account")
	require.NoError(t, err)
	assert.Equal(t, int64(1), getAffiliateLedgerCountForTest(t, AffiliateLedgerActionWithdrawRequest, ""))

	_, err = FailAffiliateWithdrawal(withdrawal.Id, 1, "not approved")
	require.Error(t, err)
	_, err = RejectAffiliateWithdrawal(withdrawal.Id, 1, "")
	require.Error(t, err)
	rejected, err := RejectAffiliateWithdrawal(withdrawal.Id, 1, "bad account")
	require.NoError(t, err)
	assert.Equal(t, AffiliateWithdrawalStatusRejected, rejected.Status)
	assert.Equal(t, int64(1), getAffiliateLedgerCountForTest(t, AffiliateLedgerActionWithdrawReject, ""))

	second, err := CreateAffiliateWithdrawal(641, int(common.QuotaPerUnit), "wechat_manual", "wechat account")
	require.NoError(t, err)
	_, err = MarkAffiliateWithdrawalPaid(second.Id, 1, "wechat", "trade-before-approve", "")
	require.Error(t, err)
	_, err = ApproveAffiliateWithdrawal(second.Id, 1, "ok")
	require.NoError(t, err)
	_, err = MarkAffiliateWithdrawalPaid(second.Id, 1, "", "trade-no-channel", "")
	require.Error(t, err)
	paid, err := MarkAffiliateWithdrawalPaid(second.Id, 1, "wechat", "trade-ok", "done")
	require.NoError(t, err)
	assert.Equal(t, AffiliateWithdrawalStatusPaid, paid.Status)
	assert.Equal(t, int64(1), getAffiliateLedgerCountForTest(t, AffiliateLedgerActionWithdrawPaid, ""))
	var paidLedger AffiliateLedger
	require.NoError(t, DB.Where("action = ?", AffiliateLedgerActionWithdrawPaid).First(&paidLedger).Error)
	assert.Equal(t, int(2*common.QuotaPerUnit), paidLedger.BalanceAfter)
	assert.Equal(t, 0, paidLedger.HistoryAfter)
}

func TestTransferAffQuotaWritesAffiliateLedger(t *testing.T) {
	truncateTables(t)
	resetAffiliateSettingsForTest(t)

	user := &User{
		Id:              521,
		Username:        "affiliate-transfer-user",
		Status:          common.UserStatusEnabled,
		AffQuota:        int(common.QuotaPerUnit),
		AffHistoryQuota: int(common.QuotaPerUnit),
		CreatedAt:       time.Now().Unix(),
	}
	require.NoError(t, DB.Create(user).Error)

	require.NoError(t, user.TransferAffQuotaToQuota(int(common.QuotaPerUnit)))

	var reloaded User
	require.NoError(t, DB.Where("id = ?", 521).First(&reloaded).Error)
	assert.Equal(t, 0, reloaded.AffQuota)
	assert.Equal(t, int(common.QuotaPerUnit), reloaded.Quota)
	assert.Equal(t, int64(1), getAffiliateLedgerCountForTest(t, AffiliateLedgerActionTransfer, ""))
}

func floatPtrForTest(v float64) *float64 {
	return &v
}

func TestCreateAffiliateWithdrawalDeductsAffQuota(t *testing.T) {
	truncateTables(t)
	resetAffiliateSettingsForTest(t)

	user := &User{
		Id:              531,
		Username:        "affiliate-withdraw-user",
		Status:          common.UserStatusEnabled,
		AffQuota:        int(2 * common.QuotaPerUnit),
		AffHistoryQuota: int(2 * common.QuotaPerUnit),
	}
	require.NoError(t, DB.Create(user).Error)

	withdrawal, err := CreateAffiliateWithdrawal(531, int(common.QuotaPerUnit), "wechat_manual", "wechat account")
	require.NoError(t, err)
	require.NotNil(t, withdrawal)
	assert.Equal(t, AffiliateWithdrawalStatusPendingReview, withdrawal.Status)

	var reloaded User
	require.NoError(t, DB.Where("id = ?", 531).First(&reloaded).Error)
	assert.Equal(t, int(common.QuotaPerUnit), reloaded.AffQuota)
}

func TestCreateAffiliateWithdrawalValidatesSettingsAndBalance(t *testing.T) {
	testCases := []struct {
		name   string
		mutate func()
		quota  int
	}{
		{
			name: "disabled",
			mutate: func() {
				common.AffiliateWithdrawEnabled = false
			},
			quota: int(common.QuotaPerUnit),
		},
		{
			name: "compliance not confirmed",
			mutate: func() {
				operation_setting.GetPaymentSetting().ComplianceConfirmed = false
			},
			quota: int(common.QuotaPerUnit),
		},
		{
			name:  "below minimum",
			quota: int(common.QuotaPerUnit) - 1,
		},
		{
			name:  "insufficient balance",
			quota: int(3 * common.QuotaPerUnit),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			truncateTables(t)
			resetAffiliateSettingsForTest(t)
			if tc.mutate != nil {
				tc.mutate()
			}
			require.NoError(t, DB.Create(&User{
				Id:       532,
				Username: "affiliate-withdraw-validation",
				Status:   common.UserStatusEnabled,
				AffQuota: int(2 * common.QuotaPerUnit),
			}).Error)

			withdrawal, err := CreateAffiliateWithdrawal(532, tc.quota, "wechat_manual", "wechat account")
			require.Error(t, err)
			assert.Nil(t, withdrawal)

			var reloaded User
			require.NoError(t, DB.Where("id = ?", 532).First(&reloaded).Error)
			assert.Equal(t, int(2*common.QuotaPerUnit), reloaded.AffQuota)
		})
	}
}

func TestCreateAffiliateWithdrawalEnforcesDailyLimit(t *testing.T) {
	truncateTables(t)
	resetAffiliateSettingsForTest(t)
	common.AffiliateWithdrawDailyLimit = 1

	require.NoError(t, DB.Create(&User{
		Id:       533,
		Username: "affiliate-withdraw-limit",
		Status:   common.UserStatusEnabled,
		AffQuota: int(3 * common.QuotaPerUnit),
	}).Error)

	first, err := CreateAffiliateWithdrawal(533, int(common.QuotaPerUnit), "wechat_manual", "wechat account")
	require.NoError(t, err)
	require.NotNil(t, first)

	second, err := CreateAffiliateWithdrawal(533, int(common.QuotaPerUnit), "wechat_manual", "wechat account")
	require.Error(t, err)
	assert.Nil(t, second)
}

func TestAffiliateWithdrawalRejectAndFailRefundQuotaOnce(t *testing.T) {
	testCases := []struct {
		name   string
		setup  func(id int, adminID int)
		action func(id int, adminID int) (*AffiliateWithdrawal, error)
		status string
	}{
		{
			name: "reject",
			action: func(id int, adminID int) (*AffiliateWithdrawal, error) {
				return RejectAffiliateWithdrawal(id, adminID, "invalid account")
			},
			status: AffiliateWithdrawalStatusRejected,
		},
		{
			name: "fail",
			setup: func(id int, adminID int) {
				approved, err := ApproveAffiliateWithdrawal(id, adminID, "ok")
				require.NoError(t, err)
				require.NotNil(t, approved)
			},
			action: func(id int, adminID int) (*AffiliateWithdrawal, error) {
				return FailAffiliateWithdrawal(id, adminID, "transfer failed")
			},
			status: AffiliateWithdrawalStatusFailed,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			truncateTables(t)
			resetAffiliateSettingsForTest(t)

			require.NoError(t, DB.Create(&User{
				Id:       534,
				Username: "affiliate-withdraw-refund",
				Status:   common.UserStatusEnabled,
				AffQuota: int(2 * common.QuotaPerUnit),
			}).Error)
			withdrawal, err := CreateAffiliateWithdrawal(534, int(common.QuotaPerUnit), "wechat_manual", "wechat account")
			require.NoError(t, err)
			if tc.setup != nil {
				tc.setup(withdrawal.Id, 1)
			}

			updated, err := tc.action(withdrawal.Id, 1)
			require.NoError(t, err)
			assert.Equal(t, tc.status, updated.Status)

			again, err := tc.action(withdrawal.Id, 1)
			require.Error(t, err)
			assert.Nil(t, again)

			var reloaded User
			require.NoError(t, DB.Where("id = ?", 534).First(&reloaded).Error)
			assert.Equal(t, int(2*common.QuotaPerUnit), reloaded.AffQuota)
		})
	}
}

func TestAffiliateWithdrawalApproveAndPaidFlow(t *testing.T) {
	truncateTables(t)
	resetAffiliateSettingsForTest(t)

	require.NoError(t, DB.Create(&User{
		Id:       535,
		Username: "affiliate-withdraw-paid",
		Status:   common.UserStatusEnabled,
		AffQuota: int(2 * common.QuotaPerUnit),
	}).Error)
	withdrawal, err := CreateAffiliateWithdrawal(535, int(common.QuotaPerUnit), "wechat_manual", "wechat account")
	require.NoError(t, err)

	approved, err := ApproveAffiliateWithdrawal(withdrawal.Id, 1, "approved")
	require.NoError(t, err)
	assert.Equal(t, AffiliateWithdrawalStatusApproved, approved.Status)

	emptyPaid, err := MarkAffiliateWithdrawalPaid(withdrawal.Id, 1, "", "", "")
	require.Error(t, err)
	assert.Nil(t, emptyPaid)

	paid, err := MarkAffiliateWithdrawalPaid(withdrawal.Id, 1, "wechat", "wx-trade-no", "paid")
	require.NoError(t, err)
	assert.Equal(t, AffiliateWithdrawalStatusPaid, paid.Status)
	assert.Equal(t, "wechat", paid.PayoutChannel)
	assert.Equal(t, "wx-trade-no", paid.PayoutTradeNo)

	rejected, err := RejectAffiliateWithdrawal(withdrawal.Id, 1, "too late")
	require.Error(t, err)
	assert.Nil(t, rejected)

	var reloaded User
	require.NoError(t, DB.Where("id = ?", 535).First(&reloaded).Error)
	assert.Equal(t, int(common.QuotaPerUnit), reloaded.AffQuota)
}

func TestAffiliateAdminRecordQueriesIncludeUserInfoAndFilters(t *testing.T) {
	truncateTables(t)
	resetAffiliateSettingsForTest(t)

	insertAffiliateUserForTest(t, 541, "admin-aff-inviter", 0)
	insertAffiliateUserForTest(t, 542, "admin-aff-invitee", 541)
	require.NoError(t, DB.Model(&User{}).Where("id = ?", 541).Updates(map[string]any{
		"email": "inviter@example.com",
	}).Error)
	require.NoError(t, DB.Model(&User{}).Where("id = ?", 542).Updates(map[string]any{
		"email": "invitee@example.com",
	}).Error)
	require.NoError(t, DB.Create(&AffiliateLedger{
		UserId:             541,
		RelatedUserId:      542,
		Action:             AffiliateLedgerActionAccrue,
		Quota:              int(common.QuotaPerUnit),
		BalanceAfter:       int(common.QuotaPerUnit),
		HistoryAfter:       int(common.QuotaPerUnit),
		SourceOrderTradeNo: "admin-aff-order",
		SourceOrderType:    AffiliateSourceOrderTypeTopUp,
		PaymentMethod:      PaymentProviderStripe,
	}).Error)
	require.NoError(t, DB.Create(&AffiliateLedger{
		UserId:       541,
		Action:       AffiliateLedgerActionTransfer,
		Quota:        int(common.QuotaPerUnit),
		BalanceAfter: 0,
		HistoryAfter: int(common.QuotaPerUnit),
		Remark:       "aff_transfer",
	}).Error)
	require.NoError(t, DB.Create(&AffiliateWithdrawal{
		UserId:            541,
		Quota:             int(common.QuotaPerUnit),
		Status:            AffiliateWithdrawalStatusPendingReview,
		PayoutMethod:      AffiliateWithdrawalPayoutMethodWechatManual,
		PayoutAccountNote: "wechat",
	}).Error)

	pageInfo := &common.PageInfo{Page: 1, PageSize: 20}
	filter := AffiliateRecordFilter{Search: "invitee@example.com"}
	invites, inviteTotal, err := GetAffiliateInviteRecords(pageInfo, filter)
	require.NoError(t, err)
	require.Equal(t, int64(1), inviteTotal)
	require.Len(t, invites, 1)
	assert.Equal(t, "admin-aff-invitee", invites[0].Username)
	assert.Equal(t, "admin-aff-inviter", invites[0].InviterUsername)

	rebates, rebateTotal, err := GetAffiliateRebateLedgers(pageInfo, filter)
	require.NoError(t, err)
	require.Equal(t, int64(1), rebateTotal)
	require.Len(t, rebates, 1)
	assert.Equal(t, "inviter@example.com", rebates[0].Email)
	assert.Equal(t, "invitee@example.com", rebates[0].RelatedEmail)

	transfers, transferTotal, err := GetAffiliateTransferLedgers(pageInfo, AffiliateRecordFilter{Search: "inviter@example.com"})
	require.NoError(t, err)
	require.Equal(t, int64(1), transferTotal)
	require.Len(t, transfers, 1)
	assert.Equal(t, "admin-aff-inviter", transfers[0].Username)

	withdrawals, withdrawalTotal, err := GetAffiliateWithdrawalRecords(pageInfo, AffiliateRecordFilter{
		Search: "inviter@example.com",
		Status: AffiliateWithdrawalStatusPendingReview,
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), withdrawalTotal)
	require.Len(t, withdrawals, 1)
	assert.Equal(t, "inviter@example.com", withdrawals[0].Email)
}
