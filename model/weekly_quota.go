package model

import (
	"errors"
	"time"

	"github.com/QuantumNous/new-api/setting/operation_setting"
	"gorm.io/gorm"
)

const (
	WeeklyQuotaStatusDisabled      = "disabled"
	WeeklyQuotaStatusClaimable     = "claimable"
	WeeklyQuotaStatusClaimed       = "claimed"
	WeeklyQuotaStatusPhoneRequired = "phone_required"
	WeeklyQuotaStatusActive        = "active_subscription"

	weeklyQuotaWindowSeconds = int64(7 * 24 * 60 * 60)
)

var (
	ErrWeeklyQuotaDisabled                 = errors.New("领取套餐未启用")
	ErrWeeklyQuotaAlreadyClaimed           = errors.New("本周期套餐已领取")
	ErrWeeklyQuotaPhoneRequired            = errors.New("领取套餐前需要先绑定手机号")
	ErrWeeklyQuotaUserNotFound             = errors.New("用户不存在")
	ErrWeeklyQuotaActiveSubscriptionExists = errors.New("当前套餐仍在有效期内，暂不可重复领取")
)

type WeeklyQuotaClaim struct {
	Id                 int   `json:"id" gorm:"primaryKey;autoIncrement"`
	UserId             int   `json:"user_id" gorm:"not null;uniqueIndex:idx_user_weekly_quota_window"`
	WindowStart        int64 `json:"window_start" gorm:"bigint;not null;uniqueIndex:idx_user_weekly_quota_window"`
	WindowEnd          int64 `json:"window_end" gorm:"bigint;not null;index"`
	QuotaAwarded       int   `json:"quota_awarded" gorm:"not null"`
	PlanId             int   `json:"plan_id" gorm:"type:int;default:0;index"`
	UserSubscriptionId int   `json:"user_subscription_id" gorm:"type:int;default:0;index"`
	CreatedAt          int64 `json:"created_at" gorm:"bigint"`
}

func (WeeklyQuotaClaim) TableName() string {
	return "weekly_quota_claims"
}

type WeeklyQuotaStatus struct {
	Enabled         bool                 `json:"enabled"`
	Amount          int                  `json:"amount"`
	PlanId          int                  `json:"plan_id"`
	PeriodDays      int                  `json:"period_days"`
	Plan            *WeeklyQuotaPlanInfo `json:"plan,omitempty"`
	Status          string               `json:"status"`
	WindowStartedAt int64                `json:"window_started_at"`
	WindowEndsAt    int64                `json:"window_ends_at"`
	ClaimedAt       int64                `json:"claimed_at"`
	NextClaimAt     int64                `json:"next_claim_at"`
	TotalClaimCount int64                `json:"total_claim_count"`
	TotalClaimQuota int64                `json:"total_claim_quota"`
}

type WeeklyQuotaPlanInfo struct {
	Id                      int    `json:"id"`
	Title                   string `json:"title"`
	Subtitle                string `json:"subtitle"`
	TotalAmount             int64  `json:"total_amount"`
	DurationUnit            string `json:"duration_unit"`
	DurationValue           int    `json:"duration_value"`
	CustomSeconds           int64  `json:"custom_seconds"`
	QuotaResetPeriod        string `json:"quota_reset_period"`
	QuotaResetCustomSeconds int64  `json:"quota_reset_custom_seconds"`
}

func getWeeklyQuotaWindow(createdAt int64, now int64, periodDays int) (int64, int64) {
	if now <= 0 {
		now = time.Now().Unix()
	}
	periodSeconds := int64(periodDays) * 24 * 60 * 60
	if periodSeconds <= 0 {
		periodSeconds = weeklyQuotaWindowSeconds
	}
	if createdAt < 0 {
		createdAt = 0
	}
	if now < createdAt {
		return createdAt, createdAt + periodSeconds
	}
	windowIndex := (now - createdAt) / periodSeconds
	windowStart := createdAt + windowIndex*periodSeconds
	return windowStart, windowStart + periodSeconds
}

func weeklyQuotaPlanInfo(plan *SubscriptionPlan) *WeeklyQuotaPlanInfo {
	if plan == nil {
		return nil
	}
	return &WeeklyQuotaPlanInfo{
		Id:                      plan.Id,
		Title:                   plan.Title,
		Subtitle:                plan.Subtitle,
		TotalAmount:             plan.TotalAmount,
		DurationUnit:            plan.DurationUnit,
		DurationValue:           plan.DurationValue,
		CustomSeconds:           plan.CustomSeconds,
		QuotaResetPeriod:        plan.QuotaResetPeriod,
		QuotaResetCustomSeconds: plan.QuotaResetCustomSeconds,
	}
}

func weeklyQuotaAwardedFromPlan(plan *SubscriptionPlan) int {
	if plan == nil || plan.TotalAmount <= 0 {
		return 0
	}
	maxInt := int64(^uint(0) >> 1)
	if plan.TotalAmount > maxInt {
		return int(maxInt)
	}
	return int(plan.TotalAmount)
}

