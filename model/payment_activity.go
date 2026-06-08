package model

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"gorm.io/gorm"
)

const (
	PaymentActivityTypeLuckyWheel       = "lucky_wheel"
	PaymentActivityTypeRechargeActivity = "recharge_activity"

	LuckyWheelInviteBonusConsumeNextSessionOnce = "next_session_once"

	RechargeActivityFulfillmentPending   = "pending"
	RechargeActivityFulfillmentFulfilled = "fulfilled"
)

type LuckyWheelAmountTier struct {
	Id            string   `json:"id"`
	Name          string   `json:"name"`
	MinAmount     float64  `json:"min_amount"`
	MaxAmount     *float64 `json:"max_amount,omitempty"`
	MinMultiplier float64  `json:"min_multiplier"`
	MaxMultiplier float64  `json:"max_multiplier"`
	DrawCount     int      `json:"draw_count"`
}

type LuckyWheelInviteBonusConfig struct {
	Enabled          bool    `json:"enabled"`
	QualifyingAmount float64 `json:"qualifying_amount"`
	BonusPerInvitee  float64 `json:"bonus_per_invitee"`
	MaxBonus         float64 `json:"max_bonus"`
	ConsumePolicy    string  `json:"consume_policy"`
}

type LuckyWheelGoldenWindowConfig struct {
	Enabled    bool    `json:"enabled"`
	Timezone   string  `json:"timezone"`
	StartTime  string  `json:"start_time"`
	EndTime    string  `json:"end_time"`
	MinAmount  float64 `json:"min_amount"`
	ExtraDraws int     `json:"extra_draws"`
	DailyQuota int     `json:"daily_quota"`
}

type LuckyWheelConfig struct {
	EligibleOrderTypes  []string                     `json:"eligible_order_types"`
	MultiplierStep      float64                      `json:"multiplier_step"`
	GlobalMaxMultiplier float64                      `json:"global_max_multiplier"`
	IntroText           string                       `json:"intro_text"`
	RulesTitle          string                       `json:"rules_title"`
	RulesItems          []string                     `json:"rules_items"`
	Prizes              []map[string]interface{}     `json:"prizes,omitempty"`
	Tiers               []map[string]interface{}     `json:"tiers,omitempty"`
	AmountTiers         []LuckyWheelAmountTier       `json:"amount_tiers"`
	InviteBonus         LuckyWheelInviteBonusConfig  `json:"invite_bonus"`
	GoldenWindow        LuckyWheelGoldenWindowConfig `json:"golden_window"`
}

type RechargeActivityPrize struct {
	Id                string  `json:"id"`
	Name              string  `json:"name"`
	RewardAmount      int64   `json:"reward_amount"`
	RewardDescription string  `json:"reward_description"`
	Probability       float64 `json:"probability"`
	MinPayAmount      float64 `json:"min_pay_amount"`
	Enabled           bool    `json:"enabled"`
	SortOrder         int     `json:"sort_order"`
}

type RechargeActivityConfig struct {
	EligibleOrderTypes []string                `json:"eligible_order_types"`
	IntroText          string                  `json:"intro_text"`
	RulesTitle         string                  `json:"rules_title"`
	RulesItems         []string                `json:"rules_items"`
	Prizes             []RechargeActivityPrize `json:"prizes"`
}

type LuckyWheelSession struct {
	Id                     int                    `json:"id"`
	UserId                 int                    `json:"user_id" gorm:"index"`
	Source                 string                 `json:"source" gorm:"type:varchar(32);uniqueIndex:idx_lucky_wheel_source_order,priority:1;index"`
	OrderType              string                 `json:"source_order_type" gorm:"type:varchar(32);uniqueIndex:idx_lucky_wheel_source_order,priority:2;index"`
	SourceOrderId          int                    `json:"source_order_id" gorm:"uniqueIndex:idx_lucky_wheel_source_order,priority:3;index"`
	SourceOrderTradeNo     string                 `json:"source_order_trade_no" gorm:"type:varchar(255);index"`
	SourcePayAmount        float64                `json:"source_pay_amount"`
	RewardBaseQuota        int64                  `json:"reward_base_quota"`
	MatchedTierId          string                 `json:"matched_tier_id" gorm:"type:varchar(64);index"`
	MatchedTierName        string                 `json:"matched_tier_name" gorm:"type:varchar(128)"`
	MinMultiplier          float64                `json:"min_multiplier"`
	MaxMultiplier          float64                `json:"max_multiplier"`
	TotalDraws             int                    `json:"total_draws"`
	CompletedDraws         int                    `json:"completed_draws" gorm:"default:0"`
	RemainingDraws         int                    `json:"remaining_draws" gorm:"-"`
	BestMultiplier         float64                `json:"best_multiplier" gorm:"default:0"`
	InviteBonusMultiplier  float64                `json:"invite_bonus_multiplier" gorm:"default:0"`
	GoldenWindowExtraDraws int                    `json:"golden_window_extra_draws" gorm:"default:0"`
	Settled                bool                   `json:"settled" gorm:"default:false;index"`
	SettledBonusQuota      int64                  `json:"settled_bonus_quota" gorm:"default:0"`
	SettledAt              int64                  `json:"settled_at" gorm:"default:0"`
	DrawRecords            []LuckyWheelDrawRecord `json:"draw_records,omitempty" gorm:"-"`
	CreateTime             int64                  `json:"created_at" gorm:"index"`
	UpdateTime             int64                  `json:"updated_at"`
}

func (s *LuckyWheelSession) BeforeCreate(tx *gorm.DB) error {
	now := common.GetTimestamp()
	if s.CreateTime == 0 {
		s.CreateTime = now
	}
	if s.UpdateTime == 0 {
		s.UpdateTime = now
	}
	return nil
}

func (s *LuckyWheelSession) BeforeUpdate(tx *gorm.DB) error {
	s.UpdateTime = common.GetTimestamp()
	return nil
}

