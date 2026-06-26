package model

import (
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func insertSubscriptionSaleControlPlan(t *testing.T, id int) *SubscriptionPlan {
	t.Helper()
	plan := &SubscriptionPlan{
		Id:            id,
		Title:         "Sale Control Plan",
		PriceAmount:   9.99,
		Currency:      "USD",
		DurationUnit:  SubscriptionDurationDay,
		DurationValue: 30,
		Enabled:       true,
		TotalAmount:   1000,
	}
	require.NoError(t, DB.Create(plan).Error)
	return plan
}

func TestSubscriptionSaleAvailabilityBlocksActiveSamePlanWhenPurchaseOnceEnabled(t *testing.T) {
	truncateTables(t)
	insertUserForPaymentGuardTest(t, 931, 0)
	plan := insertSubscriptionSaleControlPlan(t, 931)
	plan.PurchaseOncePerActiveSubscription = true
	now := time.Date(2026, 6, 8, 10, 0, 0, 0, time.Local)
	require.NoError(t, DB.Create(&UserSubscription{
		UserId:      931,
		PlanId:      plan.Id,
		AmountTotal: 1000,
		Status:      "active",
		StartTime:   now.Add(-time.Hour).Unix(),
		EndTime:     now.Add(time.Hour).Unix(),
		Source:      "order",
	}).Error)

	availability, err := GetSubscriptionPlanSaleAvailability(931, plan, now)

	require.NoError(t, err)
	assert.False(t, availability.Available)
	assert.Equal(t, SubscriptionPlanSaleBlockPurchaseOnce, availability.BlockReason)
}

func TestSubscriptionSaleAvailabilityBlocksWhenMaxPurchasePerUserReached(t *testing.T) {
	truncateTables(t)
	insertUserForPaymentGuardTest(t, 935, 0)
	plan := insertSubscriptionSaleControlPlan(t, 935)
	plan.MaxPurchasePerUser = 1
	now := time.Date(2026, 6, 8, 10, 0, 0, 0, time.Local)
	require.NoError(t, DB.Create(&UserSubscription{
		UserId:      935,
		PlanId:      plan.Id,
		AmountTotal: 1000,
		Status:      "expired",
		StartTime:   now.Add(-2 * time.Hour).Unix(),
		EndTime:     now.Add(-time.Hour).Unix(),
		Source:      "order",
	}).Error)

	availability, err := GetSubscriptionPlanSaleAvailability(935, plan, now)

	require.NoError(t, err)
	assert.False(t, availability.Available)
	assert.Equal(t, SubscriptionPlanSaleBlockPurchaseMax, availability.BlockReason)
	assert.Equal(t, "已达到该套餐购买上限", availability.BlockMessage)
}

func TestSubscriptionSaleAvailabilityBlocksWhenDailyPurchaseLimitReached(t *testing.T) {
	truncateTables(t)
	insertUserForPaymentGuardTest(t, 932, 0)
	plan := insertSubscriptionSaleControlPlan(t, 932)
	plan.DailyPurchaseLimit = 1
	now := time.Date(2026, 6, 8, 10, 0, 0, 0, time.Local)
	require.NoError(t, DB.Create(&SubscriptionOrder{
		UserId:          932,
		PlanId:          plan.Id,
		Money:           plan.PriceAmount,
		TradeNo:         "daily-limit-success",
		PaymentMethod:   PaymentMethodBalance,
		PaymentProvider: PaymentProviderBalance,
		Status:          common.TopUpStatusSuccess,
		CreateTime:      now.Add(-time.Hour).Unix(),
		CompleteTime:    now.Add(-time.Minute).Unix(),
	}).Error)

	availability, err := GetSubscriptionPlanSaleAvailability(932, plan, now)

	require.NoError(t, err)
	assert.False(t, availability.Available)
	assert.Equal(t, SubscriptionPlanSaleBlockSoldOut, availability.BlockReason)
	require.NotNil(t, availability.DailyPurchaseRemaining)
	assert.Equal(t, 0, *availability.DailyPurchaseRemaining)
}

func TestSubscriptionSaleAvailabilitySupportsCrossMidnightDailyWindow(t *testing.T) {
	truncateTables(t)
	plan := insertSubscriptionSaleControlPlan(t, 933)
	plan.DailySaleStartsAt = "22:00"
	plan.DailySaleEndsAt = "02:00"

	available, err := GetSubscriptionPlanSaleAvailability(933, plan, time.Date(2026, 6, 8, 23, 0, 0, 0, time.Local))
	require.NoError(t, err)
	assert.True(t, available.Available)

	unavailable, err := GetSubscriptionPlanSaleAvailability(933, plan, time.Date(2026, 6, 8, 15, 0, 0, 0, time.Local))
	require.NoError(t, err)
	assert.False(t, unavailable.Available)
	assert.Equal(t, SubscriptionPlanSaleBlockDailyWindow, unavailable.BlockReason)
}

func TestSubscriptionSaleAvailabilityBlocksOffWeeklySaleDay(t *testing.T) {
	truncateTables(t)
	plan := insertSubscriptionSaleControlPlan(t, 934)
	plan.WeeklySaleDays = "[2]"
	monday := time.Date(2026, 6, 8, 10, 0, 0, 0, time.Local)

	availability, err := GetSubscriptionPlanSaleAvailability(934, plan, monday)

	require.NoError(t, err)
	assert.False(t, availability.Available)
	assert.Equal(t, SubscriptionPlanSaleBlockWeeklyDay, availability.BlockReason)
}

func insertSubscriptionListUser(t *testing.T, id int, username string, email string) {
	t.Helper()
	require.NoError(t, DB.Create(&User{
		Id:          id,
		Username:    username,
		DisplayName: username + " display",
		Email:       email,
		AffCode:     username + "_aff",
		Status:      common.UserStatusEnabled,
		Group:       "default",
	}).Error)
}

func insertSubscriptionListRecord(t *testing.T, id int, userId int, planId int, status string, endTime int64) {
	t.Helper()
	require.NoError(t, DB.Create(&UserSubscription{
		Id:          id,
		UserId:      userId,
		PlanId:      planId,
		AmountTotal: 1000,
		AmountUsed:  125,
		Status:      status,
		StartTime:   endTime - 3600,
		EndTime:     endTime,
		Source:      "order",
	}).Error)
}

func TestListPlanUserSubscriptionsReturnsPagedUsersAndFilters(t *testing.T) {
	truncateTables(t)
	plan := insertSubscriptionSaleControlPlan(t, 940)
	otherPlan := insertSubscriptionSaleControlPlan(t, 941)
	now := common.GetTimestamp()
	insertSubscriptionListUser(t, 9401, "alice", "alice@example.com")
	insertSubscriptionListUser(t, 9402, "bob", "bob@example.com")
	insertSubscriptionListUser(t, 9403, "carol", "carol@example.com")
	insertSubscriptionListUser(t, 9404, "dave", "dave@example.com")
	insertSubscriptionListRecord(t, 9401, 9401, plan.Id, "active", now+3600)
	insertSubscriptionListRecord(t, 9402, 9402, plan.Id, "expired", now-3600)
	insertSubscriptionListRecord(t, 9403, 9403, otherPlan.Id, "active", now+3600)
	insertSubscriptionListRecord(t, 9404, 9404, otherPlan.Id, "expired", now-3600)

	records, total, err := ListPlanUserSubscriptions(plan.Id, "", "", 0, 10)

	require.NoError(t, err)
	require.EqualValues(t, 2, total)
	require.Len(t, records, 2)
	assert.Equal(t, 9401, records[0].Subscription.Id)
	assert.Equal(t, 9401, records[0].User.Id)
	assert.Equal(t, "alice", records[0].User.Username)
	assert.Empty(t, records[0].User.Password)
	assert.Nil(t, records[0].User.AccessToken)

	activeRecords, activeTotal, err := ListPlanUserSubscriptions(plan.Id, "active", "", 0, 10)
	require.NoError(t, err)
	require.EqualValues(t, 1, activeTotal)
	require.Len(t, activeRecords, 1)
	assert.Equal(t, 9401, activeRecords[0].Subscription.Id)

	expiredRecords, expiredTotal, err := ListPlanUserSubscriptions(plan.Id, "expired", "", 0, 10)
	require.NoError(t, err)
	require.EqualValues(t, 1, expiredTotal)
	require.Len(t, expiredRecords, 1)
	assert.Equal(t, 9402, expiredRecords[0].Subscription.Id)

	searchRecords, searchTotal, err := ListPlanUserSubscriptions(plan.Id, "", "bob@example.com", 0, 10)
	require.NoError(t, err)
	require.EqualValues(t, 1, searchTotal)
	require.Len(t, searchRecords, 1)
	assert.Equal(t, 9402, searchRecords[0].User.Id)
}

func TestListPlanUserSubscriptionsRejectsInvalidPlan(t *testing.T) {
	truncateTables(t)

	records, total, err := ListPlanUserSubscriptions(0, "", "", 0, 10)

	require.Error(t, err)
	assert.Nil(t, records)
	assert.Zero(t, total)
}

func TestAdminUpdateUserSubscriptionAdjustsUpgradeGroup(t *testing.T) {
	truncateTables(t)
	require.NoError(t, DB.Create(&User{
		Id:       9501,
		Username: "subscription-group-user",
		Group:    "pro",
		Status:   common.UserStatusEnabled,
		AffCode:  "subscription_group_user_aff",
		Password: "secret",
	}).Error)
	oldPlan := insertSubscriptionSaleControlPlan(t, 9501)
	oldPlan.UpgradeGroup = "pro"
	require.NoError(t, DB.Save(oldPlan).Error)
	newPlan := insertSubscriptionSaleControlPlan(t, 9502)
	newPlan.UpgradeGroup = "vip"
	require.NoError(t, DB.Save(newPlan).Error)
	now := common.GetTimestamp()
	require.NoError(t, DB.Create(&UserSubscription{
		Id:            9501,
		UserId:        9501,
		PlanId:        oldPlan.Id,
		AmountTotal:   1000,
		AmountUsed:    100,
		Status:        "active",
		StartTime:     now - 3600,
		EndTime:       now + 3600,
		Source:        "admin",
		UpgradeGroup:  "pro",
		PrevUserGroup: "default",
	}).Error)

	updated, msg, err := AdminUpdateUserSubscription(9501, AdminUpdateUserSubscriptionParams{
		PlanId:        newPlan.Id,
		Status:        "active",
		StartTime:     now - 1800,
		EndTime:       now + 7200,
		AmountTotal:   2000,
		AmountUsed:    50,
		NextResetTime: 0,
	})

	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Contains(t, msg, "vip")
	assert.Equal(t, newPlan.Id, updated.PlanId)
	assert.Equal(t, "vip", updated.UpgradeGroup)
	assert.Equal(t, "default", updated.PrevUserGroup)
	var user User
	require.NoError(t, DB.Where("id = ?", 9501).First(&user).Error)
	assert.Equal(t, "vip", user.Group)

	updated, _, err = AdminUpdateUserSubscription(9501, AdminUpdateUserSubscriptionParams{
		PlanId:        newPlan.Id,
		Status:        "cancelled",
		StartTime:     now - 1800,
		EndTime:       now + 7200,
		AmountTotal:   2000,
		AmountUsed:    50,
		NextResetTime: 0,
	})

	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Empty(t, updated.UpgradeGroup)
	assert.Empty(t, updated.PrevUserGroup)
	require.NoError(t, DB.Where("id = ?", 9501).First(&user).Error)
	assert.Equal(t, "default", user.Group)
}
