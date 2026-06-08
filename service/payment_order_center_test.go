package service

import (
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/stretchr/testify/require"
)

func TestListPaymentOrdersAggregatesTopUpsAndSubscriptions(t *testing.T) {
	truncate(t)
	seedUser(t, 201, 0)
	plan := seedPaymentOrderPlan(t, 301, "Pro Plan")
	now := time.Now().Unix()

	require.NoError(t, model.DB.Create(&model.TopUp{
		UserId:          201,
		Amount:          10,
		Money:           68,
		TradeNo:         "topup-order-1",
		PaymentMethod:   model.PaymentMethodWechatPay,
		PaymentProvider: model.PaymentProviderWechatPay,
		CreateTime:      now,
		CompleteTime:    now + 10,
		Status:          common.TopUpStatusSuccess,
	}).Error)
	require.NoError(t, model.DB.Create(&model.SubscriptionOrder{
		UserId:          201,
		PlanId:          plan.Id,
		Money:           29,
		TradeNo:         "sub-order-1",
		PaymentMethod:   model.PaymentMethodAlipayDirect,
		PaymentProvider: model.PaymentProviderAlipay,
		CreateTime:      now + 20,
		Status:          common.TopUpStatusPending,
	}).Error)

	orders, total, err := ListPaymentOrders(PaymentOrderListParams{
		UserID:   201,
		Page:     1,
		PageSize: 10,
	})

	require.NoError(t, err)
	require.Equal(t, 2, total)
	require.Len(t, orders, 2)
	require.Equal(t, "subscription", orders[0].OrderType)
	require.Equal(t, "sub-order-1", orders[0].TradeNo)
	require.Equal(t, "PENDING", orders[0].Status)
	require.Equal(t, "Pro Plan", orders[0].PlanTitle)
	require.Equal(t, "balance", orders[1].OrderType)
	require.Equal(t, "topup-order-1", orders[1].TradeNo)
	require.Equal(t, "COMPLETED", orders[1].Status)
}

func TestRequestPaymentOrderRefundOverlaysOrderStatus(t *testing.T) {
	truncate(t)
	seedUser(t, 202, 0)
	now := time.Now().Unix()
	require.NoError(t, model.DB.Create(&model.TopUp{
		UserId:          202,
		Amount:          10,
		Money:           70,
		TradeNo:         "refund-order-1",
		PaymentMethod:   model.PaymentMethodStripe,
		PaymentProvider: model.PaymentProviderStripe,
		CreateTime:      now,
		CompleteTime:    now + 10,
		Status:          common.TopUpStatusSuccess,
	}).Error)

	err := RequestPaymentOrderRefund(PaymentOrderRef{OrderType: "balance", ID: 1}, 202, "user changed mind")
	require.NoError(t, err)

	orders, _, err := ListPaymentOrders(PaymentOrderListParams{UserID: 202, Page: 1, PageSize: 10})
	require.NoError(t, err)
	require.Len(t, orders, 1)
	require.Equal(t, "REFUND_REQUESTED", orders[0].Status)
	require.Equal(t, "user changed mind", orders[0].RefundRequestReason)
}

func TestGrantRechargeActivityChanceForCompletedOrderIsIdempotent(t *testing.T) {
	truncate(t)
	seedUser(t, 203, 0)
	now := time.Now().Unix()
	order := &PaymentOrderItem{
		Source:      "topup",
		OrderType:   "balance",
		ID:          1001,
		UserID:      203,
		TradeNo:     "activity-order-1",
		PayAmount:   88,
		Status:      "COMPLETED",
		CreatedTime: now,
	}

	require.NoError(t, GrantPaymentActivityChanceForOrder(order, "recharge_activity", "tier_88", 1))
	require.NoError(t, GrantPaymentActivityChanceForOrder(order, "recharge_activity", "tier_88", 1))

	var count int64
	require.NoError(t, model.DB.Model(&model.PaymentActivityChance{}).Where("source_order_trade_no = ?", "activity-order-1").Count(&count).Error)
	require.Equal(t, int64(1), count)
}

func seedPaymentOrderPlan(t *testing.T, id int, title string) *model.SubscriptionPlan {
	t.Helper()
	plan := &model.SubscriptionPlan{
		Id:            id,
		Title:         title,
		PriceAmount:   29,
		Currency:      "CNY",
		DurationUnit:  model.SubscriptionDurationMonth,
		DurationValue: 1,
		Enabled:       true,
		TotalAmount:   1000,
	}
	require.NoError(t, model.DB.Create(plan).Error)
	return plan
}