type LuckyWheelDrawRecord struct {
	Id                    int     `json:"id"`
	SessionId             int     `json:"session_id" gorm:"uniqueIndex:idx_lucky_wheel_draw,priority:1;index"`
	UserId                int     `json:"user_id" gorm:"index"`
	DrawIndex             int     `json:"draw_index" gorm:"uniqueIndex:idx_lucky_wheel_draw,priority:2"`
	BaseMultiplier        float64 `json:"base_multiplier"`
	InviteBonusMultiplier float64 `json:"invite_bonus_multiplier"`
	FinalMultiplier       float64 `json:"final_multiplier"`
	IsBest                bool    `json:"is_best"`
	CreateTime            int64   `json:"created_at" gorm:"index"`
}

func (r *LuckyWheelDrawRecord) BeforeCreate(tx *gorm.DB) error {
	if r.CreateTime == 0 {
		r.CreateTime = common.GetTimestamp()
	}
	return nil
}

type LuckyWheelGoldenWindowClaim struct {
	Id                 int    `json:"id"`
	UserId             int    `json:"user_id" gorm:"index"`
	SessionId          int    `json:"session_id" gorm:"uniqueIndex"`
	SourceOrderTradeNo string `json:"source_order_trade_no" gorm:"type:varchar(255);index"`
	ClaimDate          string `json:"claim_date" gorm:"type:varchar(16);index"`
	CreateTime         int64  `json:"created_at" gorm:"index"`
}

func (c *LuckyWheelGoldenWindowClaim) BeforeCreate(tx *gorm.DB) error {
	if c.CreateTime == 0 {
		c.CreateTime = common.GetTimestamp()
	}
	return nil
}

type LuckyWheelInviteBonusEvent struct {
	Id                  int     `json:"id"`
	InviterUserId       int     `json:"inviter_user_id" gorm:"index"`
	InviteeUserId       int     `json:"invitee_user_id" gorm:"index"`
	SourceOrderTradeNo  string  `json:"source_order_trade_no" gorm:"type:varchar(255);uniqueIndex"`
	BonusMultiplier     float64 `json:"bonus_multiplier"`
	Consumed            bool    `json:"consumed" gorm:"default:false;index"`
	ConsumedBySessionId int     `json:"consumed_by_session_id" gorm:"default:0;index"`
	CreateTime          int64   `json:"created_at" gorm:"index"`
	UpdateTime          int64   `json:"updated_at"`
}

func (e *LuckyWheelInviteBonusEvent) BeforeCreate(tx *gorm.DB) error {
	now := common.GetTimestamp()
	if e.CreateTime == 0 {
		e.CreateTime = now
	}
	if e.UpdateTime == 0 {
		e.UpdateTime = now
	}
	return nil
}

func (e *LuckyWheelInviteBonusEvent) BeforeUpdate(tx *gorm.DB) error {
	e.UpdateTime = common.GetTimestamp()
	return nil
}

type RechargeActivityChance struct {
	Id                 int     `json:"id"`
	UserId             int     `json:"user_id" gorm:"index"`
	Source             string  `json:"source" gorm:"type:varchar(32);uniqueIndex:idx_recharge_activity_source_order,priority:1;index"`
	OrderType          string  `json:"source_order_type" gorm:"type:varchar(32);uniqueIndex:idx_recharge_activity_source_order,priority:2;index"`
	SourceOrderId      int     `json:"source_order_id" gorm:"uniqueIndex:idx_recharge_activity_source_order,priority:3;index"`
	SourceOrderTradeNo string  `json:"source_order_trade_no" gorm:"type:varchar(255);index"`
	SourcePayAmount    float64 `json:"source_pay_amount"`
	Drawn              bool    `json:"drawn" gorm:"default:false;index"`
	DrawnAt            int64   `json:"drawn_at" gorm:"default:0"`
	CreateTime         int64   `json:"created_at" gorm:"index"`
	UpdateTime         int64   `json:"updated_at"`
}

func (c *RechargeActivityChance) BeforeCreate(tx *gorm.DB) error {
	now := common.GetTimestamp()
	if c.CreateTime == 0 {
		c.CreateTime = now
	}
	if c.UpdateTime == 0 {
		c.UpdateTime = now
	}
	return nil
}

func (c *RechargeActivityChance) BeforeUpdate(tx *gorm.DB) error {
	c.UpdateTime = common.GetTimestamp()
	return nil
}

type RechargeActivityDrawRecord struct {
	Id                int      `json:"id"`
	ChanceId          int      `json:"chance_id" gorm:"uniqueIndex;index"`
	UserId            int      `json:"user_id" gorm:"index"`
	UserEmail         string   `json:"user_email" gorm:"-"`
	UserName          string   `json:"user_name" gorm:"-"`
	SourceOrderId     int      `json:"source_order_id" gorm:"index"`
	PrizeId           string   `json:"prize_id" gorm:"type:varchar(64);index"`
	PrizeName         string   `json:"prize_name" gorm:"type:varchar(128)"`
	RewardAmount      int64    `json:"reward_amount" gorm:"default:0"`
	RewardDescription string   `json:"reward_description" gorm:"type:text"`
	Probability       float64  `json:"probability"`
	MinPayAmount      float64  `json:"min_pay_amount"`
	PrizeSnapshot     string   `json:"prize_snapshot" gorm:"type:text"`
	EligiblePrizeIds  string   `json:"-" gorm:"type:text"`
	EligiblePrizeList []string `json:"eligible_prize_ids" gorm:"-"`
	FulfillmentStatus string   `json:"fulfillment_status" gorm:"type:varchar(32);default:'pending';index"`
	FulfillmentNote   string   `json:"fulfillment_note" gorm:"type:text"`
	FulfilledAt       int64    `json:"fulfilled_at" gorm:"default:0"`
	FulfilledBy       int      `json:"fulfilled_by" gorm:"default:0;index"`
	CreateTime        int64    `json:"created_at" gorm:"index"`
}

func (r *RechargeActivityDrawRecord) BeforeCreate(tx *gorm.DB) error {
	if r.CreateTime == 0 {
		r.CreateTime = common.GetTimestamp()
	}
	if r.FulfillmentStatus == "" {
		r.FulfillmentStatus = RechargeActivityFulfillmentPending
	}
	return nil
}

type LuckyWheelSummary struct {
	Enabled         bool                `json:"enabled"`
	Config          LuckyWheelConfig    `json:"config"`
	ActiveSession   *LuckyWheelSession  `json:"active_session"`
	PendingSessions []LuckyWheelSession `json:"pending_sessions"`
	HistorySessions []LuckyWheelSession `json:"history_sessions"`
}