func GetWeeklyQuotaStatus(userId int, now int64) (*WeeklyQuotaStatus, error) {
	setting := operation_setting.GetWeeklyQuotaSetting()
	if now <= 0 {
		now = time.Now().Unix()
	}
	user := &User{}
	if err := DB.Select("id", "created_at", "phone_number").Where("id = ?", userId).First(user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWeeklyQuotaUserNotFound
		}
		return nil, err
	}

	periodDays := setting.PeriodDays
	windowStart, windowEnd := getWeeklyQuotaWindow(user.CreatedAt, now, periodDays)
	status := &WeeklyQuotaStatus{
		Enabled:         false,
		PlanId:          setting.PlanId,
		PeriodDays:      periodDays,
		Status:          WeeklyQuotaStatusClaimable,
		WindowStartedAt: windowStart,
		WindowEndsAt:    windowEnd,
	}
	var plan *SubscriptionPlan
	if setting.PlanId > 0 {
		loadedPlan, err := GetSubscriptionPlanById(setting.PlanId)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		plan = loadedPlan
		status.Plan = weeklyQuotaPlanInfo(plan)
		status.Amount = weeklyQuotaAwardedFromPlan(plan)
	}
	status.Enabled = setting.Enabled && setting.PlanId > 0 && setting.PeriodDays > 0 && plan != nil

	var totalQuota int64
	if err := DB.Model(&WeeklyQuotaClaim{}).Where("user_id = ?", userId).Count(&status.TotalClaimCount).Error; err != nil {
		return nil, err
	}
	if err := DB.Model(&WeeklyQuotaClaim{}).Where("user_id = ?", userId).Select("COALESCE(SUM(quota_awarded), 0)").Scan(&totalQuota).Error; err != nil {
		return nil, err
	}
	status.TotalClaimQuota = totalQuota

	if !status.Enabled {
		status.Status = WeeklyQuotaStatusDisabled
		return status, nil
	}

	if user.PhoneNumber == "" {
		status.Status = WeeklyQuotaStatusPhoneRequired
		return status, nil
	}

	claim := &WeeklyQuotaClaim{}
	err := DB.Where("user_id = ? AND window_start = ?", userId, windowStart).First(claim).Error
	if err == nil {
		status.Status = WeeklyQuotaStatusClaimed
		status.ClaimedAt = claim.CreatedAt
		status.NextClaimAt = windowEnd
		return status, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if plan != nil {
		active, until, err := userHasActiveSubscriptionForPlanTx(DB, userId, plan.Id, now)
		if err != nil {
			return nil, err
		}
		if active {
			status.Status = WeeklyQuotaStatusActive
			status.NextClaimAt = until
			return status, nil
		}
	}

	return status, nil
}

func ClaimWeeklyQuota(userId int, now int64) (*WeeklyQuotaClaim, error) {
	status, err := GetWeeklyQuotaStatus(userId, now)
	if err != nil {
		return nil, err
	}
	if !status.Enabled || status.PlanId <= 0 {
		return nil, ErrWeeklyQuotaDisabled
	}
	switch status.Status {
	case WeeklyQuotaStatusPhoneRequired:
		return nil, ErrWeeklyQuotaPhoneRequired
	case WeeklyQuotaStatusClaimed:
		return nil, ErrWeeklyQuotaAlreadyClaimed
	case WeeklyQuotaStatusActive:
		return nil, ErrWeeklyQuotaActiveSubscriptionExists
	}

	if now <= 0 {
		now = time.Now().Unix()
	}
	claim := &WeeklyQuotaClaim{
		UserId:       userId,
		WindowStart:  status.WindowStartedAt,
		WindowEnd:    status.WindowEndsAt,
		QuotaAwarded: status.Amount,
		PlanId:       status.PlanId,
		CreatedAt:    now,
	}

	return claimWeeklyQuotaWithTransaction(claim)
}

func claimWeeklyQuotaWithTransaction(claim *WeeklyQuotaClaim) (*WeeklyQuotaClaim, error) {
	err := DB.Transaction(func(tx *gorm.DB) error {
		plan, err := getSubscriptionPlanByIdTx(tx, claim.PlanId)
		if err != nil {
			return err
		}
		active, _, err := userHasActiveSubscriptionForPlanTx(tx, claim.UserId, claim.PlanId, claim.CreatedAt)
		if err != nil {
			return err
		}
		if active {
			return ErrWeeklyQuotaActiveSubscriptionExists
		}
		sub, err := CreateUserSubscriptionFromPlanTx(tx, claim.UserId, plan, SubscriptionSourceGiftClaim)
		if err != nil {
			return err
		}
		claim.UserSubscriptionId = sub.Id
		if err := tx.Create(claim).Error; err != nil {
			return ErrWeeklyQuotaAlreadyClaimed
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return claim, nil
}

func claimWeeklyQuotaWithoutTransaction(claim *WeeklyQuotaClaim) (*WeeklyQuotaClaim, error) {
	return claimWeeklyQuotaWithTransaction(claim)
}
