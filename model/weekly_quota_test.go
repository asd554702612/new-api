package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/stretchr/testify/require"
)

func setWeeklyQuotaTestSetting(t *testing.T, enabled bool, planId int, periodDays int) {
	t.Helper()
	setting := operation_setting.GetWeeklyQuotaSetting()
	originalEnabled := setting.Enabled
	originalAmount := setting.Amount
	originalPlanId := setting.PlanId
	originalPeriodDays := setting.PeriodDays
	originalPhoneVerificationEnabled := common.PhoneVerificationEnabled
	t.Cleanup(func() {
		setting.Enabled = originalEnabled
		setting.Amount = originalAmount
		setting.PlanId = originalPlanId
		setting.PeriodDays = originalPeriodDays
		common.PhoneVerificationEnabled = originalPhoneVerificationEnabled
	})
	setting.Enabled = enabled
	setting.Amount = 0
	setting.PlanId = planId
	setting.PeriodDays = periodDays
}

func createWeeklyQuotaTestUser(t *testing.T, id int, createdAt int64, phoneNumber string) *User {
	t.Helper()
	user := &User{
		Id:          id,
		Username:    "weekly_quota_user",
		Password:    "hashed",
		Status:      common.UserStatusEnabled,
		Quota:       100,
		PhoneNumber: phoneNumber,
		CreatedAt:   createdAt,
	}
	require.NoError(t, DB.Create(user).Error)
	return user
}

func getWeeklyQuotaTestUserQuota(t *testing.T, userId int) int {
	t.Helper()
	var user User
	require.NoError(t, DB.Select("quota").Where("id = ?", userId).First(&user).Error)
	return user.Quota
}

func createWeeklyQuotaTestPlan(t *testing.T, id int, enabled bool) *SubscriptionPlan {
	t.Helper()
	plan := &SubscriptionPlan{
		Id:            id,
		Title:         "Gift Plan",
		Subtitle:      "Claimable subscription",
		PriceAmount:   0,
		Currency:      "USD",
		DurationUnit:  SubscriptionDurationDay,
		DurationValue: 30,
		Enabled:       enabled,
		TotalAmount:   12345,
	}
	require.NoError(t, DB.Create(plan).Error)
	return plan
}

func TestGetWeeklyQuotaStatusDisabled(t *testing.T) {
	truncateTables(t)
	plan := createWeeklyQuotaTestPlan(t, 7101, true)
	setWeeklyQuotaTestSetting(t, false, plan.Id, 7)
	user := createWeeklyQuotaTestUser(t, 7001, 1_700_000_000, "")

	status, err := GetWeeklyQuotaStatus(user.Id, 1_700_000_100)

	require.NoError(t, err)
	require.False(t, status.Enabled)
	require.Equal(t, WeeklyQuotaStatusDisabled, status.Status)
	require.Equal(t, plan.Id, status.PlanId)
	require.Equal(t, 7, status.PeriodDays)
}

func TestClaimWeeklyQuotaCreatesGiftSubscriptionAndBlocksDuplicateWindow(t *testing.T) {
	truncateTables(t)
	plan := createWeeklyQuotaTestPlan(t, 7102, true)
	setWeeklyQuotaTestSetting(t, true, plan.Id, 7)
	common.PhoneVerificationEnabled = false
	user := createWeeklyQuotaTestUser(t, 7002, 1_700_000_000, "+8613800138000")

	claim, err := ClaimWeeklyQuota(user.Id, 1_700_000_100)

	require.NoError(t, err)
	require.Equal(t, plan.Id, claim.PlanId)
	require.NotZero(t, claim.UserSubscriptionId)
	require.Equal(t, 100, getWeeklyQuotaTestUserQuota(t, user.Id))

	var sub UserSubscription
	require.NoError(t, DB.Where("id = ?", claim.UserSubscriptionId).First(&sub).Error)
	require.Equal(t, user.Id, sub.UserId)
	require.Equal(t, plan.Id, sub.PlanId)
	require.Equal(t, SubscriptionSourceGiftClaim, sub.Source)
	require.Equal(t, plan.TotalAmount, sub.AmountTotal)
	require.Equal(t, "active", sub.Status)

	_, err = ClaimWeeklyQuota(user.Id, 1_700_000_200)
	require.ErrorIs(t, err, ErrWeeklyQuotaAlreadyClaimed)
	require.Equal(t, 100, getWeeklyQuotaTestUserQuota(t, user.Id))
}