type LuckyWheelDrawResult struct {
	SessionId         int                  `json:"session_id"`
	DrawRecord        LuckyWheelDrawRecord `json:"draw_record"`
	BestMultiplier    float64              `json:"best_multiplier"`
	RemainingDraws    int                  `json:"remaining_draws"`
	Settled           bool                 `json:"settled"`
	SettledBonusQuota int64                `json:"settled_bonus_quota"`
	Session           *LuckyWheelSession   `json:"session"`
}

type LuckyWheelMultiplierStat struct {
	Multiplier float64 `json:"multiplier"`
	DrawCount  int     `json:"draw_count"`
}

type LuckyWheelStats struct {
	Enabled                bool                       `json:"enabled"`
	TotalSessions          int64                      `json:"total_sessions"`
	PendingSessions        int64                      `json:"pending_sessions"`
	SettledSessions        int64                      `json:"settled_sessions"`
	TotalBonusQuota        int64                      `json:"total_bonus_quota"`
	RecentSessions         []LuckyWheelSession        `json:"recent_sessions"`
	MultiplierStats        []LuckyWheelMultiplierStat `json:"multiplier_stats"`
	GoldenWindowUsedToday  int64                      `json:"golden_window_used_today"`
	GoldenWindowDailyQuota int                        `json:"golden_window_daily_quota"`
}

type RechargeActivitySummary struct {
	Enabled        bool                         `json:"enabled"`
	Config         RechargeActivityConfig       `json:"config"`
	PendingChances []RechargeActivityChance     `json:"pending_chances"`
	HistoryRecords []RechargeActivityDrawRecord `json:"history_records"`
}

type RechargeActivityDrawResult struct {
	ChanceId int                        `json:"chance_id"`
	Record   RechargeActivityDrawRecord `json:"record"`
	Chance   *RechargeActivityChance    `json:"chance"`
}

type RechargeActivityStats struct {
	Enabled               bool                         `json:"enabled"`
	TotalChances          int64                        `json:"total_chances"`
	PendingChances        int64                        `json:"pending_chances"`
	DrawnChances          int64                        `json:"drawn_chances"`
	PendingFulfillments   int64                        `json:"pending_fulfillments"`
	FulfilledRecords      int64                        `json:"fulfilled_records"`
	TotalRewardAmount     int64                        `json:"total_reward_amount"`
	RecentRecords         []RechargeActivityDrawRecord `json:"recent_records"`
	RecentRecordsTotal    int64                        `json:"recent_records_total"`
	RecentRecordsPage     int                          `json:"recent_records_page"`
	RecentRecordsPageSize int                          `json:"recent_records_page_size"`
	RecentRecordsKeyword  string                       `json:"recent_records_keyword"`
}

type paymentActivityOrderSnapshot struct {
	Source             string
	OrderType          string
	SourceOrderId      int
	SourceOrderTradeNo string
	UserId             int
	SourcePayAmount    float64
	RewardBaseQuota    int64
	InviteeInviterId   int
}

func defaultLuckyWheelConfig() LuckyWheelConfig {
	return LuckyWheelConfig{
		EligibleOrderTypes:  []string{PaymentOrderTypeBalance, PaymentOrderTypeSubscription},
		MultiplierStep:      0.1,
		GlobalMaxMultiplier: 2,
		IntroText:           "完成支付后可获得转盘机会。",
		RulesTitle:          "活动规则",
		RulesItems:          []string{"支付完成后按实付金额获得抽奖机会。", "多次抽奖取最高倍率结算。"},
		AmountTiers: []LuckyWheelAmountTier{
			{Id: "default", Name: "默认档位", MinAmount: 0, MinMultiplier: 0.1, MaxMultiplier: 1, DrawCount: 1},
		},
		InviteBonus: LuckyWheelInviteBonusConfig{ConsumePolicy: LuckyWheelInviteBonusConsumeNextSessionOnce},
		GoldenWindow: LuckyWheelGoldenWindowConfig{
			Timezone:  "Asia/Shanghai",
			StartTime: "20:00",
			EndTime:   "22:00",
		},
	}
}

func defaultRechargeActivityConfig() RechargeActivityConfig {
	return RechargeActivityConfig{
		EligibleOrderTypes: []string{PaymentOrderTypeBalance, PaymentOrderTypeSubscription},
		IntroText:          "完成支付后可获得充值活动抽奖机会。",
		RulesTitle:         "活动规则",
		RulesItems:         []string{"支付完成后获得一次抽奖机会。", "中奖后由管理员人工发放。"},
		Prizes:             []RechargeActivityPrize{},
	}
}

func GetLuckyWheelConfig() (bool, LuckyWheelConfig, error) {
	cfg := defaultLuckyWheelConfig()
	enabled, err := loadPaymentActivityConfig(PaymentActivityTypeLuckyWheel, &cfg)
	return enabled, cfg, err
}

func UpdateLuckyWheelConfig(enabled bool, cfg LuckyWheelConfig) error {
	if err := validateLuckyWheelConfig(cfg); err != nil {
		return err
	}
	return savePaymentActivityConfig(PaymentActivityTypeLuckyWheel, enabled, cfg)
}

func GetRechargeActivityConfig() (bool, RechargeActivityConfig, error) {
	cfg := defaultRechargeActivityConfig()
	enabled, err := loadPaymentActivityConfig(PaymentActivityTypeRechargeActivity, &cfg)
	return enabled, cfg, err
}

func UpdateRechargeActivityConfig(enabled bool, cfg RechargeActivityConfig) error {
	if err := validateRechargeActivityConfig(cfg); err != nil {
		return err
	}
	return savePaymentActivityConfig(PaymentActivityTypeRechargeActivity, enabled, cfg)
}

func loadPaymentActivityConfig(activityType string, target any) (bool, error) {
	var stored PaymentActivityConfig
	if err := DB.Where("activity_type = ?", activityType).First(&stored).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	if strings.TrimSpace(stored.Config) != "" {
		if err := common.UnmarshalJsonStr(stored.Config, target); err != nil {
			return stored.Enabled, err
		}
	}
	return stored.Enabled, nil
}

