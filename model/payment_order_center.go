package model

import (
	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

const (
	PaymentOrderSourceTopUp        = "topup"
	PaymentOrderSourceSubscription = "subscription"

	PaymentOrderTypeBalance      = "balance"
	PaymentOrderTypeSubscription = "subscription"

	PaymentRefundStatusRequested = "requested"
	PaymentRefundStatusProcessed = "processed"
	PaymentRefundStatusFailed    = "failed"
)

type PaymentOrderRefund struct {
	Id                 int     `json:"id"`
	Source             string  `json:"source" gorm:"type:varchar(32);index:idx_payment_refund_source_order,priority:1"`
	OrderType          string  `json:"order_type" gorm:"type:varchar(32);index"`
	SourceOrderId      int     `json:"source_order_id" gorm:"index:idx_payment_refund_source_order,priority:2"`
	SourceOrderTradeNo string  `json:"source_order_trade_no" gorm:"type:varchar(255);index"`
	UserId             int     `json:"user_id" gorm:"index"`
	Amount             float64 `json:"amount"`
	Reason             string  `json:"reason" gorm:"type:text"`
	Status             string  `json:"status" gorm:"type:varchar(32);index"`
	RequestedBy        string  `json:"requested_by" gorm:"type:varchar(32);default:''"`
	Force              bool    `json:"force" gorm:"default:false"`
	DeductBalance      bool    `json:"deduct_balance" gorm:"default:false"`
	CreateTime         int64   `json:"create_time" gorm:"index"`
	UpdateTime         int64   `json:"update_time"`
	ProcessTime        int64   `json:"process_time" gorm:"default:0"`
}

func (r *PaymentOrderRefund) BeforeCreate(tx *gorm.DB) error {
	now := common.GetTimestamp()
	if r.CreateTime == 0 {
		r.CreateTime = now
	}
	if r.UpdateTime == 0 {
		r.UpdateTime = now
	}
	if r.Status == "" {
		r.Status = PaymentRefundStatusRequested
	}
	return nil
}

func (r *PaymentOrderRefund) BeforeUpdate(tx *gorm.DB) error {
	r.UpdateTime = common.GetTimestamp()
	return nil
}

type PaymentOrderAuditLog struct {
	Id                 int    `json:"id"`
	Source             string `json:"source" gorm:"type:varchar(32);index"`
	OrderType          string `json:"order_type" gorm:"type:varchar(32);index"`
	SourceOrderId      int    `json:"source_order_id" gorm:"index"`
	SourceOrderTradeNo string `json:"source_order_trade_no" gorm:"type:varchar(255);index"`
	UserId             int    `json:"user_id" gorm:"index"`
	Action             string `json:"action" gorm:"type:varchar(64);index"`
	Detail             string `json:"detail" gorm:"type:text"`
	Operator           string `json:"operator" gorm:"type:varchar(64);default:''"`
	CreateTime         int64  `json:"create_time" gorm:"index"`
}

func (l *PaymentOrderAuditLog) BeforeCreate(tx *gorm.DB) error {
	if l.CreateTime == 0 {
		l.CreateTime = common.GetTimestamp()
	}
	return nil
}

type PaymentActivityChance struct {
	Id                 int     `json:"id"`
	ActivityType       string  `json:"activity_type" gorm:"type:varchar(64);uniqueIndex:idx_payment_activity_order_tier,priority:1;index"`
	TierId             string  `json:"tier_id" gorm:"type:varchar(64);uniqueIndex:idx_payment_activity_order_tier,priority:3;index"`
	UserId             int     `json:"user_id" gorm:"index"`
	Source             string  `json:"source" gorm:"type:varchar(32)"`
	OrderType          string  `json:"order_type" gorm:"type:varchar(32);index"`
	SourceOrderId      int     `json:"source_order_id" gorm:"index"`
	SourceOrderTradeNo string  `json:"source_order_trade_no" gorm:"type:varchar(255);uniqueIndex:idx_payment_activity_order_tier,priority:2;index"`
	PayAmount          float64 `json:"pay_amount"`
	Chances            int     `json:"chances" gorm:"default:1"`
	UsedChances        int     `json:"used_chances" gorm:"default:0"`
	Status             string  `json:"status" gorm:"type:varchar(32);default:'pending';index"`
	CreateTime         int64   `json:"create_time" gorm:"index"`
	UpdateTime         int64   `json:"update_time"`
}

func (c *PaymentActivityChance) BeforeCreate(tx *gorm.DB) error {
	now := common.GetTimestamp()
	if c.CreateTime == 0 {
		c.CreateTime = now
	}
	if c.UpdateTime == 0 {
		c.UpdateTime = now
	}
	if c.Chances <= 0 {
		c.Chances = 1
	}
	if c.Status == "" {
		c.Status = "pending"
	}
	return nil
}

func (c *PaymentActivityChance) BeforeUpdate(tx *gorm.DB) error {
	c.UpdateTime = common.GetTimestamp()
	return nil
}

type PaymentActivityConfig struct {
	Id           int    `json:"id"`
	ActivityType string `json:"activity_type" gorm:"type:varchar(64);uniqueIndex"`
	Enabled      bool   `json:"enabled" gorm:"default:false"`
	Config       string `json:"config" gorm:"type:text"`
	CreateTime   int64  `json:"create_time" gorm:"index"`
	UpdateTime   int64  `json:"update_time"`
}

func (c *PaymentActivityConfig) BeforeCreate(tx *gorm.DB) error {
	now := common.GetTimestamp()
	if c.CreateTime == 0 {
		c.CreateTime = now
	}
	if c.UpdateTime == 0 {
		c.UpdateTime = now
	}
	return nil
}

func (c *PaymentActivityConfig) BeforeUpdate(tx *gorm.DB) error {
	c.UpdateTime = common.GetTimestamp()
	return nil
}
