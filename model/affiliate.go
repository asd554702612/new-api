package model

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const (
	AffiliateLedgerActionAccrue          = "accrue"
	AffiliateLedgerActionTransfer        = "transfer"
	AffiliateLedgerActionSignupBonus     = "signup_bonus"
	AffiliateLedgerActionWithdrawRequest = "withdraw_request"
	AffiliateLedgerActionWithdrawPaid    = "withdraw_paid"
	AffiliateLedgerActionWithdrawReject  = "withdraw_reject"
	AffiliateLedgerActionWithdrawFail    = "withdraw_fail"

	AffiliateSourceOrderTypeTopUp        = "topup"
	AffiliateSourceOrderTypeSubscription = "subscription"
	AffiliateSourceOrderTypeManual       = "manual"

	AffiliateWithdrawalStatusPendingReview = "pending_review"
	AffiliateWithdrawalStatusApproved      = "approved"
	AffiliateWithdrawalStatusPaid          = "paid"
	AffiliateWithdrawalStatusRejected      = "rejected"
	AffiliateWithdrawalStatusFailed        = "failed"
	AffiliateWithdrawalStatusCancelled     = "cancelled"

	AffiliateWithdrawalPayoutMethodWechatManual = "wechat_manual"

	AffiliateIdentityTypeInviter = "inviter"
	AffiliateIdentityTypeInvitee = "invitee"

	AffiliateIdentityStatusActive  = "active"
	AffiliateIdentityStatusRevoked = "revoked"
)

type AffiliateLedger struct {
	Id                 int     `json:"id"`
	UserId             int     `json:"user_id" gorm:"index"`
	RelatedUserId      int     `json:"related_user_id" gorm:"index"`
	Action             string  `json:"action" gorm:"type:varchar(32);index"`
	Quota              int     `json:"quota" gorm:"type:int;default:0"`
	BalanceAfter       int     `json:"balance_after" gorm:"type:int;default:0"`
	HistoryAfter       int     `json:"history_after" gorm:"type:int;default:0"`
	DedupKey           *string `json:"dedup_key,omitempty" gorm:"type:varchar(128);uniqueIndex"`
	SourceOrderTradeNo string  `json:"source_order_trade_no" gorm:"type:varchar(255);index"`
	SourceOrderType    string  `json:"source_order_type" gorm:"type:varchar(64);index"`
	PaymentMethod      string  `json:"payment_method" gorm:"type:varchar(64)"`
	Remark             string  `json:"remark" gorm:"type:varchar(255)"`
	CreatedAt          int64   `json:"created_at" gorm:"autoCreateTime;index"`
}

type AffiliateWithdrawal struct {
	Id                int    `json:"id"`
	UserId            int    `json:"user_id" gorm:"index"`
	Quota             int    `json:"quota" gorm:"type:int;default:0"`
	Status            string `json:"status" gorm:"type:varchar(32);index"`
	PayoutMethod      string `json:"payout_method" gorm:"type:varchar(32)"`
	PayoutAccountNote string `json:"payout_account_note" gorm:"type:text"`
	AdminNote         string `json:"admin_note" gorm:"type:text"`
	PayoutChannel     string `json:"payout_channel" gorm:"type:varchar(64)"`
	PayoutTradeNo     string `json:"payout_trade_no" gorm:"type:varchar(128)"`
	RejectReason      string `json:"reject_reason" gorm:"type:text"`
	FailureReason     string `json:"failure_reason" gorm:"type:text"`
	ReviewedBy        int    `json:"reviewed_by" gorm:"index"`
	ReviewedAt        int64  `json:"reviewed_at"`
	PaidBy            int    `json:"paid_by" gorm:"index"`
	PaidAt            int64  `json:"paid_at"`
	CreatedAt         int64  `json:"created_at" gorm:"autoCreateTime;index"`
	UpdatedAt         int64  `json:"updated_at" gorm:"autoUpdateTime"`
}

type AffiliateIdentity struct {
	Id                    int     `json:"id"`
	UserId                int     `json:"user_id" gorm:"uniqueIndex:idx_affiliate_identity_user_type"`
	IdentityType          string  `json:"identity_type" gorm:"type:varchar(32);uniqueIndex:idx_affiliate_identity_user_type;index"`
	RateMultiplier        float64 `json:"rate_multiplier" gorm:"type:decimal(10,4);default:1"`
	SourceInviterId       int     `json:"source_inviter_id" gorm:"index"`
	GrantedAt             int64   `json:"granted_at" gorm:"index"`
	ExpiresAt             int64   `json:"expires_at" gorm:"index"`
	Status                string  `json:"status" gorm:"type:varchar(32);default:'active';index"`
	QualificationSnapshot string  `json:"qualification_snapshot" gorm:"type:text"`
	CreatedAt             int64   `json:"created_at" gorm:"autoCreateTime;index"`
	UpdatedAt             int64   `json:"updated_at" gorm:"autoUpdateTime"`
}

type AffiliateSignupFingerprint struct {
	Id             int    `json:"id"`
	UserId         int    `json:"user_id" gorm:"uniqueIndex"`
	CompositeHash  string `json:"composite_hash" gorm:"type:varchar(128);index"`
	CanvasHash     string `json:"canvas_hash" gorm:"type:varchar(128)"`
	WebGLHash      string `json:"webgl_hash" gorm:"type:varchar(128)"`
	Components     string `json:"components" gorm:"type:text"`
	DuplicateCount int    `json:"duplicate_count" gorm:"type:int;default:0"`
	RiskFlagged    bool   `json:"risk_flagged" gorm:"default:false;index"`
	RiskReason     string `json:"risk_reason" gorm:"type:varchar(128)"`
	CreatedAt      int64  `json:"created_at" gorm:"autoCreateTime;index"`
	UpdatedAt      int64  `json:"updated_at" gorm:"autoUpdateTime"`
}

type AffiliateIdentityConfig struct {
	InviterRateMultiplier         float64  `json:"inviter_rate_multiplier"`
	InviteeRateMultiplier         float64  `json:"invitee_rate_multiplier"`
	DurationHours                 int      `json:"duration_hours"`
	QualifiedInviteeCount         int      `json:"qualified_invitee_count"`
	QualifiedPayAmount            float64  `json:"qualified_pay_amount"`
	EligibleOrderTypes            []string `json:"eligible_order_types"`
	FingerprintEnforcementEnabled bool     `json:"fingerprint_enforcement_enabled"`
	MaxAccountsPerFingerprintHash int      `json:"max_accounts_per_fingerprint_hash"`
}