func savePaymentActivityConfig(activityType string, enabled bool, cfg any) error {
	payload, err := common.Marshal(cfg)
	if err != nil {
		return err
	}
	var stored PaymentActivityConfig
	err = DB.Where("activity_type = ?", activityType).First(&stored).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return DB.Create(&PaymentActivityConfig{
			ActivityType: activityType,
			Enabled:      enabled,
			Config:       string(payload),
		}).Error
	}
	if err != nil {
		return err
	}
	stored.Enabled = enabled
	stored.Config = string(payload)
	return DB.Save(&stored).Error
}

func validateLuckyWheelConfig(cfg LuckyWheelConfig) error {
	if len(cfg.EligibleOrderTypes) == 0 {
		return errors.New("参与订单类型不能为空")
	}
	for _, orderType := range cfg.EligibleOrderTypes {
		if orderType != PaymentOrderTypeBalance && orderType != PaymentOrderTypeSubscription {
			return errors.New("参与订单类型不合法")
		}
	}
	if cfg.MultiplierStep <= 0 {
		return errors.New("倍率步长必须大于 0")
	}
	if cfg.GlobalMaxMultiplier <= 0 {
		return errors.New("全局最大倍率必须大于 0")
	}
	if len(cfg.AmountTiers) == 0 {
		return errors.New("金额档位不能为空")
	}
	seen := map[string]bool{}
	for _, tier := range cfg.AmountTiers {
		if strings.TrimSpace(tier.Id) == "" {
			return errors.New("档位 ID 不能为空")
		}
		if seen[tier.Id] {
			return errors.New("档位 ID 不能重复")
		}
		seen[tier.Id] = true
		if tier.DrawCount <= 0 {
			return errors.New("抽奖次数必须大于 0")
		}
		if tier.MinAmount < 0 || (tier.MaxAmount != nil && *tier.MaxAmount < tier.MinAmount) {
			return errors.New("金额档位范围不合法")
		}
		if tier.MinMultiplier < 0 || tier.MaxMultiplier < tier.MinMultiplier {
			return errors.New("倍率范围不合法")
		}
		if tier.MaxMultiplier > cfg.GlobalMaxMultiplier {
			return errors.New("档位最大倍率不能超过全局最大倍率")
		}
	}
	if cfg.GoldenWindow.Enabled {
		if cfg.GoldenWindow.ExtraDraws <= 0 || cfg.GoldenWindow.DailyQuota <= 0 {
			return errors.New("黄金窗口次数和名额必须大于 0")
		}
		if _, err := parseActivityClock(cfg.GoldenWindow.StartTime); err != nil {
			return errors.New("黄金窗口开始时间不合法")
		}
		if _, err := parseActivityClock(cfg.GoldenWindow.EndTime); err != nil {
			return errors.New("黄金窗口结束时间不合法")
		}
	}
	return nil
}

func validateRechargeActivityConfig(cfg RechargeActivityConfig) error {
	if len(cfg.EligibleOrderTypes) == 0 {
		return errors.New("参与订单类型不能为空")
	}
	for _, orderType := range cfg.EligibleOrderTypes {
		if orderType != PaymentOrderTypeBalance && orderType != PaymentOrderTypeSubscription {
			return errors.New("参与订单类型不合法")
		}
	}
	seen := map[string]bool{}
	total := 0.0
	for _, prize := range cfg.Prizes {
		if strings.TrimSpace(prize.Id) == "" {
			return errors.New("奖品 ID 不能为空")
		}
		if seen[prize.Id] {
			return errors.New("奖品 ID 不能重复")
		}
		seen[prize.Id] = true
		if prize.Probability < 0 {
			return errors.New("奖品概率不能为负数")
		}
		if prize.MinPayAmount < 0 {
			return errors.New("最低实付金额不能为负数")
		}
		if prize.Enabled {
			total += prize.Probability
		}
	}
	if total > 0 && math.Abs(total-100) > 0.000001 {
		return errors.New("启用奖品概率合计必须等于 100")
	}
	return nil
}

func GrantPaymentActivitiesForTopUpTx(tx *gorm.DB, topUp *TopUp, rewardBaseQuota int64) error {
	if tx == nil || topUp == nil || topUp.Status != common.TopUpStatusSuccess {
		return nil
	}
	if rewardBaseQuota <= 0 {
		rewardBaseQuota = topUpRewardBaseQuota(topUp)
	}
	snapshot := paymentActivityOrderSnapshot{
		Source:             PaymentOrderSourceTopUp,
		OrderType:          PaymentOrderTypeBalance,
		SourceOrderId:      topUp.Id,
		SourceOrderTradeNo: topUp.TradeNo,
		UserId:             topUp.UserId,
		SourcePayAmount:    topUp.Money,
		RewardBaseQuota:    rewardBaseQuota,
	}
	if err := grantLuckyWheelForOrderTx(tx, snapshot); err != nil {
		return err
	}
	if err := grantRechargeActivityForOrderTx(tx, snapshot); err != nil {
		return err
	}
	return grantLuckyWheelInviteBonusEventTx(tx, snapshot)
}

func GrantPaymentActivitiesForSubscriptionTx(tx *gorm.DB, order *SubscriptionOrder, plan *SubscriptionPlan) error {
	if tx == nil || order == nil || plan == nil || order.Status != common.TopUpStatusSuccess {
		return nil
	}
	snapshot := paymentActivityOrderSnapshot{
		Source:             PaymentOrderSourceSubscription,
		OrderType:          PaymentOrderTypeSubscription,
		SourceOrderId:      order.Id,
		SourceOrderTradeNo: order.TradeNo,
		UserId:             order.UserId,
		SourcePayAmount:    order.Money,
		RewardBaseQuota:    int64(plan.TotalAmount),
	}
	if err := grantLuckyWheelForOrderTx(tx, snapshot); err != nil {
		return err
	}
	if err := grantRechargeActivityForOrderTx(tx, snapshot); err != nil {
		return err
	}
	return grantLuckyWheelInviteBonusEventTx(tx, snapshot)
}

func topUpRewardBaseQuota(topUp *TopUp) int64 {
	if topUp == nil {
		return 0
	}
	switch topUp.PaymentProvider {
	case PaymentProviderStripe:
		return int64(math.Round(topUp.Money * common.QuotaPerUnit))
	case PaymentProviderCreem:
		return topUp.Amount
	default:
		return int64(math.Round(float64(topUp.Amount) * common.QuotaPerUnit))
	}
}