func TestClaimWeeklyQuotaNextWindowBlockedWhenSamePlanStillActive(t *testing.T) {
	truncateTables(t)
	plan := createWeeklyQuotaTestPlan(t, 7103, true)
	setWeeklyQuotaTestSetting(t, true, plan.Id, 7)
	common.PhoneVerificationEnabled = false
	user := createWeeklyQuotaTestUser(t, 7003, 1_700_000_000, "+8613800138000")

	_, err := ClaimWeeklyQuota(user.Id, 1_700_000_100)
	require.NoError(t, err)
	_, err = ClaimWeeklyQuota(user.Id, 1_700_000_000+7*24*60*60+10)

	require.ErrorIs(t, err, ErrWeeklyQuotaActiveSubscriptionExists)
	require.Equal(t, 100, getWeeklyQuotaTestUserQuota(t, user.Id))
}

func TestClaimWeeklyQuotaAllowsDisabledGiftPlan(t *testing.T) {
	truncateTables(t)
	plan := createWeeklyQuotaTestPlan(t, 7104, false)
	setWeeklyQuotaTestSetting(t, true, plan.Id, 7)
	common.PhoneVerificationEnabled = false
	user := createWeeklyQuotaTestUser(t, 7007, 1_700_000_000, "+8613800138000")

	claim, err := ClaimWeeklyQuota(user.Id, 1_700_000_100)

	require.NoError(t, err)
	require.Equal(t, plan.Id, claim.PlanId)
	require.NotZero(t, claim.UserSubscriptionId)
}

func TestClaimWeeklyQuotaRequiresBoundPhoneWhenPhoneVerificationEnabled(t *testing.T) {
	truncateTables(t)
	plan := createWeeklyQuotaTestPlan(t, 7105, true)
	setWeeklyQuotaTestSetting(t, true, plan.Id, 7)
	common.PhoneVerificationEnabled = true
	user := createWeeklyQuotaTestUser(t, 7004, 1_700_000_000, "")

	_, err := ClaimWeeklyQuota(user.Id, 1_700_000_100)

	require.ErrorIs(t, err, ErrWeeklyQuotaPhoneRequired)
	require.Equal(t, 100, getWeeklyQuotaTestUserQuota(t, user.Id))
}

func TestClaimWeeklyQuotaRequiresBoundPhoneWhenPhoneVerificationDisabled(t *testing.T) {
	truncateTables(t)
	plan := createWeeklyQuotaTestPlan(t, 7106, true)
	setWeeklyQuotaTestSetting(t, true, plan.Id, 7)
	common.PhoneVerificationEnabled = false
	user := createWeeklyQuotaTestUser(t, 7006, 1_700_000_000, "")

	status, err := GetWeeklyQuotaStatus(user.Id, 1_700_000_100)
	require.NoError(t, err)
	require.True(t, status.Enabled)
	require.Equal(t, WeeklyQuotaStatusPhoneRequired, status.Status)
	require.Equal(t, plan.Id, status.PlanId)

	_, err = ClaimWeeklyQuota(user.Id, 1_700_000_100)
	require.ErrorIs(t, err, ErrWeeklyQuotaPhoneRequired)
	require.Equal(t, 100, getWeeklyQuotaTestUserQuota(t, user.Id))
}

func TestClaimWeeklyQuotaAllowsBoundPhone(t *testing.T) {
	truncateTables(t)
	plan := createWeeklyQuotaTestPlan(t, 7107, true)
	setWeeklyQuotaTestSetting(t, true, plan.Id, 7)
	common.PhoneVerificationEnabled = true
	user := createWeeklyQuotaTestUser(t, 7005, 1_700_000_000, "+8613800138000")

	claim, err := ClaimWeeklyQuota(user.Id, 1_700_000_100)

	require.NoError(t, err)
	require.Equal(t, plan.Id, claim.PlanId)
	require.NotZero(t, claim.UserSubscriptionId)
	require.Equal(t, 100, getWeeklyQuotaTestUserQuota(t, user.Id))
}