type AffiliateSignupFingerprintInput struct {
	CompositeHash string            `json:"composite_hash"`
	CanvasHash    string            `json:"canvas_hash"`
	WebGLHash     string            `json:"webgl_hash"`
	Components    map[string]string `json:"components,omitempty"`
}

type AffiliateInviteRelationChange struct {
	InviterId         int `json:"inviter_id"`
	InviteeId         int `json:"invitee_id"`
	PreviousInviterId int `json:"previous_inviter_id"`
}

type AffiliateAdminUserRecord struct {
	UserId               int      `json:"user_id"`
	Username             string   `json:"username"`
	Email                string   `json:"email"`
	AffCode              string   `json:"aff_code"`
	AffCodeCustom        bool     `json:"aff_code_custom"`
	AffRebateRatePercent *float64 `json:"aff_rebate_rate_percent"`
	AffCount             int      `json:"aff_count"`
	AffQuota             int      `json:"aff_quota"`
	AffHistoryQuota      int      `json:"aff_history_quota"`
	InviterId            int      `json:"inviter_id"`
	CreatedAt            int64    `json:"created_at"`
}

type AffiliateUserOverview struct {
	AffiliateAdminUserRecord
	InviterUsername string             `json:"inviter_username"`
	InviterEmail    string             `json:"inviter_email"`
	ActiveIdentity  *AffiliateIdentity `json:"active_identity,omitempty" gorm:"-"`
}

type AffiliateFingerprintRecord struct {
	Id             int    `json:"id"`
	UserId         int    `json:"user_id"`
	Username       string `json:"username"`
	Email          string `json:"email"`
	CompositeHash  string `json:"composite_hash"`
	CanvasHash     string `json:"canvas_hash"`
	WebGLHash      string `json:"webgl_hash"`
	DuplicateCount int    `json:"duplicate_count"`
	RiskFlagged    bool   `json:"risk_flagged"`
	RiskReason     string `json:"risk_reason"`
	CreatedAt      int64  `json:"created_at"`
	UpdatedAt      int64  `json:"updated_at"`
}

type AffiliateInviteRecord struct {
	UserId          int    `json:"user_id"`
	Username        string `json:"username"`
	Email           string `json:"email"`
	AffCode         string `json:"aff_code"`
	InviterId       int    `json:"inviter_id"`
	InviterUsername string `json:"inviter_username"`
	InviterEmail    string `json:"inviter_email"`
	CreatedAt       int64  `json:"created_at"`
}

type AffiliateLedgerRecord struct {
	Id                 int    `json:"id"`
	UserId             int    `json:"user_id"`
	Username           string `json:"username"`
	Email              string `json:"email"`
	RelatedUserId      int    `json:"related_user_id"`
	RelatedUsername    string `json:"related_username"`
	RelatedEmail       string `json:"related_email"`
	Action             string `json:"action"`
	Quota              int    `json:"quota"`
	BalanceAfter       int    `json:"balance_after"`
	HistoryAfter       int    `json:"history_after"`
	SourceOrderTradeNo string `json:"source_order_trade_no"`
	SourceOrderType    string `json:"source_order_type"`
	PaymentMethod      string `json:"payment_method"`
	Remark             string `json:"remark"`
	CreatedAt          int64  `json:"created_at"`
}

type AffiliateWithdrawalRecord struct {
	Id                int    `json:"id"`
	UserId            int    `json:"user_id"`
	Username          string `json:"username"`
	Email             string `json:"email"`
	Quota             int    `json:"quota"`
	Status            string `json:"status"`
	PayoutMethod      string `json:"payout_method"`
	PayoutAccountNote string `json:"payout_account_note"`
	AdminNote         string `json:"admin_note"`
	PayoutChannel     string `json:"payout_channel"`
	PayoutTradeNo     string `json:"payout_trade_no"`
	RejectReason      string `json:"reject_reason"`
	FailureReason     string `json:"failure_reason"`
	ReviewedBy        int    `json:"reviewed_by"`
	ReviewedAt        int64  `json:"reviewed_at"`
	PaidBy            int    `json:"paid_by"`
	PaidAt            int64  `json:"paid_at"`
	CreatedAt         int64  `json:"created_at"`
	UpdatedAt         int64  `json:"updated_at"`
}

type AffiliateRecordFilter struct {
	Search    string
	Status    string
	StartTime int64
	EndTime   int64
}

func DefaultAffiliateIdentityConfig() AffiliateIdentityConfig {
	return AffiliateIdentityConfig{
		InviterRateMultiplier:         1.5,
		InviteeRateMultiplier:         1.4,
		DurationHours:                 24 * 30,
		QualifiedInviteeCount:         0,
		QualifiedPayAmount:            50,
		EligibleOrderTypes:            []string{AffiliateSourceOrderTypeTopUp, AffiliateSourceOrderTypeSubscription},
		FingerprintEnforcementEnabled: true,
		MaxAccountsPerFingerprintHash: 3,
	}
}

func DefaultAffiliateIdentityConfigJSON() string {
	data, err := common.Marshal(DefaultAffiliateIdentityConfig())
	if err != nil {
		return "{}"
	}
	return string(data)
}

func GetAffiliateIdentityConfig() AffiliateIdentityConfig {
	cfg := DefaultAffiliateIdentityConfig()
	raw := strings.TrimSpace(common.AffiliateIdentityConfig)
	if raw == "" {
		return cfg
	}
	var parsed AffiliateIdentityConfig
	if err := common.Unmarshal([]byte(raw), &parsed); err != nil {
		return cfg
	}
	if parsed.InviterRateMultiplier > 0 && !math.IsNaN(parsed.InviterRateMultiplier) && !math.IsInf(parsed.InviterRateMultiplier, 0) {
		cfg.InviterRateMultiplier = parsed.InviterRateMultiplier
	}
	if parsed.InviteeRateMultiplier > 0 && !math.IsNaN(parsed.InviteeRateMultiplier) && !math.IsInf(parsed.InviteeRateMultiplier, 0) {
		cfg.InviteeRateMultiplier = parsed.InviteeRateMultiplier
	}
	if parsed.DurationHours > 0 {
		cfg.DurationHours = parsed.DurationHours
	}
	if parsed.QualifiedInviteeCount >= 0 {
		cfg.QualifiedInviteeCount = parsed.QualifiedInviteeCount
	}
	if parsed.QualifiedPayAmount >= 0 && !math.IsNaN(parsed.QualifiedPayAmount) && !math.IsInf(parsed.QualifiedPayAmount, 0) {
		cfg.QualifiedPayAmount = parsed.QualifiedPayAmount
	}
	if len(parsed.EligibleOrderTypes) > 0 {
		cfg.EligibleOrderTypes = normalizeAffiliateEligibleOrderTypes(parsed.EligibleOrderTypes)
	}
	cfg.FingerprintEnforcementEnabled = parsed.FingerprintEnforcementEnabled
	if parsed.MaxAccountsPerFingerprintHash > 0 {
		cfg.MaxAccountsPerFingerprintHash = parsed.MaxAccountsPerFingerprintHash
	}
	return cfg
}