func grantLuckyWheelForOrderTx(tx *gorm.DB, order paymentActivityOrderSnapshot) error {
	enabled, cfg, err := getLuckyWheelConfigTx(tx)
	if err != nil || !enabled {
		return err
	}
	if !activityOrderTypeAllowed(cfg.EligibleOrderTypes, order.OrderType) {
		return nil
	}
	tier, ok := matchLuckyWheelTier(cfg.AmountTiers, order.SourcePayAmount)
	if !ok {
		return nil
	}
	extraDraws, err := claimGoldenWindowTx(tx, cfg, order)
	if err != nil {
		return err
	}
	inviteBonus, err := consumeLuckyWheelInviteBonusTx(tx, cfg, order.UserId)
	if err != nil {
		return err
	}
	session := LuckyWheelSession{
		UserId:                 order.UserId,
		Source:                 order.Source,
		OrderType:              order.OrderType,
		SourceOrderId:          order.SourceOrderId,
		SourceOrderTradeNo:     order.SourceOrderTradeNo,
		SourcePayAmount:        order.SourcePayAmount,
		RewardBaseQuota:        order.RewardBaseQuota,
		MatchedTierId:          tier.Id,
		MatchedTierName:        tier.Name,
		MinMultiplier:          tier.MinMultiplier,
		MaxMultiplier:          tier.MaxMultiplier,
		TotalDraws:             tier.DrawCount + extraDraws,
		InviteBonusMultiplier:  inviteBonus,
		GoldenWindowExtraDraws: extraDraws,
	}
	return tx.Where("source = ? AND order_type = ? AND source_order_id = ?", order.Source, order.OrderType, order.SourceOrderId).
		FirstOrCreate(&session).Error
}

func grantRechargeActivityForOrderTx(tx *gorm.DB, order paymentActivityOrderSnapshot) error {
	enabled, cfg, err := getRechargeActivityConfigTx(tx)
	if err != nil || !enabled {
		return err
	}
	if !activityOrderTypeAllowed(cfg.EligibleOrderTypes, order.OrderType) {
		return nil
	}
	chance := RechargeActivityChance{
		UserId:             order.UserId,
		Source:             order.Source,
		OrderType:          order.OrderType,
		SourceOrderId:      order.SourceOrderId,
		SourceOrderTradeNo: order.SourceOrderTradeNo,
		SourcePayAmount:    order.SourcePayAmount,
	}
	return tx.Where("source = ? AND order_type = ? AND source_order_id = ?", order.Source, order.OrderType, order.SourceOrderId).
		FirstOrCreate(&chance).Error
}

func grantLuckyWheelInviteBonusEventTx(tx *gorm.DB, order paymentActivityOrderSnapshot) error {
	enabled, cfg, err := getLuckyWheelConfigTx(tx)
	if err != nil || !enabled || !cfg.InviteBonus.Enabled || order.SourcePayAmount < cfg.InviteBonus.QualifyingAmount {
		return err
	}
	var user User
	if err := tx.Select("id", "inviter_id").Where("id = ?", order.UserId).First(&user).Error; err != nil {
		return err
	}
	if user.InviterId <= 0 {
		return nil
	}
	event := LuckyWheelInviteBonusEvent{
		InviterUserId:      user.InviterId,
		InviteeUserId:      order.UserId,
		SourceOrderTradeNo: order.SourceOrderTradeNo,
		BonusMultiplier:    cfg.InviteBonus.BonusPerInvitee,
	}
	return tx.Where("source_order_trade_no = ?", order.SourceOrderTradeNo).FirstOrCreate(&event).Error
}

func getLuckyWheelConfigTx(tx *gorm.DB) (bool, LuckyWheelConfig, error) {
	cfg := defaultLuckyWheelConfig()
	enabled, err := loadPaymentActivityConfigTx(tx, PaymentActivityTypeLuckyWheel, &cfg)
	return enabled, cfg, err
}

func getRechargeActivityConfigTx(tx *gorm.DB) (bool, RechargeActivityConfig, error) {
	cfg := defaultRechargeActivityConfig()
	enabled, err := loadPaymentActivityConfigTx(tx, PaymentActivityTypeRechargeActivity, &cfg)
	return enabled, cfg, err
}

func loadPaymentActivityConfigTx(tx *gorm.DB, activityType string, target any) (bool, error) {
	var stored PaymentActivityConfig
	if err := tx.Where("activity_type = ?", activityType).First(&stored).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	if strings.TrimSpace(stored.Config) != "" {
		if err := common.UnmarshalJsonStr(stored.Config, target); err != nil {
			return stored.Enabled, err
		}
	}
	return stored.Enabled, nil
}

func activityOrderTypeAllowed(allowed []string, orderType string) bool {
	for _, item := range allowed {
		if item == orderType {
			return true
		}
	}
	return false
}

func matchLuckyWheelTier(tiers []LuckyWheelAmountTier, amount float64) (LuckyWheelAmountTier, bool) {
	sort.SliceStable(tiers, func(i, j int) bool { return tiers[i].MinAmount > tiers[j].MinAmount })
	for _, tier := range tiers {
		if amount < tier.MinAmount {
			continue
		}
		if tier.MaxAmount != nil && amount > *tier.MaxAmount {
			continue
		}
		return tier, true
	}
	return LuckyWheelAmountTier{}, false
}

func claimGoldenWindowTx(tx *gorm.DB, cfg LuckyWheelConfig, order paymentActivityOrderSnapshot) (int, error) {
	window := cfg.GoldenWindow
	if !window.Enabled || order.SourcePayAmount < window.MinAmount {
		return 0, nil
	}
	location, err := time.LoadLocation(window.Timezone)
	if err != nil {
		location = time.Local
	}
	now := time.Now().In(location)
	if !clockInWindow(now, window.StartTime, window.EndTime) {
		return 0, nil
	}
	claimDate := now.Format("2006-01-02")
	var used int64
	if err := tx.Model(&LuckyWheelGoldenWindowClaim{}).Where("claim_date = ?", claimDate).Count(&used).Error; err != nil {
		return 0, err
	}
	if used >= int64(window.DailyQuota) {
		return 0, nil
	}
	claim := LuckyWheelGoldenWindowClaim{
		UserId:             order.UserId,
		SourceOrderTradeNo: order.SourceOrderTradeNo,
		ClaimDate:          claimDate,
	}
	if err := tx.Where("source_order_trade_no = ?", order.SourceOrderTradeNo).FirstOrCreate(&claim).Error; err != nil {
		return 0, err
	}
	return window.ExtraDraws, nil
}

