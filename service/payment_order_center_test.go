package service

import (
	"errors"
	"strconv"
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
	require.Equal(t, "CNY", orders[0].PayCurrency)
	require.Equal(t, "balance", orders[1].OrderType)
	require.Equal(t, "topup-order-1", orders[1].TradeNo)
	require.Equal(t, "COMPLETED", orders[1].Status)
	require.Empty(t, orders[1].PayCurrency)
}

func TestPaymentOrderDetailIncludesSubscriptionPayCurrency(t *testing.T) {
	truncate(t)
	seedUser(t, 206, 0)
	plan := seedPaymentOrderPlanWithCurrency(t, 306, "USD Plan", "USD")
	now := time.Now().Unix()
	order := model.SubscriptionOrder{
		UserId:          206,
		PlanId:          plan.Id,
		Money:           9.99,
		TradeNo:         "sub-order-usd",
		PaymentMethod:   model.PaymentMethodStripe,
		PaymentProvider: model.PaymentProviderStripe,
		CreateTime:      now,
		Status:          common.TopUpStatusPending,
	}
	require.NoError(t, model.DB.Create(&order).Error)

	detail, err := GetPaymentOrderDetail(PaymentOrderRef{OrderType: "subscription", ID: order.Id}, 206)

	require.NoError(t, err)
	require.NotNil(t, detail)
	require.NotNil(t, detail.Order)
	require.Equal(t, "subscription", detail.Order.OrderType)
	require.Equal(t, "USD Plan", detail.Order.PlanTitle)
	require.Equal(t, "USD", detail.Order.PayCurrency)
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

func TestProcessPaymentOrderRefundRejectsUnsupportedProvider(t *testing.T) {
	truncate(t)
	seedUser(t, 207, 0)
	now := time.Now().Unix()
	topUp := model.TopUp{
		UserId:          207,
		Amount:          10,
		Money:           70,
		TradeNo:         "unsupported-refund-order",
		PaymentMethod:   model.PaymentMethodStripe,
		PaymentProvider: model.PaymentProviderStripe,
		CreateTime:      now,
		CompleteTime:    now + 10,
		Status:          common.TopUpStatusSuccess,
	}
	require.NoError(t, model.DB.Create(&topUp).Error)

	refund, err := ProcessPaymentOrderRefund(PaymentOrderRef{OrderType: "balance", ID: topUp.Id}, 70, "unsupported", false, true, "admin")

	require.Nil(t, refund)
	require.ErrorContains(t, err, "不支持自动退款")
	var count int64
	require.NoError(t, model.DB.Model(&model.PaymentOrderRefund{}).Where("source_order_trade_no = ?", topUp.TradeNo).Count(&count).Error)
	require.Zero(t, count)
}

func TestProcessBalanceTopUpRefundRollsBackQuota(t *testing.T) {
	truncate(t)
	seedUser(t, 208, 500)
	oldQuotaPerUnit := common.QuotaPerUnit
	common.QuotaPerUnit = 1
	t.Cleanup(func() { common.QuotaPerUnit = oldQuotaPerUnit })
	now := time.Now().Unix()
	topUp := model.TopUp{
		UserId:          208,
		Amount:          100,
		Money:           100,
		TradeNo:         "balance-topup-refund",
		PaymentMethod:   model.PaymentMethodBalance,
		PaymentProvider: model.PaymentProviderBalance,
		CreateTime:      now,
		CompleteTime:    now + 10,
		Status:          common.TopUpStatusSuccess,
	}
	require.NoError(t, model.DB.Create(&topUp).Error)

	refund, err := ProcessPaymentOrderRefund(PaymentOrderRef{OrderType: "balance", ID: topUp.Id}, 25, "partial refund", false, true, "admin")

	require.NoError(t, err)
	require.NotNil(t, refund)
	require.Equal(t, model.PaymentRefundStatusProcessed, refund.Status)
	require.Equal(t, "balance", refund.ProviderRefundStatus)
	require.Equal(t, int64(25), refund.RollbackQuota)
	var user model.User
	require.NoError(t, model.DB.Where("id = ?", 208).First(&user).Error)
	require.Equal(t, 475, user.Quota)
}

func TestProcessBalanceSubscriptionRefundCancelsSubscriptionAndReturnsQuota(t *testing.T) {
	truncate(t)
	seedUser(t, 209, 100)
	plan := seedPaymentOrderPlan(t, 309, "Balance Plan")
	now := time.Now().Unix()
	order := model.SubscriptionOrder{
		UserId:          209,
		PlanId:          plan.Id,
		Money:           30,
		TradeNo:         "balance-sub-refund",
		PaymentMethod:   model.PaymentMethodBalance,
		PaymentProvider: model.PaymentProviderBalance,
		Status:          common.TopUpStatusSuccess,
		CreateTime:      now,
		CompleteTime:    now,
		ProviderPayload: "charged_quota=3000",
	}
	require.NoError(t, model.DB.Create(&order).Error)
	sub := model.UserSubscription{
		UserId:        209,
		PlanId:        plan.Id,
		AmountTotal:   3000,
		AmountUsed:    100,
		StartTime:     now,
		EndTime:       now + 3600,
		Status:        "active",
		Source:        model.SubscriptionSourceOrder,
		UpgradeGroup:  "vip",
		PrevUserGroup: "default",
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	require.NoError(t, model.DB.Create(&sub).Error)

	refund, err := ProcessPaymentOrderRefund(PaymentOrderRef{OrderType: "subscription", ID: order.Id}, 15, "partial refund", false, true, "admin")

	require.NoError(t, err)
	require.NotNil(t, refund)
	require.Equal(t, int64(1500), refund.RollbackQuota)
	var user model.User
	require.NoError(t, model.DB.Where("id = ?", 209).First(&user).Error)
	require.Equal(t, 1600, user.Quota)
	var updatedSub model.UserSubscription
	require.NoError(t, model.DB.Where("id = ?", sub.Id).First(&updatedSub).Error)
	require.Equal(t, "cancelled", updatedSub.Status)
	require.LessOrEqual(t, updatedSub.EndTime, common.GetTimestamp())
}

func TestProcessPaymentOrderRefundUsesProviderGatewayAndStoresProviderMetadata(t *testing.T) {
	truncate(t)
	seedUser(t, 210, 0)
	now := time.Now().Unix()
	topUp := model.TopUp{
		UserId:          210,
		Amount:          100,
		Money:           69.9,
		TradeNo:         "wechat-provider-refund",
		PaymentMethod:   model.PaymentMethodWechatPay,
		PaymentProvider: model.PaymentProviderWechatPay,
		CreateTime:      now,
		CompleteTime:    now + 10,
		Status:          common.TopUpStatusSuccess,
	}
	require.NoError(t, model.DB.Create(&topUp).Error)
	var got paymentProviderRefundRequest
	restore := setPaymentProviderRefunderForTest(func(req paymentProviderRefundRequest) (*paymentProviderRefundResult, error) {
		got = req
		return &paymentProviderRefundResult{
			ProviderRefundID:     "wx-refund-id",
			ProviderRefundStatus: "PROCESSING",
			ProviderResponse:     `{"status":"PROCESSING"}`,
		}, nil
	})
	defer restore()

	refund, err := ProcessPaymentOrderRefund(PaymentOrderRef{OrderType: "balance", ID: topUp.Id}, 69.9, "wechat refund", false, true, "admin")

	require.NoError(t, err)
	require.NotNil(t, refund)
	require.Equal(t, model.PaymentProviderWechatPay, got.Provider)
	require.Equal(t, topUp.TradeNo, got.TradeNo)
	require.Equal(t, int64(6990), got.RefundAmountCents)
	require.Equal(t, int64(6990), got.TotalAmountCents)
	require.NotEmpty(t, got.ProviderRefundNo)
	require.Equal(t, got.ProviderRefundNo, refund.ProviderRefundNo)
	require.Equal(t, "wx-refund-id", refund.ProviderRefundId)
	require.Equal(t, "PROCESSING", refund.ProviderRefundStatus)
	require.Equal(t, int64(6990), refund.RefundAmountCents)
	require.Equal(t, int64(6990), refund.TotalAmountCents)
}

func TestProcessPaymentOrderRefundRecordsFailedProviderRefundWithoutRollback(t *testing.T) {
	truncate(t)
	seedUser(t, 211, 500)
	now := time.Now().Unix()
	topUp := model.TopUp{
		UserId:          211,
		Amount:          100,
		Money:           69.9,
		TradeNo:         "wechat-provider-fails",
		PaymentMethod:   model.PaymentMethodWechatPay,
		PaymentProvider: model.PaymentProviderWechatPay,
		CreateTime:      now,
		CompleteTime:    now + 10,
		Status:          common.TopUpStatusSuccess,
	}
	require.NoError(t, model.DB.Create(&topUp).Error)
	restore := setPaymentProviderRefunderForTest(func(req paymentProviderRefundRequest) (*paymentProviderRefundResult, error) {
		return nil, errors.New("provider rejected")
	})
	defer restore()

	refund, err := ProcessPaymentOrderRefund(PaymentOrderRef{OrderType: "balance", ID: topUp.Id}, 69.9, "wechat refund", false, true, "admin")

	require.Nil(t, refund)
	require.ErrorContains(t, err, "provider rejected")
	var failed model.PaymentOrderRefund
	require.NoError(t, model.DB.Where("source_order_trade_no = ?", topUp.TradeNo).First(&failed).Error)
	require.Equal(t, model.PaymentRefundStatusFailed, failed.Status)
	require.Contains(t, failed.ProviderResponse, "provider rejected")
	var user model.User
	require.NoError(t, model.DB.Where("id = ?", 211).First(&user).Error)
	require.Equal(t, 500, user.Quota)
}

func TestProcessPaymentOrderRefundRejectsAmountAboveRemainingProcessedRefunds(t *testing.T) {
	truncate(t)
	seedUser(t, 212, 0)
	now := time.Now().Unix()
	topUp := model.TopUp{
		UserId:          212,
		Amount:          100,
		Money:           100,
		TradeNo:         "remaining-refund-order",
		PaymentMethod:   model.PaymentMethodBalance,
		PaymentProvider: model.PaymentProviderBalance,
		CreateTime:      now,
		CompleteTime:    now + 10,
		Status:          common.TopUpStatusSuccess,
	}
	require.NoError(t, model.DB.Create(&topUp).Error)
	require.NoError(t, model.DB.Create(&model.PaymentOrderRefund{
		Source:             model.PaymentOrderSourceTopUp,
		OrderType:          model.PaymentOrderTypeBalance,
		SourceOrderId:      topUp.Id,
		SourceOrderTradeNo: topUp.TradeNo,
		UserId:             topUp.UserId,
		Amount:             80,
		Reason:             "first",
		Status:             model.PaymentRefundStatusProcessed,
		RequestedBy:        "admin",
	}).Error)

	refund, err := ProcessPaymentOrderRefund(PaymentOrderRef{OrderType: "balance", ID: topUp.Id}, 30, "too much", false, true, "admin")

	require.Nil(t, refund)
	require.ErrorContains(t, err, "累计退款金额不能大于订单支付金额")
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

func TestGetPaymentDashboardCountsMoreThanListPageLimit(t *testing.T) {
	truncate(t)
	seedUser(t, 204, 0)
	now := time.Now().Unix()
	for i := 0; i < 120; i++ {
		require.NoError(t, model.DB.Create(&model.TopUp{
			UserId:          204,
			Amount:          1,
			Money:           1,
			TradeNo:         "dashboard-many-orders-" + strconv.Itoa(i),
			PaymentMethod:   model.PaymentMethodWechatPay,
			PaymentProvider: model.PaymentProviderWechatPay,
			CreateTime:      now - int64(i),
			CompleteTime:    now - int64(i),
			Status:          common.TopUpStatusSuccess,
		}).Error)
	}

	stats, err := GetPaymentDashboard(30)

	require.NoError(t, err)
	require.Equal(t, 120, stats.CompletedOrders)
	require.Equal(t, 120, stats.TotalOrders)
	require.Equal(t, 120.0, stats.TotalAmount)
}

func TestGetPaymentDashboardUsesCompleteTimeWindow(t *testing.T) {
	truncate(t)
	seedUser(t, 205, 0)
	now := time.Now()
	oldCreated := now.AddDate(0, 0, -60).Unix()
	recentComplete := now.Unix()
	oldComplete := now.AddDate(0, 0, -60).Unix()
	require.NoError(t, model.DB.Create(&model.TopUp{
		UserId:          205,
		Amount:          1,
		Money:           88,
		TradeNo:         "dashboard-created-old-completed-now",
		PaymentMethod:   model.PaymentMethodAlipayDirect,
		PaymentProvider: model.PaymentProviderAlipay,
		CreateTime:      oldCreated,
		CompleteTime:    recentComplete,
		Status:          common.TopUpStatusSuccess,
	}).Error)
	require.NoError(t, model.DB.Create(&model.TopUp{
		UserId:          205,
		Amount:          1,
		Money:           99,
		TradeNo:         "dashboard-created-now-completed-old",
		PaymentMethod:   model.PaymentMethodAlipayDirect,
		PaymentProvider: model.PaymentProviderAlipay,
		CreateTime:      recentComplete,
		CompleteTime:    oldComplete,
		Status:          common.TopUpStatusSuccess,
	}).Error)

	stats, err := GetPaymentDashboard(30)

	require.NoError(t, err)
	require.Equal(t, 1, stats.CompletedOrders)
	require.Equal(t, 1, stats.TotalOrders)
	require.Equal(t, 88.0, stats.TotalAmount)
}

func seedPaymentOrderPlan(t *testing.T, id int, title string) *model.SubscriptionPlan {
	t.Helper()
	return seedPaymentOrderPlanWithCurrency(t, id, title, "CNY")
}

func seedPaymentOrderPlanWithCurrency(t *testing.T, id int, title string, currency string) *model.SubscriptionPlan {
	t.Helper()
	plan := &model.SubscriptionPlan{
		Id:            id,
		Title:         title,
		PriceAmount:   29,
		Currency:      currency,
		DurationUnit:  model.SubscriptionDurationMonth,
		DurationValue: 1,
		Enabled:       true,
		TotalAmount:   1000,
	}
	require.NoError(t, model.DB.Create(plan).Error)
	return plan
}