func normalizeAffiliateEligibleOrderTypes(values []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if value == "balance" {
			value = AffiliateSourceOrderTypeTopUp
		}
		if value != AffiliateSourceOrderTypeTopUp && value != AffiliateSourceOrderTypeSubscription {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	if len(out) == 0 {
		return []string{AffiliateSourceOrderTypeTopUp, AffiliateSourceOrderTypeSubscription}
	}
	return out
}

func AdminSetAffiliateUserSettings(userID int, affCode string, ratePercent *float64) error {
	if err := validateAffiliateRebateRate(ratePercent); err != nil {
		return err
	}
	updates := map[string]any{}
	if strings.TrimSpace(affCode) != "" {
		code := strings.ToUpper(strings.TrimSpace(affCode))
		if len(code) > 32 {
			return errors.New("邀请码长度不能超过32")
		}
		var existing User
		err := DB.Select("id").Where("aff_code = ? AND id <> ?", code, userID).First(&existing).Error
		if err == nil {
			return errors.New("邀请码已被使用")
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		updates["aff_code"] = code
		updates["aff_code_custom"] = true
	}
	updates["aff_rebate_rate_percent"] = ratePercent
	if len(updates) == 0 {
		return nil
	}
	return DB.Model(&User{}).Where("id = ?", userID).Updates(updates).Error
}

func AdminBatchSetAffiliateRebateRate(userIDs []int, ratePercent *float64) error {
	if err := validateAffiliateRebateRate(ratePercent); err != nil {
		return err
	}
	cleaned := make([]int, 0, len(userIDs))
	for _, userID := range userIDs {
		if userID > 0 {
			cleaned = append(cleaned, userID)
		}
	}
	if len(cleaned) == 0 {
		return nil
	}
	return DB.Model(&User{}).Where("id IN ?", cleaned).Update("aff_rebate_rate_percent", ratePercent).Error
}

func validateAffiliateRebateRate(ratePercent *float64) error {
	if ratePercent == nil {
		return nil
	}
	value := *ratePercent
	if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 || value > 100 {
		return errors.New("返利比例必须在0到100之间")
	}
	return nil
}

func AdminSetInviteRelation(inviterID int, inviteeID int, overwrite bool) (*AffiliateInviteRelationChange, error) {
	if inviterID <= 0 || inviteeID <= 0 {
		return nil, errors.New("invalid user")
	}
	if inviterID == inviteeID {
		return nil, errors.New("邀请人和被邀请人不能相同")
	}
	var change *AffiliateInviteRelationChange
	err := DB.Transaction(func(tx *gorm.DB) error {
		var inviter User
		if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("id = ?", inviterID).First(&inviter).Error; err != nil {
			return err
		}
		var invitee User
		if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("id = ?", inviteeID).First(&invitee).Error; err != nil {
			return err
		}
		previous := invitee.InviterId
		if previous > 0 && previous != inviterID && !overwrite {
			return errors.New("被邀请人已有邀请关系")
		}
		if previous == inviterID {
			change = &AffiliateInviteRelationChange{InviterId: inviterID, InviteeId: inviteeID, PreviousInviterId: previous}
			return nil
		}
		if previous > 0 {
			if err := tx.Model(&User{}).Where("id = ? AND aff_count > 0", previous).Update("aff_count", gorm.Expr("aff_count - ?", 1)).Error; err != nil {
				return err
			}
		}
		invitee.InviterId = inviterID
		if err := tx.Save(&invitee).Error; err != nil {
			return err
		}
		if err := tx.Model(&User{}).Where("id = ?", inviterID).Update("aff_count", gorm.Expr("aff_count + ?", 1)).Error; err != nil {
			return err
		}
		change = &AffiliateInviteRelationChange{InviterId: inviterID, InviteeId: inviteeID, PreviousInviterId: previous}
		return nil
	})
	return change, err
}

func GetAffiliateAdminUsers(pageInfo *common.PageInfo, filter AffiliateRecordFilter) ([]*AffiliateAdminUserRecord, int64, error) {
	var records []*AffiliateAdminUserRecord
	var total int64
	if pageInfo == nil {
		pageInfo = &common.PageInfo{}
	}
	query := DB.Table("users AS u").
		Where("u.aff_code_custom = ? OR u.aff_rebate_rate_percent IS NOT NULL", true)
	query = applyAffiliateSingleUserSearch(query, filter.Search, "u")
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Select("u.id AS user_id, u.username, u.email, u.aff_code, u.aff_code_custom, u.aff_rebate_rate_percent, u.aff_count, u.aff_quota, u.aff_history AS aff_history_quota, u.inviter_id, u.created_at").
		Order("u.id desc").
		Limit(pageInfo.GetPageSize()).
		Offset(pageInfo.GetStartIdx()).
		Scan(&records).Error; err != nil {
		return nil, 0, err
	}
	return records, total, nil
}

func LookupAffiliateUsers(keyword string) ([]*AffiliateAdminUserRecord, error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return []*AffiliateAdminUserRecord{}, nil
	}
	var records []*AffiliateAdminUserRecord
	query := DB.Table("users AS u")
	query = applyAffiliateSingleUserSearch(query, keyword, "u")
	err := query.Select("u.id AS user_id, u.username, u.email, u.aff_code, u.aff_code_custom, u.aff_rebate_rate_percent, u.aff_count, u.aff_quota, u.aff_history AS aff_history_quota, u.inviter_id, u.created_at").
		Order("u.id desc").
		Limit(20).
		Scan(&records).Error
	return records, err
}

func GetAffiliateUserOverview(userID int) (*AffiliateUserOverview, error) {
	var overview AffiliateUserOverview
	err := DB.Table("users AS u").
		Joins("LEFT JOIN users AS inviter ON inviter.id = u.inviter_id").
		Where("u.id = ?", userID).
		Select("u.id AS user_id, u.username, u.email, u.aff_code, u.aff_code_custom, u.aff_rebate_rate_percent, u.aff_count, u.aff_quota, u.aff_history AS aff_history_quota, u.inviter_id, u.created_at, inviter.username AS inviter_username, inviter.email AS inviter_email").
		Scan(&overview).Error
	if err != nil {
		return nil, err
	}
	if overview.UserId == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	overview.ActiveIdentity = GetActiveAffiliateIdentity(userID)
	return &overview, nil
}

func ClearAffiliateUserSettings(userID int) error {
	code := common.GetRandomString(4)
	return DB.Model(&User{}).Where("id = ?", userID).Updates(map[string]any{
		"aff_rebate_rate_percent": nil,
		"aff_code":                code,
		"aff_code_custom":         false,
	}).Error
}

func GetAffiliateFingerprintRecords(pageInfo *common.PageInfo, filter AffiliateRecordFilter) ([]*AffiliateFingerprintRecord, int64, error) {
	var records []*AffiliateFingerprintRecord
	var total int64
	if pageInfo == nil {
		pageInfo = &common.PageInfo{}
	}
	query := DB.Table("affiliate_signup_fingerprints AS f").Joins("LEFT JOIN users AS u ON u.id = f.user_id")
	query = applyAffiliateSingleUserSearch(query, filter.Search, "u")
	if strings.TrimSpace(filter.Status) == "risk" {
		query = query.Where("f.risk_flagged = ?", true)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Select("f.*, u.username, u.email").
		Order("f.id desc").
		Limit(pageInfo.GetPageSize()).
		Offset(pageInfo.GetStartIdx()).
		Scan(&records).Error; err != nil {
		return nil, 0, err
	}
	return records, total, nil
}

func SaveAffiliateIdentityConfig(enabled bool, cfg AffiliateIdentityConfig) error {
	normalized := cfg
	if normalized.InviterRateMultiplier <= 0 {
		normalized.InviterRateMultiplier = DefaultAffiliateIdentityConfig().InviterRateMultiplier
	}
	if normalized.InviteeRateMultiplier <= 0 {
		normalized.InviteeRateMultiplier = DefaultAffiliateIdentityConfig().InviteeRateMultiplier
	}
	if normalized.DurationHours <= 0 {
		normalized.DurationHours = DefaultAffiliateIdentityConfig().DurationHours
	}
	if normalized.MaxAccountsPerFingerprintHash <= 0 {
		normalized.MaxAccountsPerFingerprintHash = 1
	}
	normalized.EligibleOrderTypes = normalizeAffiliateEligibleOrderTypes(normalized.EligibleOrderTypes)
	data, err := common.Marshal(normalized)
	if err != nil {
		return err
	}
	common.AffiliateIdentityEnabled = enabled
	common.AffiliateIdentityConfig = string(data)
	if err := UpdateOption("AffiliateIdentityEnabled", strconv.FormatBool(enabled)); err != nil {
		return err
	}
	return UpdateOption("AffiliateIdentityConfig", common.AffiliateIdentityConfig)
}

func CreateAffiliateWithdrawal(userID int, quota int, payoutMethod string, payoutAccountNote string) (*AffiliateWithdrawal, error) {
	if !common.AffiliateWithdrawEnabled {
		return nil, errors.New("邀请返利提现未启用")
	}
	if !operation_setting.IsPaymentComplianceConfirmed() {
		return nil, errors.New("支付合规确认后才可申请提现")
	}
	if userID <= 0 {
		return nil, errors.New("invalid user")
	}
	if quota < common.AffiliateWithdrawMinQuota {
		return nil, fmt.Errorf("提现金额最低为%d", common.AffiliateWithdrawMinQuota)
	}
	payoutMethod = strings.TrimSpace(payoutMethod)
	payoutAccountNote = strings.TrimSpace(payoutAccountNote)
	if payoutMethod != AffiliateWithdrawalPayoutMethodWechatManual {
		return nil, errors.New("不支持的提现方式")
	}
	if payoutAccountNote == "" {
		return nil, errors.New("请填写收款说明")
	}

	var withdrawal *AffiliateWithdrawal
	err := DB.Transaction(func(tx *gorm.DB) error {
		if common.AffiliateWithdrawDailyLimit > 0 {
			cutoff := time.Now().Add(-24 * time.Hour).Unix()
			var count int64
			if err := tx.Model(&AffiliateWithdrawal{}).
				Where("user_id = ? AND created_at >= ?", userID, cutoff).
				Count(&count).Error; err != nil {
				return err
			}
			if count >= int64(common.AffiliateWithdrawDailyLimit) {
				return errors.New("今日提现申请次数已达上限")
			}
		}

		var user User
		if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("id = ?", userID).First(&user).Error; err != nil {
			return err
		}
		if user.AffQuota < quota {
			return errors.New("邀请额度不足！")
		}
		user.AffQuota -= quota
		if err := tx.Save(&user).Error; err != nil {
			return err
		}

		record := &AffiliateWithdrawal{
			UserId:            userID,
			Quota:             quota,
			Status:            AffiliateWithdrawalStatusPendingReview,
			PayoutMethod:      payoutMethod,
			PayoutAccountNote: payoutAccountNote,
		}
		if err := tx.Create(record).Error; err != nil {
			return err
		}
		if err := CreateAffiliateWithdrawalLedgerTx(tx, userID, AffiliateLedgerActionWithdrawRequest, quota, user.AffQuota, user.AffHistoryQuota, record.Id, ""); err != nil {
			return err
		}
		withdrawal = record
		return nil
	})
	if err != nil {
		return nil, err
	}
	return withdrawal, nil
}

func ApproveAffiliateWithdrawal(id int, adminID int, note string) (*AffiliateWithdrawal, error) {
	var result *AffiliateWithdrawal
	err := DB.Transaction(func(tx *gorm.DB) error {
		withdrawal, err := getAffiliateWithdrawalForUpdate(tx, id)
		if err != nil {
			return err
		}
		if withdrawal.Status != AffiliateWithdrawalStatusPendingReview {
			return errors.New("提现申请状态不可审核通过")
		}
		now := time.Now().Unix()
		withdrawal.Status = AffiliateWithdrawalStatusApproved
		withdrawal.ReviewedBy = adminID
		withdrawal.ReviewedAt = now
		withdrawal.AdminNote = strings.TrimSpace(note)
		if err := tx.Save(withdrawal).Error; err != nil {
			return err
		}
		result = withdrawal
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func RejectAffiliateWithdrawal(id int, adminID int, reason string) (*AffiliateWithdrawal, error) {
	if strings.TrimSpace(reason) == "" {
		return nil, errors.New("请填写拒绝原因")
	}
	return refundAffiliateWithdrawal(id, adminID, AffiliateWithdrawalStatusRejected, reason)
}

func FailAffiliateWithdrawal(id int, adminID int, reason string) (*AffiliateWithdrawal, error) {
	if strings.TrimSpace(reason) == "" {
		return nil, errors.New("请填写失败原因")
	}
	return refundAffiliateWithdrawal(id, adminID, AffiliateWithdrawalStatusFailed, reason)
}

func MarkAffiliateWithdrawalPaid(id int, adminID int, payoutChannel string, payoutTradeNo string, adminNote string) (*AffiliateWithdrawal, error) {
	var result *AffiliateWithdrawal
	payoutChannel = strings.TrimSpace(payoutChannel)
	payoutTradeNo = strings.TrimSpace(payoutTradeNo)
	adminNote = strings.TrimSpace(adminNote)
	if payoutChannel == "" {
		return nil, errors.New("请填写打款渠道")
	}
	err := DB.Transaction(func(tx *gorm.DB) error {
		withdrawal, err := getAffiliateWithdrawalForUpdate(tx, id)
		if err != nil {
			return err
		}
		if withdrawal.Status != AffiliateWithdrawalStatusApproved {
			return errors.New("提现申请状态不可标记为已打款")
		}
		now := time.Now().Unix()
		withdrawal.Status = AffiliateWithdrawalStatusPaid
		withdrawal.PaidBy = adminID
		withdrawal.PaidAt = now
		withdrawal.PayoutChannel = payoutChannel
		withdrawal.PayoutTradeNo = payoutTradeNo
		withdrawal.AdminNote = adminNote
		if err := tx.Save(withdrawal).Error; err != nil {
			return err
		}
		var user User
		if err := tx.Select("aff_quota", "aff_history").Where("id = ?", withdrawal.UserId).First(&user).Error; err != nil {
			return err
		}
		if err := CreateAffiliateWithdrawalLedgerTx(tx, withdrawal.UserId, AffiliateLedgerActionWithdrawPaid, withdrawal.Quota, user.AffQuota, user.AffHistoryQuota, withdrawal.Id, payoutTradeNo); err != nil {
			return err
		}
		result = withdrawal
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func refundAffiliateWithdrawal(id int, adminID int, targetStatus string, reason string) (*AffiliateWithdrawal, error) {
	var result *AffiliateWithdrawal
	err := DB.Transaction(func(tx *gorm.DB) error {
		withdrawal, err := getAffiliateWithdrawalForUpdate(tx, id)
		if err != nil {
			return err
		}
		if targetStatus == AffiliateWithdrawalStatusRejected && withdrawal.Status != AffiliateWithdrawalStatusPendingReview {
			return errors.New("提现申请状态不可拒绝")
		}
		if targetStatus == AffiliateWithdrawalStatusFailed && withdrawal.Status != AffiliateWithdrawalStatusApproved {
			return errors.New("提现申请状态不可标记失败")
		}

		var user User
		if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("id = ?", withdrawal.UserId).First(&user).Error; err != nil {
			return err
		}
		user.AffQuota += withdrawal.Quota
		if err := tx.Save(&user).Error; err != nil {
			return err
		}

		now := time.Now().Unix()
		withdrawal.Status = targetStatus
		withdrawal.ReviewedBy = adminID
		withdrawal.ReviewedAt = now
		if targetStatus == AffiliateWithdrawalStatusRejected {
			withdrawal.RejectReason = strings.TrimSpace(reason)
		} else {
			withdrawal.FailureReason = strings.TrimSpace(reason)
		}
		if err := tx.Save(withdrawal).Error; err != nil {
			return err
		}
		action := AffiliateLedgerActionWithdrawReject
		if targetStatus == AffiliateWithdrawalStatusFailed {
			action = AffiliateLedgerActionWithdrawFail
		}
		if err := CreateAffiliateWithdrawalLedgerTx(tx, withdrawal.UserId, action, withdrawal.Quota, user.AffQuota, user.AffHistoryQuota, withdrawal.Id, strings.TrimSpace(reason)); err != nil {
			return err
		}
		result = withdrawal
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func getAffiliateWithdrawalForUpdate(tx *gorm.DB, id int) (*AffiliateWithdrawal, error) {
	if tx == nil {
		return nil, errors.New("invalid transaction")
	}
	if id <= 0 {
		return nil, errors.New("invalid withdrawal")
	}
	var withdrawal AffiliateWithdrawal
	if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("id = ?", id).First(&withdrawal).Error; err != nil {
		return nil, err
	}
	return &withdrawal, nil
}

func ApplyAffiliateRebateTx(tx *gorm.DB, userID int, rebateBaseQuota int, sourceTradeNo string, sourceOrderType string, paymentMethod string) (int, error) {
	if tx == nil {
		return 0, errors.New("invalid transaction")
	}
	sourceTradeNo = strings.TrimSpace(sourceTradeNo)
	if !common.AffiliateEnabled || !operation_setting.IsPaymentComplianceConfirmed() {
		return 0, nil
	}
	if userID <= 0 || rebateBaseQuota <= 0 || sourceTradeNo == "" {
		return 0, nil
	}

	var existing int64
	if err := tx.Model(&AffiliateLedger{}).
		Where("action = ? AND source_order_trade_no = ? AND source_order_type = ?", AffiliateLedgerActionAccrue, sourceTradeNo, sourceOrderType).
		Count(&existing).Error; err != nil {
		return 0, err
	}
	if existing > 0 {
		return 0, nil
	}

	var invitee User
	if err := tx.Select("id", "inviter_id").Where("id = ?", userID).First(&invitee).Error; err != nil {
		return 0, err
	}
	if invitee.InviterId <= 0 {
		return 0, nil
	}

	var inviter User
	if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("id = ?", invitee.InviterId).First(&inviter).Error; err != nil {
		return 0, err
	}
	rate := common.AffiliateRebateRate
	if inviter.AffRebateRatePercent != nil {
		rate = *inviter.AffRebateRatePercent
	}
	if rate <= 0 {
		return 0, nil
	}
	rebate := int(decimal.NewFromInt(int64(rebateBaseQuota)).
		Mul(decimal.NewFromFloat(rate)).
		Div(decimal.NewFromInt(100)).
		IntPart())
	if rebate <= 0 {
		return 0, nil
	}

	inviter.AffQuota += rebate
	inviter.AffHistoryQuota += rebate
	if err := tx.Save(&inviter).Error; err != nil {
		return 0, err
	}

	ledger := AffiliateLedger{
		UserId:             inviter.Id,
		RelatedUserId:      invitee.Id,
		Action:             AffiliateLedgerActionAccrue,
		Quota:              rebate,
		BalanceAfter:       inviter.AffQuota,
		HistoryAfter:       inviter.AffHistoryQuota,
		SourceOrderTradeNo: sourceTradeNo,
		SourceOrderType:    sourceOrderType,
		PaymentMethod:      paymentMethod,
	}
	if err := tx.Create(&ledger).Error; err != nil {
		return 0, err
	}
	if err := refreshAffiliateIdentitiesForInviter(tx, inviter.Id); err != nil {
		return 0, err
	}
	return rebate, nil
}

func CreateAffiliateTransferLedgerTx(tx *gorm.DB, user *User, quota int) error {
	if tx == nil || user == nil {
		return errors.New("invalid affiliate transfer")
	}
	ledger := AffiliateLedger{
		UserId:       user.Id,
		Action:       AffiliateLedgerActionTransfer,
		Quota:        quota,
		BalanceAfter: user.AffQuota,
		HistoryAfter: user.AffHistoryQuota,
		Remark:       "aff_transfer",
	}
	return tx.Create(&ledger).Error
}

func CreateAffiliateWithdrawalLedgerTx(tx *gorm.DB, userID int, action string, quota int, balanceAfter int, historyAfter int, withdrawalID int, remark string) error {
	if tx == nil || userID <= 0 || withdrawalID <= 0 {
		return errors.New("invalid affiliate withdrawal ledger")
	}
	dedup := fmt.Sprintf("%s:%d", action, withdrawalID)
	existing := AffiliateLedger{}
	err := tx.Where("dedup_key = ?", dedup).First(&existing).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	ledger := AffiliateLedger{
		UserId:       userID,
		Action:       action,
		Quota:        quota,
		BalanceAfter: balanceAfter,
		HistoryAfter: historyAfter,
		DedupKey:     &dedup,
		Remark:       strings.TrimSpace(remark),
	}
	return tx.Create(&ledger).Error
}

func ApplyAffiliateSignupBonus(inviteeUserID int) (bool, error) {
	if !common.AffiliateEnabled || !common.AffiliateSignupRewardEnabled || common.AffiliateSignupRewardQuota <= 0 {
		return false, nil
	}
	if !operation_setting.IsPaymentComplianceConfirmed() {
		return false, nil
	}
	if inviteeUserID <= 0 {
		return false, nil
	}
	applied := false
	err := DB.Transaction(func(tx *gorm.DB) error {
		var invitee User
		if err := tx.Select("id", "inviter_id").Where("id = ?", inviteeUserID).First(&invitee).Error; err != nil {
			return err
		}
		if invitee.InviterId <= 0 {
			return nil
		}
		dedup := fmt.Sprintf("%s:%d", AffiliateLedgerActionSignupBonus, invitee.Id)
		var existing int64
		if err := tx.Model(&AffiliateLedger{}).Where("dedup_key = ?", dedup).Count(&existing).Error; err != nil {
			return err
		}
		if existing > 0 {
			return nil
		}
		var inviter User
		if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("id = ?", invitee.InviterId).First(&inviter).Error; err != nil {
			return err
		}
		inviter.AffQuota += common.AffiliateSignupRewardQuota
		inviter.AffHistoryQuota += common.AffiliateSignupRewardQuota
		if err := tx.Save(&inviter).Error; err != nil {
			return err
		}
		ledger := AffiliateLedger{
			UserId:        inviter.Id,
			RelatedUserId: invitee.Id,
			Action:        AffiliateLedgerActionSignupBonus,
			Quota:         common.AffiliateSignupRewardQuota,
			BalanceAfter:  inviter.AffQuota,
			HistoryAfter:  inviter.AffHistoryQuota,
			DedupKey:      &dedup,
			Remark:        "signup_bonus",
		}
		if err := tx.Create(&ledger).Error; err != nil {
			return err
		}
		applied = true
		return nil
	})
	return applied, err
}

func RecordAffiliateSignupFingerprint(userID int, input AffiliateSignupFingerprintInput) error {
	if userID <= 0 {
		return nil
	}
	if !DB.Migrator().HasTable(&AffiliateSignupFingerprint{}) {
		return nil
	}
	cfg := GetAffiliateIdentityConfig()
	input.CompositeHash = strings.TrimSpace(input.CompositeHash)
	input.CanvasHash = strings.TrimSpace(input.CanvasHash)
	input.WebGLHash = strings.TrimSpace(input.WebGLHash)
	components := "{}"
	if input.Components != nil {
		data, err := common.Marshal(input.Components)
		if err != nil {
			return err
		}
		components = string(data)
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		var duplicateCount int64
		if input.CompositeHash != "" {
			if err := tx.Model(&AffiliateSignupFingerprint{}).Where("composite_hash = ? AND user_id <> ?", input.CompositeHash, userID).Count(&duplicateCount).Error; err != nil {
				return err
			}
		}
		riskFlagged := false
		riskReason := ""
		if cfg.FingerprintEnforcementEnabled && input.CompositeHash != "" && int(duplicateCount) >= cfg.MaxAccountsPerFingerprintHash {
			riskFlagged = true
			riskReason = "duplicate_fingerprint"
		}
		var existing AffiliateSignupFingerprint
		err := tx.Where("user_id = ?", userID).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tx.Create(&AffiliateSignupFingerprint{
				UserId:         userID,
				CompositeHash:  input.CompositeHash,
				CanvasHash:     input.CanvasHash,
				WebGLHash:      input.WebGLHash,
				Components:     components,
				DuplicateCount: int(duplicateCount),
				RiskFlagged:    riskFlagged,
				RiskReason:     riskReason,
			}).Error
		}
		if err != nil {
			return err
		}
		existing.CompositeHash = input.CompositeHash
		existing.CanvasHash = input.CanvasHash
		existing.WebGLHash = input.WebGLHash
		existing.Components = components
		existing.DuplicateCount = int(duplicateCount)
		existing.RiskFlagged = riskFlagged
		existing.RiskReason = riskReason
		return tx.Save(&existing).Error
	})
}

func RefreshAffiliateIdentitiesForInviter(inviterID int) error {
	return refreshAffiliateIdentitiesForInviter(DB, inviterID)
}

func refreshAffiliateIdentitiesForInviter(db *gorm.DB, inviterID int) error {
	if inviterID <= 0 {
		return nil
	}
	if !common.AffiliateIdentityEnabled {
		return revokeAffiliateIdentitiesForInviter(db, inviterID)
	}
	cfg := GetAffiliateIdentityConfig()
	var invitees []User
	if err := db.Select("id").Where("inviter_id = ?", inviterID).Find(&invitees).Error; err != nil {
		return err
	}
	eligibleTypes := normalizeAffiliateEligibleOrderTypes(cfg.EligibleOrderTypes)
	eligibleSet := map[string]struct{}{}
	for _, value := range eligibleTypes {
		eligibleSet[value] = struct{}{}
	}
	qualified := make([]int, 0, len(invitees))
	totalAmount := 0.0
	for _, invitee := range invitees {
		if affiliateInviteeRiskFlagged(db, invitee.Id) {
			continue
		}
		var quota int64
		query := db.Model(&AffiliateLedger{}).
			Where("user_id = ? AND related_user_id = ? AND action IN ?", inviterID, invitee.Id, []string{AffiliateLedgerActionAccrue, AffiliateLedgerActionSignupBonus})
		if len(eligibleSet) > 0 {
			query = query.Where("(source_order_type IN ? OR source_order_type = '')", eligibleTypes)
		}
		if err := query.Select("COALESCE(SUM(quota), 0)").Scan(&quota).Error; err != nil {
			return err
		}
		amount := float64(quota) / common.QuotaPerUnit
		if amount+1e-9 >= cfg.QualifiedPayAmount {
			qualified = append(qualified, invitee.Id)
			totalAmount += amount
		}
	}
	if len(qualified) < cfg.QualifiedInviteeCount || totalAmount+1e-9 < cfg.QualifiedPayAmount {
		return revokeAffiliateIdentitiesForInviter(db, inviterID)
	}
	now := time.Now().Unix()
	expiresAt := time.Now().Add(time.Duration(cfg.DurationHours) * time.Hour).Unix()
	snapshot := map[string]any{
		"qualified_invitee_count": len(qualified),
		"qualified_pay_amount":    totalAmount,
	}
	if err := upsertAffiliateIdentity(db, inviterID, AffiliateIdentityTypeInviter, cfg.InviterRateMultiplier, 0, now, expiresAt, snapshot); err != nil {
		return err
	}
	for _, inviteeID := range qualified {
		if err := upsertAffiliateIdentity(db, inviteeID, AffiliateIdentityTypeInvitee, cfg.InviteeRateMultiplier, inviterID, now, expiresAt, snapshot); err != nil {
			return err
		}
	}
	return db.Model(&AffiliateIdentity{}).
		Where("source_inviter_id = ? AND identity_type = ? AND user_id NOT IN ?", inviterID, AffiliateIdentityTypeInvitee, qualified).
		Update("status", AffiliateIdentityStatusRevoked).Error
}

func RefreshAffiliateIdentitiesForUser(inviteeUserID int) error {
	var user User
	if err := DB.Select("inviter_id").Where("id = ?", inviteeUserID).First(&user).Error; err != nil {
		return err
	}
	if user.InviterId <= 0 {
		return nil
	}
	return RefreshAffiliateIdentitiesForInviter(user.InviterId)
}

func GetActiveAffiliateIdentity(userID int) *AffiliateIdentity {
	if !common.AffiliateIdentityEnabled || userID <= 0 {
		return nil
	}
	var identity AffiliateIdentity
	err := DB.Where("user_id = ? AND status = ? AND expires_at > ?", userID, AffiliateIdentityStatusActive, time.Now().Unix()).
		Order("rate_multiplier asc").
		First(&identity).Error
	if err != nil {
		return nil
	}
	return &identity
}

func ResolveAffiliateIdentityMultiplier(userID int, currentMultiplier float64) float64 {
	if currentMultiplier <= 0 || math.IsNaN(currentMultiplier) || math.IsInf(currentMultiplier, 0) {
		currentMultiplier = 1
	}
	identity := GetActiveAffiliateIdentity(userID)
	if identity == nil || identity.RateMultiplier <= 0 {
		return currentMultiplier
	}
	if identity.RateMultiplier < currentMultiplier {
		return identity.RateMultiplier
	}
	return currentMultiplier
}

func affiliateInviteeRiskFlagged(db *gorm.DB, userID int) bool {
	var fp AffiliateSignupFingerprint
	err := db.Select("risk_flagged").Where("user_id = ?", userID).First(&fp).Error
	return err == nil && fp.RiskFlagged
}

func revokeAffiliateIdentitiesForInviter(db *gorm.DB, inviterID int) error {
	if err := db.Model(&AffiliateIdentity{}).Where("user_id = ? AND identity_type = ?", inviterID, AffiliateIdentityTypeInviter).Update("status", AffiliateIdentityStatusRevoked).Error; err != nil {
		return err
	}
	return db.Model(&AffiliateIdentity{}).Where("source_inviter_id = ? AND identity_type = ?", inviterID, AffiliateIdentityTypeInvitee).Update("status", AffiliateIdentityStatusRevoked).Error
}

func upsertAffiliateIdentity(db *gorm.DB, userID int, identityType string, rateMultiplier float64, sourceInviterID int, grantedAt int64, expiresAt int64, snapshot map[string]any) error {
	if rateMultiplier <= 0 {
		return nil
	}
	snapshotJSON := "{}"
	if snapshot != nil {
		data, err := common.Marshal(snapshot)
		if err != nil {
			return err
		}
		snapshotJSON = string(data)
	}
	var identity AffiliateIdentity
	err := db.Where("user_id = ? AND identity_type = ?", userID, identityType).First(&identity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return db.Create(&AffiliateIdentity{
			UserId:                userID,
			IdentityType:          identityType,
			RateMultiplier:        rateMultiplier,
			SourceInviterId:       sourceInviterID,
			GrantedAt:             grantedAt,
			ExpiresAt:             expiresAt,
			Status:                AffiliateIdentityStatusActive,
			QualificationSnapshot: snapshotJSON,
		}).Error
	}
	if err != nil {
		return err
	}
	identity.RateMultiplier = rateMultiplier
	identity.SourceInviterId = sourceInviterID
	identity.GrantedAt = grantedAt
	identity.ExpiresAt = expiresAt
	identity.Status = AffiliateIdentityStatusActive
	identity.QualificationSnapshot = snapshotJSON
	return db.Save(&identity).Error
}

func GetUserAffiliateLedgers(userID int, pageInfo *common.PageInfo) ([]*AffiliateLedger, int64, error) {
	return listAffiliateLedgers(DB.Where("user_id = ?", userID), pageInfo)
}

func GetAffiliateRebateLedgers(pageInfo *common.PageInfo, filter AffiliateRecordFilter) ([]*AffiliateLedgerRecord, int64, error) {
	return listAffiliateLedgerRecords([]string{AffiliateLedgerActionAccrue, AffiliateLedgerActionSignupBonus}, pageInfo, filter)
}

func GetAffiliateTransferLedgers(pageInfo *common.PageInfo, filter AffiliateRecordFilter) ([]*AffiliateLedgerRecord, int64, error) {
	return listAffiliateLedgerRecords([]string{AffiliateLedgerActionTransfer}, pageInfo, filter)
}

func GetAffiliateInviteRecords(pageInfo *common.PageInfo, filter AffiliateRecordFilter) ([]*AffiliateInviteRecord, int64, error) {
	var records []*AffiliateInviteRecord
	var total int64
	if pageInfo == nil {
		pageInfo = &common.PageInfo{}
	}
	query := DB.Table("users AS u").
		Joins("LEFT JOIN users AS inviter ON inviter.id = u.inviter_id").
		Where("u.inviter_id > ?", 0)
	query = applyAffiliateUserSearch(query, filter.Search, "u", "inviter")
	query = applyAffiliateTimeFilter(query, filter.StartTime, filter.EndTime, "u.created_at")
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.
		Select("u.id AS user_id, u.username, u.email, u.aff_code, u.inviter_id, inviter.username AS inviter_username, inviter.email AS inviter_email, u.created_at").
		Order("u.id desc").
		Limit(pageInfo.GetPageSize()).
		Offset(pageInfo.GetStartIdx()).
		Scan(&records).Error; err != nil {
		return nil, 0, err
	}
	return records, total, nil
}

func GetUserAffiliateWithdrawals(userID int, pageInfo *common.PageInfo) ([]*AffiliateWithdrawal, int64, error) {
	var records []*AffiliateWithdrawal
	var total int64
	if pageInfo == nil {
		pageInfo = &common.PageInfo{}
	}
	query := DB.Model(&AffiliateWithdrawal{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("id desc").Limit(pageInfo.GetPageSize()).Offset(pageInfo.GetStartIdx()).Find(&records).Error; err != nil {
		return nil, 0, err
	}
	return records, total, nil
}

func GetAffiliateWithdrawalRecords(pageInfo *common.PageInfo, filter AffiliateRecordFilter) ([]*AffiliateWithdrawalRecord, int64, error) {
	var records []*AffiliateWithdrawalRecord
	var total int64
	if pageInfo == nil {
		pageInfo = &common.PageInfo{}
	}
	query := DB.Table("affiliate_withdrawals AS w").
		Joins("LEFT JOIN users AS u ON u.id = w.user_id")
	query = applyAffiliateSingleUserSearch(query, filter.Search, "u")
	query = applyAffiliateTimeFilter(query, filter.StartTime, filter.EndTime, "w.created_at")
	if strings.TrimSpace(filter.Status) != "" {
		query = query.Where("w.status = ?", strings.TrimSpace(filter.Status))
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.
		Select("w.*, u.username, u.email").
		Order("w.id desc").
		Limit(pageInfo.GetPageSize()).
		Offset(pageInfo.GetStartIdx()).
		Scan(&records).Error; err != nil {
		return nil, 0, err
	}
	return records, total, nil
}

func listAffiliateLedgerRecords(actions []string, pageInfo *common.PageInfo, filter AffiliateRecordFilter) ([]*AffiliateLedgerRecord, int64, error) {
	var records []*AffiliateLedgerRecord
	var total int64
	if pageInfo == nil {
		pageInfo = &common.PageInfo{}
	}
	query := DB.Table("affiliate_ledgers AS l").
		Joins("LEFT JOIN users AS u ON u.id = l.user_id").
		Joins("LEFT JOIN users AS related ON related.id = l.related_user_id").
		Where("l.action IN ?", actions)
	query = applyAffiliateUserSearch(query, filter.Search, "u", "related")
	query = applyAffiliateTimeFilter(query, filter.StartTime, filter.EndTime, "l.created_at")
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.
		Select("l.*, u.username, u.email, related.username AS related_username, related.email AS related_email").
		Order("l.id desc").
		Limit(pageInfo.GetPageSize()).
		Offset(pageInfo.GetStartIdx()).
		Scan(&records).Error; err != nil {
		return nil, 0, err
	}
	return records, total, nil
}

func applyAffiliateUserSearch(query *gorm.DB, search string, primaryAlias string, relatedAlias string) *gorm.DB {
	search = strings.TrimSpace(search)
	if search == "" {
		return query
	}
	like := "%" + search + "%"
	if id, err := strconv.Atoi(search); err == nil {
		return query.Where(
			primaryAlias+".id = ? OR "+relatedAlias+".id = ? OR "+primaryAlias+".username LIKE ? OR "+primaryAlias+".email LIKE ? OR "+relatedAlias+".username LIKE ? OR "+relatedAlias+".email LIKE ?",
			id, id, like, like, like, like,
		)
	}
	return query.Where(
		primaryAlias+".username LIKE ? OR "+primaryAlias+".email LIKE ? OR "+relatedAlias+".username LIKE ? OR "+relatedAlias+".email LIKE ?",
		like, like, like, like,
	)
}

func applyAffiliateSingleUserSearch(query *gorm.DB, search string, userAlias string) *gorm.DB {
	search = strings.TrimSpace(search)
	if search == "" {
		return query
	}
	like := "%" + search + "%"
	if id, err := strconv.Atoi(search); err == nil {
		return query.Where(userAlias+".id = ? OR "+userAlias+".username LIKE ? OR "+userAlias+".email LIKE ?", id, like, like)
	}
	return query.Where(userAlias+".username LIKE ? OR "+userAlias+".email LIKE ?", like, like)
}

func applyAffiliateTimeFilter(query *gorm.DB, startTime int64, endTime int64, column string) *gorm.DB {
	if startTime > 0 {
		query = query.Where(column+" >= ?", startTime)
	}
	if endTime > 0 {
		query = query.Where(column+" <= ?", endTime)
	}
	return query
}

func listAffiliateLedgers(query *gorm.DB, pageInfo *common.PageInfo) ([]*AffiliateLedger, int64, error) {
	var ledgers []*AffiliateLedger
	var total int64
	if pageInfo == nil {
		pageInfo = &common.PageInfo{}
	}
	if err := query.Model(&AffiliateLedger{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("id desc").Limit(pageInfo.GetPageSize()).Offset(pageInfo.GetStartIdx()).Find(&ledgers).Error; err != nil {
		return nil, 0, err
	}
	return ledgers, total, nil
}
