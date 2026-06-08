package model

import (
	"errors"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"gorm.io/gorm"
)

const (
	WeeklyQuotaStatusDisabled      = "disabled"
	WeeklyQuotaStatusClaimable     = "claimable"
	WeeklyQuotaStatusClaimed       = "claimed"
	WeeklyQuotaStatusPhoneRequired = "phone_required"

	weeklyQuotaWindowSeconds = int64(7 * 24 * 60 * 60)
)

var (
	ErrWeeklyQuotaDisabled       = errors.New("周额度领取未启用")
	ErrWeeklyQuotaAlreadyClaimed = errors.New("本周额度已领取")
	ErrWeeklyQuotaPhoneRequired  = errors.New("领取周额度前需要先绑定手机号")
	ErrWeeklyQuotaUserNotFound   = errors.New("用户不存在")
)

type WeeklyQuotaClaim struct {
	Id           int   `json:"id" gorm:"primaryKey;autoIncrement"`
	UserId       int   `json:"user_id" gorm:"not null;uniqueIndex:idx_user_weekly_quota_window"`
	WindowStart  int64 `json:"window_start" gorm:"bigint;not null;uniqueIndex:idx_user_weekly_quota_window"`
	WindowEnd    int64 `json:"window_end" gorm:"bigint;not null;index"`
	QuotaAwarded int   `json:"quota_awarded" gorm:"not null"`
	CreatedAt    int64 `json:"created_at" gorm:"bigint"`
}

func (WeeklyQuotaClaim) TableName() string {
	return "weekly_quota_claims"
}

type WeeklyQuotaStatus struct {
	Enabled         bool   `json:"enabled"`
	Amount          int    `json:"amount"`
	Status          string `json:"status"`
	WindowStartedAt int64  `json:"window_started_at"`
	WindowEndsAt    int64  `json:"window_ends_at"`
	ClaimedAt       int64  `json:"claimed_at"`
	NextClaimAt     int64  `json:"next_claim_at"`
	TotalClaimCount int64  `json:"total_claim_count"`
	TotalClaimQuota int64  `json:"total_claim_quota"`
}

func getWeeklyQuotaWindow(createdAt int64, now int64) (int64, int64) {
	if now <= 0 {
		now = time.Now().Unix()
	}
	if createdAt < 0 {
		createdAt = 0
	}
	if now < createdAt {
		return createdAt, createdAt + weeklyQuotaWindowSeconds
	}
	windowIndex := (now - createdAt) / weeklyQuotaWindowSeconds
	windowStart := createdAt + windowIndex*weeklyQuotaWindowSeconds
	return windowStart, windowStart + weeklyQuotaWindowSeconds
}

func GetWeeklyQuotaStatus(userId int, now int64) (*WeeklyQuotaStatus, error) {
	setting := operation_setting.GetWeeklyQuotaSetting()
	user := &User{}
	if err := DB.Select("id", "created_at", "phone_number").Where("id = ?", userId).First(user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWeeklyQuotaUserNotFound
		}
		return nil, err
	}

	windowStart, windowEnd := getWeeklyQuotaWindow(user.CreatedAt, now)
	status := &WeeklyQuotaStatus{
		Enabled:         setting.Enabled && setting.Amount > 0,
		Amount:          setting.Amount,
		Status:          WeeklyQuotaStatusClaimable,
		WindowStartedAt: windowStart,
		WindowEndsAt:    windowEnd,
	}

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

	if common.PhoneVerificationEnabled && user.PhoneNumber == "" {
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

	return status, nil
}

func ClaimWeeklyQuota(userId int, now int64) (*WeeklyQuotaClaim, error) {
	status, err := GetWeeklyQuotaStatus(userId, now)
	if err != nil {
		return nil, err
	}
	if !status.Enabled || status.Amount <= 0 {
		return nil, ErrWeeklyQuotaDisabled
	}
	switch status.Status {
	case WeeklyQuotaStatusPhoneRequired:
		return nil, ErrWeeklyQuotaPhoneRequired
	case WeeklyQuotaStatusClaimed:
		return nil, ErrWeeklyQuotaAlreadyClaimed
	}

	if now <= 0 {
		now = time.Now().Unix()
	}
	claim := &WeeklyQuotaClaim{
		UserId:       userId,
		WindowStart:  status.WindowStartedAt,
		WindowEnd:    status.WindowEndsAt,
		QuotaAwarded: status.Amount,
		CreatedAt:    now,
	}

	if common.UsingSQLite {
		return claimWeeklyQuotaWithoutTransaction(claim)
	}
	return claimWeeklyQuotaWithTransaction(claim)
}

func claimWeeklyQuotaWithTransaction(claim *WeeklyQuotaClaim) (*WeeklyQuotaClaim, error) {
	err := DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(claim).Error; err != nil {
			return ErrWeeklyQuotaAlreadyClaimed
		}
		if err := tx.Model(&User{}).Where("id = ?", claim.UserId).
			Update("quota", gorm.Expr("quota + ?", claim.QuotaAwarded)).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	go func() {
		_ = cacheIncrUserQuota(claim.UserId, int64(claim.QuotaAwarded))
	}()
	return claim, nil
}

func claimWeeklyQuotaWithoutTransaction(claim *WeeklyQuotaClaim) (*WeeklyQuotaClaim, error) {
	if err := DB.Create(claim).Error; err != nil {
		return nil, ErrWeeklyQuotaAlreadyClaimed
	}
	if err := IncreaseUserQuota(claim.UserId, claim.QuotaAwarded, true); err != nil {
		DB.Delete(claim)
		return nil, err
	}
	return claim, nil
}