func consumeLuckyWheelInviteBonusTx(tx *gorm.DB, cfg LuckyWheelConfig, userID int) (float64, error) {
	if !cfg.InviteBonus.Enabled || cfg.InviteBonus.ConsumePolicy != LuckyWheelInviteBonusConsumeNextSessionOnce {
		return 0, nil
	}
	var events []LuckyWheelInviteBonusEvent
	if err := tx.Where("inviter_user_id = ? AND consumed = ?", userID, false).Order("id asc").Find(&events).Error; err != nil {
		return 0, err
	}
	total := 0.0
	ids := make([]int, 0)
	for _, event := range events {
		if total >= cfg.InviteBonus.MaxBonus {
			break
		}
		total += event.BonusMultiplier
		ids = append(ids, event.Id)
	}
	if total > cfg.InviteBonus.MaxBonus {
		total = cfg.InviteBonus.MaxBonus
	}
	if len(ids) > 0 {
		if err := tx.Model(&LuckyWheelInviteBonusEvent{}).Where("id IN ?", ids).Updates(map[string]interface{}{"consumed": true}).Error; err != nil {
			return 0, err
		}
	}
	return total, nil
}

func GetLuckyWheelSummary(userID int) (*LuckyWheelSummary, error) {
	enabled, cfg, err := GetLuckyWheelConfig()
	if err != nil {
		return nil, err
	}
	var pending []LuckyWheelSession
	if err := DB.Where("user_id = ? AND settled = ?", userID, false).Order("id asc").Limit(20).Find(&pending).Error; err != nil {
		return nil, err
	}
	fillLuckyWheelSessions(&pending)
	var history []LuckyWheelSession
	if err := DB.Where("user_id = ? AND settled = ?", userID, true).Order("id desc").Limit(20).Find(&history).Error; err != nil {
		return nil, err
	}
	fillLuckyWheelSessions(&history)
	var active *LuckyWheelSession
	for i := range pending {
		if pending[i].RemainingDraws > 0 {
			copySession := pending[i]
			active = &copySession
			break
		}
	}
	return &LuckyWheelSummary{Enabled: enabled, Config: cfg, ActiveSession: active, PendingSessions: pending, HistorySessions: history}, nil
}

