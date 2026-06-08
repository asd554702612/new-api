package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/stretchr/testify/require"
)

func setWeeklyQuotaTestSetting(t *testing.T, enabled bool, amount int) {
	t.Helper()
	setting := operation_setting.GetWeeklyQuotaSetting()
	originalEnabled := setting.Enabled
	originalAmount := setting.Amount
	originalPhoneVerificationEnabled := common.PhoneVerificationEnabled
	t.Cleanup(func() {
		setting.Enabled = originalEnabled
		setting.Amount = originalAmount
		common.PhoneVerificationEnabled = originalPhoneVerificationEnabled
	})
	setting.Enabled = enabled
	setting.Amount = amount
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

func TestGetWeeklyQuotaStatusDisabled(t *testing.T) {
	truncateTables(t)
	setWeeklyQuotaTestSetting(t, false, 500)
	user := createWeeklyQuotaTestUser(t, 7001, 1_700_000_000, "")

	status, err := GetWeeklyQuotaStatus(user.Id, 1_700_000_100)

	require.NoError(t, err)
	require.False(t, status.Enabled)
	require.Equal(t, WeeklyQuotaStatusDisabled, status.Status)
	require.Equal(t, 500, status.Amount)
}

func TestClaimWeeklyQuotaIncreasesQuotaAndBlocksDuplicateWindow(t *testing.T) {
	truncateTables(t)
	setWeeklyQuotaTestSetting(t, true, 500)
	common.PhoneVerificationEnabled = false
	user := createWeeklyQuotaTestUser(t, 7002, 1_700_000_000, "")

	claim, err := ClaimWeeklyQuota(user.Id, 1_700_000_100)

	require.NoError(t, err)
	require.Equal(t, 500, claim.QuotaAwarded)
	require.Equal(t, 600, getWeeklyQuotaTestUserQuota(t, user.Id))

	_, err = ClaimWeeklyQuota(user.Id, 1_700_000_200)
	require.ErrorIs(t, err, ErrWeeklyQuotaAlreadyClaimed)
	require.Equal(t, 600, getWeeklyQuotaTestUserQuota(t, user.Id))
}

func TestClaimWeeklyQuotaNextWindowAllowed(t *testing.T) {
	truncateTables(t)
	setWeeklyQuotaTestSetting(t, true, 500)
	common.PhoneVerificationEnabled = false
	user := createWeeklyQuotaTestUser(t, 7003, 1_700_000_000, "")

	_, err := ClaimWeeklyQuota(user.Id, 1_700_000_100)
	require.NoError(t, err)
	claim, err := ClaimWeeklyQuota(user.Id, 1_700_000_000+7*24*60*60+10)

	require.NoError(t, err)
	require.Equal(t, int64(1_700_000_000+7*24*60*60), claim.WindowStart)
	require.Equal(t, 1100, getWeeklyQuotaTestUserQuota(t, user.Id))
}

func TestClaimWeeklyQuotaRequiresBoundPhoneWhenPhoneVerificationEnabled(t *testing.T) {
	truncateTables(t)
	setWeeklyQuotaTestSetting(t, true, 500)
	common.PhoneVerificationEnabled = true
	user := createWeeklyQuotaTestUser(t, 7004, 1_700_000_000, "")

	_, err := ClaimWeeklyQuota(user.Id, 1_700_000_100)

	require.ErrorIs(t, err, ErrWeeklyQuotaPhoneRequired)
	require.Equal(t, 100, getWeeklyQuotaTestUserQuota(t, user.Id))
}

func TestClaimWeeklyQuotaAllowsBoundPhone(t *testing.T) {
	truncateTables(t)
	setWeeklyQuotaTestSetting(t, true, 500)
	common.PhoneVerificationEnabled = true
	user := createWeeklyQuotaTestUser(t, 7005, 1_700_000_000, "+8613800138000")

	claim, err := ClaimWeeklyQuota(user.Id, 1_700_000_100)

	require.NoError(t, err)
	require.Equal(t, 500, claim.QuotaAwarded)
	require.Equal(t, 600, getWeeklyQuotaTestUserQuota(t, user.Id))
}