func DrawLuckyWheel(userID int, sessionID int) (*LuckyWheelDrawResult, error) {
	enabled, cfg, err := GetLuckyWheelConfig()
	if err != nil {
		return nil, err
	}
	if !enabled {
		return nil, errors.New("转盘活动未开启")
	}
	var result LuckyWheelDrawResult
	var logUserID int
	var logBonus int64
	err = DB.Transaction(func(tx *gorm.DB) error {
		var session LuckyWheelSession
		if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("id = ? AND user_id = ?", sessionID, userID).First(&session).Error; err != nil {
			return errors.New("转盘机会不存在")
		}
		if session.Settled || session.CompletedDraws >= session.TotalDraws {
			return errors.New("转盘机会已用完")
		}
		base := randomLuckyWheelMultiplier(session.MinMultiplier, session.MaxMultiplier, cfg.MultiplierStep)
		finalMultiplier := base + session.InviteBonusMultiplier
		if finalMultiplier > cfg.GlobalMaxMultiplier {
			finalMultiplier = cfg.GlobalMaxMultiplier
		}
		finalMultiplier = roundFloat(finalMultiplier, 4)
		drawIndex := session.CompletedDraws + 1
		isBest := finalMultiplier >= session.BestMultiplier
		record := LuckyWheelDrawRecord{
			SessionId:             session.Id,
			UserId:                userID,
			DrawIndex:             drawIndex,
			BaseMultiplier:        base,
			InviteBonusMultiplier: session.InviteBonusMultiplier,
			FinalMultiplier:       finalMultiplier,
			IsBest:                isBest,
		}
		if err := tx.Create(&record).Error; err != nil {
			return err
		}
		session.CompletedDraws = drawIndex
		if isBest {
			session.BestMultiplier = finalMultiplier
		}
		if session.CompletedDraws >= session.TotalDraws {
			session.Settled = true
			session.SettledAt = common.GetTimestamp()
			session.SettledBonusQuota = int64(math.Round(float64(session.RewardBaseQuota) * session.BestMultiplier))
			if session.SettledBonusQuota > 0 {
				if err := tx.Model(&User{}).Where("id = ?", userID).Update("quota", gorm.Expr("quota + ?", session.SettledBonusQuota)).Error; err != nil {
					return err
				}
				logUserID = userID
				logBonus = session.SettledBonusQuota
			}
		}
		if err := tx.Save(&session).Error; err != nil {
			return err
		}
		session.RemainingDraws = session.TotalDraws - session.CompletedDraws
		result = LuckyWheelDrawResult{
			SessionId:         session.Id,
			DrawRecord:        record,
			BestMultiplier:    session.BestMultiplier,
			RemainingDraws:    session.RemainingDraws,
			Settled:           session.Settled,
			SettledBonusQuota: session.SettledBonusQuota,
			Session:           &session,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if logUserID > 0 && logBonus > 0 {
		RecordLog(logUserID, LogTypeTopup, fmt.Sprintf("转盘活动奖励发放成功，奖励额度: %s", logger.FormatQuota(int(logBonus))))
	}
	return &result, nil
}

func randomLuckyWheelMultiplier(minValue float64, maxValue float64, step float64) float64 {
	if maxValue <= minValue || step <= 0 {
		return roundFloat(minValue, 4)
	}
	steps := int(math.Floor((maxValue-minValue)/step + 0.000001))
	if steps <= 0 {
		return roundFloat(minValue, 4)
	}
	return roundFloat(minValue+float64(rand.Intn(steps+1))*step, 4)
}

func fillLuckyWheelSessions(sessions *[]LuckyWheelSession) {
	for i := range *sessions {
		(*sessions)[i].RemainingDraws = (*sessions)[i].TotalDraws - (*sessions)[i].CompletedDraws
		var records []LuckyWheelDrawRecord
		if err := DB.Where("session_id = ?", (*sessions)[i].Id).Order("draw_index asc").Find(&records).Error; err == nil {
			(*sessions)[i].DrawRecords = records
		}
	}
}

func GetLuckyWheelStats() (*LuckyWheelStats, error) {
	enabled, cfg, err := GetLuckyWheelConfig()
	if err != nil {
		return nil, err
	}
	stats := &LuckyWheelStats{Enabled: enabled, GoldenWindowDailyQuota: cfg.GoldenWindow.DailyQuota}
	if err := DB.Model(&LuckyWheelSession{}).Count(&stats.TotalSessions).Error; err != nil {
		return nil, err
	}
	if err := DB.Model(&LuckyWheelSession{}).Where("settled = ?", false).Count(&stats.PendingSessions).Error; err != nil {
		return nil, err
	}
	if err := DB.Model(&LuckyWheelSession{}).Where("settled = ?", true).Count(&stats.SettledSessions).Error; err != nil {
		return nil, err
	}
	if err := DB.Model(&LuckyWheelSession{}).Select("COALESCE(SUM(settled_bonus_quota), 0)").Scan(&stats.TotalBonusQuota).Error; err != nil {
		return nil, err
	}
	if err := DB.Order("id desc").Limit(20).Find(&stats.RecentSessions).Error; err != nil {
		return nil, err
	}
	fillLuckyWheelSessions(&stats.RecentSessions)
	var records []LuckyWheelDrawRecord
	if err := DB.Select("final_multiplier").Find(&records).Error; err != nil {
		return nil, err
	}
	buckets := map[float64]int{}
	for _, record := range records {
		buckets[record.FinalMultiplier]++
	}
	for multiplier, count := range buckets {
		stats.MultiplierStats = append(stats.MultiplierStats, LuckyWheelMultiplierStat{Multiplier: multiplier, DrawCount: count})
	}
	sort.Slice(stats.MultiplierStats, func(i, j int) bool { return stats.MultiplierStats[i].Multiplier < stats.MultiplierStats[j].Multiplier })
	claimDate := time.Now().Format("2006-01-02")
	_ = DB.Model(&LuckyWheelGoldenWindowClaim{}).Where("claim_date = ?", claimDate).Count(&stats.GoldenWindowUsedToday).Error
	return stats, nil
}

func GetRechargeActivitySummary(userID int) (*RechargeActivitySummary, error) {
	enabled, cfg, err := GetRechargeActivityConfig()
	if err != nil {
		return nil, err
	}
	var chances []RechargeActivityChance
	if err := DB.Where("user_id = ? AND drawn = ?", userID, false).Order("id asc").Limit(20).Find(&chances).Error; err != nil {
		return nil, err
	}
	var records []RechargeActivityDrawRecord
	if err := DB.Where("user_id = ?", userID).Order("id desc").Limit(20).Find(&records).Error; err != nil {
		return nil, err
	}
	fillRechargeActivityRecords(records)
	return &RechargeActivitySummary{Enabled: enabled, Config: cfg, PendingChances: chances, HistoryRecords: records}, nil
}

func DrawRechargeActivity(userID int, chanceID int) (*RechargeActivityDrawResult, error) {
	enabled, cfg, err := GetRechargeActivityConfig()
	if err != nil {
		return nil, err
	}
	if !enabled {
		return nil, errors.New("充值活动未开启")
	}
	var result RechargeActivityDrawResult
	err = DB.Transaction(func(tx *gorm.DB) error {
		var chance RechargeActivityChance
		if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("id = ? AND user_id = ?", chanceID, userID).First(&chance).Error; err != nil {
			return errors.New("抽奖机会不存在")
		}
		if chance.Drawn {
			return errors.New("抽奖机会已使用")
		}
		prize, eligible, err := pickRechargeActivityPrize(cfg.Prizes, chance.SourcePayAmount)
		if err != nil {
			return err
		}
		snapshot, err := common.Marshal(prize)
		if err != nil {
			return err
		}
		eligiblePayload, err := common.Marshal(eligible)
		if err != nil {
			return err
		}
		record := RechargeActivityDrawRecord{
			ChanceId:          chance.Id,
			UserId:            userID,
			SourceOrderId:     chance.SourceOrderId,
			PrizeId:           prize.Id,
			PrizeName:         prize.Name,
			RewardAmount:      0,
			RewardDescription: prize.RewardDescription,
			Probability:       prize.Probability,
			MinPayAmount:      prize.MinPayAmount,
			PrizeSnapshot:     string(snapshot),
			EligiblePrizeIds:  string(eligiblePayload),
			EligiblePrizeList: eligible,
			FulfillmentStatus: RechargeActivityFulfillmentPending,
		}
		if err := tx.Create(&record).Error; err != nil {
			return err
		}
		chance.Drawn = true
		chance.DrawnAt = common.GetTimestamp()
		if err := tx.Save(&chance).Error; err != nil {
			return err
		}
		result = RechargeActivityDrawResult{ChanceId: chance.Id, Record: record, Chance: &chance}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func pickRechargeActivityPrize(prizes []RechargeActivityPrize, payAmount float64) (RechargeActivityPrize, []string, error) {
	eligible := make([]RechargeActivityPrize, 0)
	ids := make([]string, 0)
	total := 0.0
	for _, prize := range prizes {
		if !prize.Enabled || payAmount < prize.MinPayAmount {
			continue
		}
		eligible = append(eligible, prize)
		ids = append(ids, prize.Id)
		total += prize.Probability
	}
	if len(eligible) == 0 || total <= 0 {
		return RechargeActivityPrize{}, ids, errors.New("当前订单没有可抽奖品")
	}
	if len(eligible) == 1 {
		return eligible[0], ids, nil
	}
	point := rand.Float64() * total
	acc := 0.0
	for _, prize := range eligible {
		acc += prize.Probability
		if point <= acc {
			return prize, ids, nil
		}
	}
	return eligible[len(eligible)-1], ids, nil
}

func GetRechargeActivityStats(page int, pageSize int, keyword string) (*RechargeActivityStats, error) {
	enabled, _, err := GetRechargeActivityConfig()
	if err != nil {
		return nil, err
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = common.ItemsPerPage
	}
	if pageSize > 100 {
		pageSize = 100
	}
	keyword = strings.TrimSpace(keyword)
	stats := &RechargeActivityStats{Enabled: enabled, RecentRecordsPage: page, RecentRecordsPageSize: pageSize, RecentRecordsKeyword: keyword}
	if err := DB.Model(&RechargeActivityChance{}).Count(&stats.TotalChances).Error; err != nil {
		return nil, err
	}
	if err := DB.Model(&RechargeActivityChance{}).Where("drawn = ?", false).Count(&stats.PendingChances).Error; err != nil {
		return nil, err
	}
	if err := DB.Model(&RechargeActivityChance{}).Where("drawn = ?", true).Count(&stats.DrawnChances).Error; err != nil {
		return nil, err
	}
	if err := DB.Model(&RechargeActivityDrawRecord{}).Where("fulfillment_status = ?", RechargeActivityFulfillmentPending).Count(&stats.PendingFulfillments).Error; err != nil {
		return nil, err
	}
	if err := DB.Model(&RechargeActivityDrawRecord{}).Where("fulfillment_status = ?", RechargeActivityFulfillmentFulfilled).Count(&stats.FulfilledRecords).Error; err != nil {
		return nil, err
	}
	if err := DB.Model(&RechargeActivityDrawRecord{}).Select("COALESCE(SUM(reward_amount), 0)").Scan(&stats.TotalRewardAmount).Error; err != nil {
		return nil, err
	}
	query := DB.Model(&RechargeActivityDrawRecord{})
	if keyword != "" {
		var users []User
		pattern, err := sanitizeLikePattern(keyword)
		if err != nil {
			return nil, err
		}
		if err := DB.Select("id").Where("username LIKE ? ESCAPE '!' OR email LIKE ? ESCAPE '!'", pattern, pattern).Find(&users).Error; err != nil {
			return nil, err
		}
		ids := make([]int, 0, len(users))
		for _, user := range users {
			ids = append(ids, user.Id)
		}
		if len(ids) == 0 {
			query = query.Where("1 = 0")
		} else {
			query = query.Where("user_id IN ?", ids)
		}
	}
	if err := query.Count(&stats.RecentRecordsTotal).Error; err != nil {
		return nil, err
	}
	if err := query.Order("id desc").Limit(pageSize).Offset((page - 1) * pageSize).Find(&stats.RecentRecords).Error; err != nil {
		return nil, err
	}
	fillRechargeActivityRecords(stats.RecentRecords)
	return stats, nil
}

func UpdateRechargeActivityRecordFulfillment(id int, status string, note string, adminUserID int) (*RechargeActivityDrawRecord, error) {
	status = strings.TrimSpace(status)
	note = strings.TrimSpace(note)
	if status != RechargeActivityFulfillmentPending && status != RechargeActivityFulfillmentFulfilled {
		return nil, errors.New("履约状态不合法")
	}
	if status == RechargeActivityFulfillmentFulfilled && note == "" {
		return nil, errors.New("标记已发放必须填写备注")
	}
	var record RechargeActivityDrawRecord
	err := DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("id = ?", id).First(&record).Error; err != nil {
			return err
		}
		record.FulfillmentStatus = status
		record.FulfillmentNote = note
		if status == RechargeActivityFulfillmentFulfilled {
			record.FulfilledAt = common.GetTimestamp()
			record.FulfilledBy = adminUserID
		} else {
			record.FulfilledAt = 0
			record.FulfilledBy = 0
		}
		return tx.Save(&record).Error
	})
	if err != nil {
		return nil, err
	}
	fillRechargeActivityRecords([]RechargeActivityDrawRecord{record})
	return &record, nil
}

func fillRechargeActivityRecords(records []RechargeActivityDrawRecord) {
	userIDs := make([]int, 0)
	seen := map[int]bool{}
	for i := range records {
		var ids []string
		if strings.TrimSpace(records[i].EligiblePrizeIds) != "" {
			_ = common.UnmarshalJsonStr(records[i].EligiblePrizeIds, &ids)
		}
		records[i].EligiblePrizeList = ids
		if !seen[records[i].UserId] {
			seen[records[i].UserId] = true
			userIDs = append(userIDs, records[i].UserId)
		}
	}
	if len(userIDs) == 0 {
		return
	}
	var users []User
	if err := DB.Select("id", "username", "email").Where("id IN ?", userIDs).Find(&users).Error; err != nil {
		return
	}
	userMap := map[int]User{}
	for _, user := range users {
		userMap[user.Id] = user
	}
	for i := range records {
		if user, ok := userMap[records[i].UserId]; ok {
			records[i].UserName = user.Username
			records[i].UserEmail = user.Email
		}
	}
}

func parseActivityClock(value string) (int, error) {
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return 0, errors.New("invalid clock")
	}
	hour, err := parseActivityClockPart(parts[0], 23)
	if err != nil {
		return 0, err
	}
	minute, err := parseActivityClockPart(parts[1], 59)
	if err != nil {
		return 0, err
	}
	return hour*60 + minute, nil
}

func parseActivityClockPart(value string, max int) (int, error) {
	if len(value) == 0 || len(value) > 2 {
		return 0, errors.New("invalid clock part")
	}
	n := 0
	for _, r := range value {
		if r < '0' || r > '9' {
			return 0, errors.New("invalid clock part")
		}
		n = n*10 + int(r-'0')
	}
	if n < 0 || n > max {
		return 0, errors.New("invalid clock part")
	}
	return n, nil
}

func clockInWindow(now time.Time, start string, end string) bool {
	startMinutes, err := parseActivityClock(start)
	if err != nil {
		return false
	}
	endMinutes, err := parseActivityClock(end)
	if err != nil {
		return false
	}
	current := now.Hour()*60 + now.Minute()
	if startMinutes <= endMinutes {
		return current >= startMinutes && current <= endMinutes
	}
	return current >= startMinutes || current <= endMinutes
}

func roundFloat(value float64, places int) float64 {
	pow := math.Pow(10, float64(places))
	return math.Round(value*pow) / pow
}
