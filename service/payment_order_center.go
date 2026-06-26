package service

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"gorm.io/gorm"
)

const (
	UnifiedOrderStatusPending         = "PENDING"
	UnifiedOrderStatusCompleted       = "COMPLETED"
	UnifiedOrderStatusExpired         = "EXPIRED"
	UnifiedOrderStatusCancelled       = "CANCELLED"
	UnifiedOrderStatusFailed          = "FAILED"
	UnifiedOrderStatusRefundRequested = "REFUND_REQUESTED"
	UnifiedOrderStatusRefunded        = "REFUNDED"
	UnifiedOrderStatusRefundFailed    = "REFUND_FAILED"
)

type PaymentOrderListParams struct {
	UserID      int
	Page        int
	PageSize    int
	Status      string
	OrderType   string
	PaymentType string
	Keyword     string
}

type PaymentOrderRef struct {
	OrderType string `json:"order_type"`
	ID        int    `json:"id"`
}

type PaymentOrderItem struct {
	Source              string  `json:"source"`
	OrderType           string  `json:"order_type"`
	ID                  int     `json:"id"`
	UserID              int     `json:"user_id"`
	TradeNo             string  `json:"trade_no"`
	PaymentMethod       string  `json:"payment_method"`
	PaymentProvider     string  `json:"payment_provider"`
	Status              string  `json:"status"`
	Amount              int64   `json:"amount"`
	Money               float64 `json:"money"`
	PayAmount           float64 `json:"pay_amount"`
	PayCurrency         string  `json:"pay_currency,omitempty"`
	PlanID              int     `json:"plan_id,omitempty"`
	PlanTitle           string  `json:"plan_title,omitempty"`
	RefundAmount        float64 `json:"refund_amount,omitempty"`
	RefundReason        string  `json:"refund_reason,omitempty"`
	RefundRequestReason string  `json:"refund_request_reason,omitempty"`
	CreatedTime         int64   `json:"created_time"`
	CompleteTime        int64   `json:"complete_time"`
}

type PaymentOrderDetail struct {
	Order     *PaymentOrderItem            `json:"order"`
	AuditLogs []model.PaymentOrderAuditLog `json:"audit_logs"`
	Refunds   []model.PaymentOrderRefund   `json:"refunds"`
}

type PaymentDashboardStats struct {
	TotalOrders     int                         `json:"total_orders"`
	CompletedOrders int                         `json:"completed_orders"`
	PendingOrders   int                         `json:"pending_orders"`
	FailedOrders    int                         `json:"failed_orders"`
	RefundedOrders  int                         `json:"refunded_orders"`
	TotalAmount     float64                     `json:"total_amount"`
	CompletionRate  float64                     `json:"completion_rate"`
	PaymentMethods  []PaymentDashboardMethod    `json:"payment_methods"`
	TopUsers        []PaymentDashboardUser      `json:"top_users"`
	DailySeries     []PaymentDashboardDailyItem `json:"daily_series"`
}

type PaymentDashboardMethod struct {
	Type   string  `json:"type"`
	Count  int     `json:"count"`
	Amount float64 `json:"amount"`
}

type PaymentDashboardUser struct {
	UserID int     `json:"user_id"`
	Amount float64 `json:"amount"`
	Count  int     `json:"count"`
}

type PaymentDashboardDailyItem struct {
	Date   string  `json:"date"`
	Count  int     `json:"count"`
	Amount float64 `json:"amount"`
}

type paymentDashboardRawOrder struct {
	UserID       int
	PaymentType  string
	Status       string
	PayAmount    float64
	CreatedTime  int64
	CompleteTime int64
}

type subscriptionPlanOrderMeta struct {
	Title    string
	Currency string
}

func ListPaymentOrders(params PaymentOrderListParams) ([]PaymentOrderItem, int, error) {
	normalizePaymentOrderListParams(&params)
	items := make([]PaymentOrderItem, 0)

	if params.OrderType == "" || params.OrderType == model.PaymentOrderTypeBalance {
		topUps, err := listTopUpOrders(params)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, topUps...)
	}
	if params.OrderType == "" || params.OrderType == model.PaymentOrderTypeSubscription {
		subscriptions, err := listSubscriptionOrders(params)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, subscriptions...)
	}

	if len(items) > 0 {
		if err := overlayPaymentOrderRefunds(items); err != nil {
			return nil, 0, err
		}
	}
	items = filterPaymentOrderItems(items, params)
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].CreatedTime == items[j].CreatedTime {
			return items[i].ID > items[j].ID
		}
		return items[i].CreatedTime > items[j].CreatedTime
	})
	total := len(items)
	start := (params.Page - 1) * params.PageSize
	if start >= total {
		return []PaymentOrderItem{}, total, nil
	}
	end := start + params.PageSize
	if end > total {
		end = total
	}
	return items[start:end], total, nil
}

func GetPaymentOrderDetail(ref PaymentOrderRef, userID int) (*PaymentOrderDetail, error) {
	item, err := getPaymentOrderItem(ref)
	if err != nil {
		return nil, err
	}
	if userID > 0 && item.UserID != userID {
		return nil, errors.New("订单不存在")
	}
	var audits []model.PaymentOrderAuditLog
	if err := model.DB.Where("source = ? AND source_order_id = ?", item.Source, item.ID).Order("id desc").Find(&audits).Error; err != nil {
		return nil, err
	}
	var refunds []model.PaymentOrderRefund
	if err := model.DB.Where("source = ? AND source_order_id = ?", item.Source, item.ID).Order("id desc").Find(&refunds).Error; err != nil {
		return nil, err
	}
	return &PaymentOrderDetail{Order: item, AuditLogs: audits, Refunds: refunds}, nil
}

func CancelPaymentOrder(ref PaymentOrderRef, userID int, operator string) error {
	item, err := getPaymentOrderItem(ref)
	if err != nil {
		return err
	}
	if userID > 0 && item.UserID != userID {
		return errors.New("订单不存在")
	}
	if item.Status != UnifiedOrderStatusPending {
		return errors.New("只有待支付订单可以取消")
	}
	return model.DB.Transaction(func(tx *gorm.DB) error {
		switch item.Source {
		case model.PaymentOrderSourceTopUp:
			if err := tx.Model(&model.TopUp{}).Where("id = ? AND status = ?", item.ID, common.TopUpStatusPending).Updates(map[string]interface{}{
				"status":        "cancelled",
				"complete_time": common.GetTimestamp(),
			}).Error; err != nil {
				return err
			}
		case model.PaymentOrderSourceSubscription:
			if err := tx.Model(&model.SubscriptionOrder{}).Where("id = ? AND status = ?", item.ID, common.TopUpStatusPending).Updates(map[string]interface{}{
				"status":        "cancelled",
				"complete_time": common.GetTimestamp(),
			}).Error; err != nil {
				return err
			}
		default:
			return errors.New("不支持的订单类型")
		}
		return createPaymentOrderAuditTx(tx, item, "cancelled", "订单已取消", operator)
	})
}

func RetryPaymentOrderFulfillment(ref PaymentOrderRef, operator string) error {
	item, err := getPaymentOrderItem(ref)
	if err != nil {
		return err
	}
	if item.Status != UnifiedOrderStatusFailed && item.Status != UnifiedOrderStatusPending {
		return errors.New("当前订单不支持重试履约")
	}
	switch item.Source {
	case model.PaymentOrderSourceTopUp:
		if item.Status == UnifiedOrderStatusFailed {
			if err := model.DB.Model(&model.TopUp{}).Where("id = ?", item.ID).Update("status", common.TopUpStatusPending).Error; err != nil {
				return err
			}
		}
		if err := model.ManualCompleteTopUp(item.TradeNo, operator); err != nil {
			return err
		}
	case model.PaymentOrderSourceSubscription:
		if item.Status == UnifiedOrderStatusFailed {
			if err := model.DB.Model(&model.SubscriptionOrder{}).Where("id = ?", item.ID).Update("status", common.TopUpStatusPending).Error; err != nil {
				return err
			}
		}
		if err := model.CompleteSubscriptionOrder(item.TradeNo, `{"manual_retry":true}`, item.PaymentProvider, item.PaymentMethod); err != nil {
			return err
		}
	default:
		return errors.New("不支持的订单类型")
	}
	refreshed, err := getPaymentOrderItem(ref)
	if err != nil {
		return err
	}
	return model.DB.Transaction(func(tx *gorm.DB) error {
		return createPaymentOrderAuditTx(tx, refreshed, "retry_fulfillment", "管理员重试履约", operator)
	})
}

func ProcessPaymentOrderRefund(ref PaymentOrderRef, amount float64, reason string, force bool, deductBalance bool, operator string) (*model.PaymentOrderRefund, error) {
	if amount <= 0 {
		return nil, errors.New("退款金额必须大于 0")
	}
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return nil, errors.New("退款原因不能为空")
	}
	item, err := getPaymentOrderItem(ref)
	if err != nil {
		return nil, err
	}
	if !isRefundableOrderStatus(item.Status) {
		return nil, errors.New("当前订单不支持退款")
	}
	provider := normalizePaymentRefundProvider(item)
	if !isAutoRefundProviderSupported(provider) {
		return nil, fmt.Errorf("支付渠道 %s 不支持自动退款，请手动处理", provider)
	}
	processedAmount, processedRollbackQuota, err := processedRefundTotals(item)
	if err != nil {
		return nil, err
	}
	if !force && processedAmount+amount-item.PayAmount > 0.000001 {
		return nil, errors.New("累计退款金额不能大于订单支付金额")
	}

	refundNo := buildProviderRefundNo(item)
	totalCents := paymentOrderProviderTotalCents(item)
	refundCents := paymentOrderRefundCents(amount, item.PayAmount, totalCents)
	if totalCents == 0 {
		totalCents = MoneyToCNYCents(item.PayAmount)
	}

	providerResult := &paymentProviderRefundResult{ProviderRefundStatus: "balance", ProviderResponse: "{}"}
	if provider != model.PaymentProviderBalance {
		req := paymentProviderRefundRequest{
			Provider:          provider,
			TradeNo:           item.TradeNo,
			ProviderRefundNo:  refundNo,
			Reason:            reason,
			RefundAmountCents: refundCents,
			TotalAmountCents:  totalCents,
		}
		providerResult, err = paymentProviderRefunder(req)
		if err != nil {
			_ = createFailedPaymentOrderRefund(item, amount, reason, force, refundNo, refundCents, totalCents, err, operator)
			return nil, err
		}
	}

	var refund model.PaymentOrderRefund
	var quotaCacheDelta int64
	var groupCacheUserID int
	var groupCacheValue string
	err = model.DB.Transaction(func(tx *gorm.DB) error {
		rollbackQuota, quotaDelta, err := rollbackPaymentOrderBenefitsTx(tx, item, amount, processedRollbackQuota)
		if err != nil {
			return err
		}
		refund = model.PaymentOrderRefund{
			Source:               item.Source,
			OrderType:            item.OrderType,
			SourceOrderId:        item.ID,
			SourceOrderTradeNo:   item.TradeNo,
			UserId:               item.UserID,
			Amount:               amount,
			Reason:               reason,
			Status:               model.PaymentRefundStatusProcessed,
			RequestedBy:          "admin",
			Force:                force,
			DeductBalance:        rollbackQuota > 0 || deductBalance,
			ProviderRefundNo:     refundNo,
			ProviderRefundId:     providerResult.ProviderRefundID,
			ProviderRefundStatus: providerResult.ProviderRefundStatus,
			ProviderResponse:     providerResult.ProviderResponse,
			RefundCurrency:       "CNY",
			RefundAmountCents:    refundCents,
			TotalAmountCents:     totalCents,
			RollbackQuota:        rollbackQuota,
			ProcessTime:          common.GetTimestamp(),
		}
		if err := tx.Create(&refund).Error; err != nil {
			return err
		}
		if item.Source == model.PaymentOrderSourceSubscription {
			cacheUserID, cacheGroup, err := cancelSubscriptionForRefundTx(tx, item)
			if err != nil {
				return err
			}
			groupCacheUserID = cacheUserID
			groupCacheValue = cacheGroup
		}
		quotaCacheDelta = quotaDelta
		return createPaymentOrderAuditTx(tx, item, "refund_processed", fmt.Sprintf("退款 %.2f：%s", amount, reason), operator)
	})
	if err != nil {
		return nil, err
	}
	if quotaCacheDelta != 0 {
		_ = model.UpdateUserQuotaCacheDelta(item.UserID, quotaCacheDelta)
	}
	if groupCacheUserID > 0 && groupCacheValue != "" {
		_ = model.UpdateUserGroupCache(groupCacheUserID, groupCacheValue)
	}
	return &refund, nil
}

func isRefundableOrderStatus(status string) bool {
	switch status {
	case UnifiedOrderStatusCompleted, UnifiedOrderStatusRefundRequested, UnifiedOrderStatusRefunded, UnifiedOrderStatusRefundFailed:
		return true
	default:
		return false
	}
}

func normalizePaymentRefundProvider(item *PaymentOrderItem) string {
	if item == nil {
		return ""
	}
	provider := strings.TrimSpace(item.PaymentProvider)
	if provider == model.PaymentMethodAlipayDirect {
		return model.PaymentProviderAlipay
	}
	if provider != "" {
		return provider
	}
	switch strings.TrimSpace(item.PaymentMethod) {
	case model.PaymentMethodWechatPay:
		return model.PaymentProviderWechatPay
	case model.PaymentMethodAlipayDirect:
		return model.PaymentProviderAlipay
	case model.PaymentMethodBalance:
		return model.PaymentProviderBalance
	default:
		return strings.TrimSpace(item.PaymentMethod)
	}
}

func isAutoRefundProviderSupported(provider string) bool {
	switch provider {
	case model.PaymentProviderWechatPay, model.PaymentProviderAlipay, model.PaymentProviderBalance:
		return true
	default:
		return false
	}
}

func processedRefundTotals(item *PaymentOrderItem) (float64, int64, error) {
	var refunds []model.PaymentOrderRefund
	if err := model.DB.Where("source = ? AND source_order_id = ? AND status = ?", item.Source, item.ID, model.PaymentRefundStatusProcessed).Find(&refunds).Error; err != nil {
		return 0, 0, err
	}
	var amount float64
	var rollbackQuota int64
	for _, refund := range refunds {
		amount += refund.Amount
		rollbackQuota += refund.RollbackQuota
	}
	return amount, rollbackQuota, nil
}

func buildProviderRefundNo(item *PaymentOrderItem) string {
	return fmt.Sprintf("PAYREF-%s-%d-%d-%s", item.OrderType, item.ID, common.GetTimestamp(), common.GetRandomString(6))
}

func paymentOrderProviderTotalCents(item *PaymentOrderItem) int64 {
	if item == nil || item.Source != model.PaymentOrderSourceSubscription {
		return 0
	}
	var order model.SubscriptionOrder
	if err := model.DB.Select("provider_payload").Where("id = ?", item.ID).First(&order).Error; err != nil {
		return 0
	}
	return providerPayloadTotalCents(order.ProviderPayload)
}

func providerPayloadTotalCents(payload string) int64 {
	payload = strings.TrimSpace(payload)
	if payload == "" {
		return 0
	}
	var parsed struct {
		Amount struct {
			Total      int64 `json:"total"`
			PayerTotal int64 `json:"payer_total"`
		} `json:"amount"`
		TotalAmount string `json:"total_amount"`
	}
	if err := common.UnmarshalJsonStr(payload, &parsed); err != nil {
		return 0
	}
	if parsed.Amount.Total > 0 {
		return parsed.Amount.Total
	}
	if parsed.Amount.PayerTotal > 0 {
		return parsed.Amount.PayerTotal
	}
	if parsed.TotalAmount != "" {
		value, err := strconv.ParseFloat(parsed.TotalAmount, 64)
		if err == nil {
			return MoneyToCNYCents(value)
		}
	}
	return 0
}

func paymentOrderRefundCents(refundAmount float64, payAmount float64, totalCents int64) int64 {
	if totalCents <= 0 || payAmount <= 0 {
		return MoneyToCNYCents(refundAmount)
	}
	if math.Abs(refundAmount-payAmount) <= 0.01 {
		return totalCents
	}
	return int64(math.Round(float64(totalCents) * refundAmount / payAmount))
}

func createFailedPaymentOrderRefund(item *PaymentOrderItem, amount float64, reason string, force bool, refundNo string, refundCents int64, totalCents int64, cause error, operator string) error {
	return model.DB.Transaction(func(tx *gorm.DB) error {
		refund := model.PaymentOrderRefund{
			Source:               item.Source,
			OrderType:            item.OrderType,
			SourceOrderId:        item.ID,
			SourceOrderTradeNo:   item.TradeNo,
			UserId:               item.UserID,
			Amount:               amount,
			Reason:               reason,
			Status:               model.PaymentRefundStatusFailed,
			RequestedBy:          "admin",
			Force:                force,
			ProviderRefundNo:     refundNo,
			ProviderRefundStatus: "failed",
			ProviderResponse:     cause.Error(),
			RefundCurrency:       "CNY",
			RefundAmountCents:    refundCents,
			TotalAmountCents:     totalCents,
			ProcessTime:          common.GetTimestamp(),
		}
		if err := tx.Create(&refund).Error; err != nil {
			return err
		}
		return createPaymentOrderAuditTx(tx, item, "refund_failed", cause.Error(), operator)
	})
}

func rollbackPaymentOrderBenefitsTx(tx *gorm.DB, item *PaymentOrderItem, refundAmount float64, processedRollbackQuota int64) (int64, int64, error) {
	switch item.Source {
	case model.PaymentOrderSourceTopUp:
		totalQuota := int64(math.Round(float64(item.Amount) * common.QuotaPerUnit))
		rollbackQuota := proportionalQuota(totalQuota, refundAmount, item.PayAmount, processedRollbackQuota)
		if rollbackQuota <= 0 {
			return 0, 0, nil
		}
		actualDelta, err := applyUserQuotaDeltaTx(tx, item.UserID, -rollbackQuota)
		return -actualDelta, actualDelta, err
	case model.PaymentOrderSourceSubscription:
		if normalizePaymentRefundProvider(item) != model.PaymentProviderBalance {
			return 0, 0, nil
		}
		totalQuota, err := subscriptionBalanceChargedQuotaTx(tx, item.ID)
		if err != nil {
			return 0, 0, err
		}
		rollbackQuota := proportionalQuota(totalQuota, refundAmount, item.PayAmount, processedRollbackQuota)
		if rollbackQuota <= 0 {
			return 0, 0, nil
		}
		actualDelta, err := applyUserQuotaDeltaTx(tx, item.UserID, rollbackQuota)
		return actualDelta, actualDelta, err
	default:
		return 0, 0, errors.New("不支持的订单类型")
	}
}

func proportionalQuota(totalQuota int64, refundAmount float64, payAmount float64, processedRollbackQuota int64) int64 {
	if totalQuota <= 0 || refundAmount <= 0 || payAmount <= 0 {
		return 0
	}
	var quota int64
	if math.Abs(refundAmount-payAmount) <= 0.01 {
		quota = totalQuota
	} else {
		quota = int64(math.Round(float64(totalQuota) * refundAmount / payAmount))
	}
	remaining := totalQuota - processedRollbackQuota
	if remaining < 0 {
		remaining = 0
	}
	if quota > remaining {
		quota = remaining
	}
	return quota
}

func subscriptionBalanceChargedQuotaTx(tx *gorm.DB, orderID int) (int64, error) {
	var order model.SubscriptionOrder
	if err := tx.Select("provider_payload", "money").Where("id = ?", orderID).First(&order).Error; err != nil {
		return 0, err
	}
	payload := strings.TrimSpace(order.ProviderPayload)
	if strings.HasPrefix(payload, "charged_quota=") {
		value, err := strconv.ParseInt(strings.TrimPrefix(payload, "charged_quota="), 10, 64)
		if err == nil && value > 0 {
			return value, nil
		}
	}
	return int64(math.Ceil(order.Money * common.QuotaPerUnit)), nil
}

func applyUserQuotaDeltaTx(tx *gorm.DB, userID int, delta int64) (int64, error) {
	if delta == 0 {
		return 0, nil
	}
	if delta > 0 {
		return delta, model.DeltaUpdateUserQuotaTx(tx, userID, delta)
	}
	var user model.User
	if err := tx.Set("gorm:query_option", "FOR UPDATE").Select("id", "quota").Where("id = ?", userID).First(&user).Error; err != nil {
		return 0, err
	}
	deduct := -delta
	if int64(user.Quota) < deduct {
		deduct = int64(user.Quota)
	}
	if deduct <= 0 {
		return 0, nil
	}
	return -deduct, model.DeltaUpdateUserQuotaTx(tx, userID, -deduct)
}

func cancelSubscriptionForRefundTx(tx *gorm.DB, item *PaymentOrderItem) (int, string, error) {
	var order model.SubscriptionOrder
	if err := tx.Select("user_id", "plan_id", "complete_time").Where("id = ?", item.ID).First(&order).Error; err != nil {
		return 0, "", err
	}
	_, cacheGroup, err := model.CancelOrderUserSubscriptionForRefundTx(tx, order.UserId, order.PlanId, order.CompleteTime)
	if err != nil {
		return 0, "", err
	}
	if cacheGroup == "" {
		return 0, "", nil
	}
	return order.UserId, cacheGroup, nil
}

func GetPaymentDashboard(days int) (*PaymentDashboardStats, error) {
	if days <= 0 {
		days = 30
	}
	if days > 365 {
		days = 365
	}
	orders, err := listPaymentDashboardOrders()
	if err != nil {
		return nil, err
	}
	start := time.Now().AddDate(0, 0, -days+1)
	startUnix := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.Local).Unix()
	stats := &PaymentDashboardStats{}
	methods := map[string]PaymentDashboardMethod{}
	users := map[int]PaymentDashboardUser{}
	daily := map[string]PaymentDashboardDailyItem{}
	for _, order := range orders {
		switch order.Status {
		case UnifiedOrderStatusCompleted:
			if order.CompleteTime < startUnix {
				continue
			}
			stats.TotalOrders++
			stats.CompletedOrders++
			stats.TotalAmount += order.PayAmount
			method := methods[order.PaymentType]
			method.Type = order.PaymentType
			method.Count++
			method.Amount += order.PayAmount
			methods[order.PaymentType] = method
			user := users[order.UserID]
			user.UserID = order.UserID
			user.Count++
			user.Amount += order.PayAmount
			users[order.UserID] = user
			date := time.Unix(order.CompleteTime, 0).Format("2006-01-02")
			day := daily[date]
			day.Date = date
			day.Count++
			day.Amount += order.PayAmount
			daily[date] = day
		case UnifiedOrderStatusPending:
			stats.TotalOrders++
			stats.PendingOrders++
		case UnifiedOrderStatusFailed:
			stats.TotalOrders++
			stats.FailedOrders++
		case UnifiedOrderStatusRefunded:
			stats.TotalOrders++
			stats.RefundedOrders++
		}
	}
	if stats.TotalOrders > 0 {
		stats.CompletionRate = float64(stats.CompletedOrders) / float64(stats.TotalOrders)
	}
	for _, method := range methods {
		stats.PaymentMethods = append(stats.PaymentMethods, method)
	}
	sort.Slice(stats.PaymentMethods, func(i, j int) bool { return stats.PaymentMethods[i].Amount > stats.PaymentMethods[j].Amount })
	for _, user := range users {
		stats.TopUsers = append(stats.TopUsers, user)
	}
	sort.Slice(stats.TopUsers, func(i, j int) bool { return stats.TopUsers[i].Amount > stats.TopUsers[j].Amount })
	if len(stats.TopUsers) > 10 {
		stats.TopUsers = stats.TopUsers[:10]
	}
	for i := days - 1; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		item := daily[date]
		item.Date = date
		stats.DailySeries = append(stats.DailySeries, item)
	}
	return stats, nil
}

func listPaymentDashboardOrders() ([]paymentDashboardRawOrder, error) {
	items := make([]paymentDashboardRawOrder, 0)
	topUps, err := listPaymentDashboardTopUpOrders()
	if err != nil {
		return nil, err
	}
	items = append(items, topUps...)
	subscriptions, err := listPaymentDashboardSubscriptionOrders()
	if err != nil {
		return nil, err
	}
	items = append(items, subscriptions...)
	return items, nil
}

func listPaymentDashboardTopUpOrders() ([]paymentDashboardRawOrder, error) {
	var topUps []model.TopUp
	query := model.DB.Model(&model.TopUp{}).
		Where("trade_no NOT IN (?)", model.DB.Model(&model.SubscriptionOrder{}).Select("trade_no"))
	if err := query.Find(&topUps).Error; err != nil {
		return nil, err
	}
	items := make([]paymentDashboardRawOrder, 0, len(topUps))
	for _, topUp := range topUps {
		items = append(items, paymentDashboardRawOrder{
			UserID:       topUp.UserId,
			PaymentType:  firstNonEmpty(topUp.PaymentMethod, topUp.PaymentProvider),
			Status:       normalizePaymentOrderStatus(topUp.Status),
			PayAmount:    topUp.Money,
			CreatedTime:  topUp.CreateTime,
			CompleteTime: topUp.CompleteTime,
		})
	}
	return items, nil
}

func listPaymentDashboardSubscriptionOrders() ([]paymentDashboardRawOrder, error) {
	var orders []model.SubscriptionOrder
	if err := model.DB.Model(&model.SubscriptionOrder{}).Find(&orders).Error; err != nil {
		return nil, err
	}
	items := make([]paymentDashboardRawOrder, 0, len(orders))
	for _, order := range orders {
		items = append(items, paymentDashboardRawOrder{
			UserID:       order.UserId,
			PaymentType:  firstNonEmpty(order.PaymentMethod, order.PaymentProvider),
			Status:       normalizePaymentOrderStatus(order.Status),
			PayAmount:    order.Money,
			CreatedTime:  order.CreateTime,
			CompleteTime: order.CompleteTime,
		})
	}
	return items, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return "unknown"
}

func normalizePaymentOrderListParams(params *PaymentOrderListParams) {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = common.ItemsPerPage
	}
	if params.PageSize > 100 {
		params.PageSize = 100
	}
	params.Status = strings.ToUpper(strings.TrimSpace(params.Status))
	params.OrderType = strings.TrimSpace(params.OrderType)
	params.PaymentType = strings.TrimSpace(params.PaymentType)
	params.Keyword = strings.TrimSpace(params.Keyword)
}

func listTopUpOrders(params PaymentOrderListParams) ([]PaymentOrderItem, error) {
	var topUps []model.TopUp
	query := model.DB.Model(&model.TopUp{})
	if params.UserID > 0 {
		query = query.Where("user_id = ?", params.UserID)
	}
	if params.PaymentType != "" {
		query = query.Where("payment_method = ? OR payment_provider = ?", params.PaymentType, params.PaymentType)
	}
	if params.Keyword != "" {
		query = query.Where("trade_no LIKE ?", "%"+params.Keyword+"%")
	}
	query = query.Where("trade_no NOT IN (?)", model.DB.Model(&model.SubscriptionOrder{}).Select("trade_no"))
	if err := query.Find(&topUps).Error; err != nil {
		return nil, err
	}
	items := make([]PaymentOrderItem, 0, len(topUps))
	for _, topUp := range topUps {
		items = append(items, topUpOrderItem(topUp))
	}
	return items, nil
}

func listSubscriptionOrders(params PaymentOrderListParams) ([]PaymentOrderItem, error) {
	var orders []model.SubscriptionOrder
	query := model.DB.Model(&model.SubscriptionOrder{})
	if params.UserID > 0 {
		query = query.Where("user_id = ?", params.UserID)
	}
	if params.PaymentType != "" {
		query = query.Where("payment_method = ? OR payment_provider = ?", params.PaymentType, params.PaymentType)
	}
	if params.Keyword != "" {
		query = query.Where("trade_no LIKE ?", "%"+params.Keyword+"%")
	}
	if err := query.Find(&orders).Error; err != nil {
		return nil, err
	}
	planMeta := subscriptionPlanOrderMetaMap(orders)
	items := make([]PaymentOrderItem, 0, len(orders))
	for _, order := range orders {
		items = append(items, subscriptionOrderItem(order, planMeta[order.PlanId]))
	}
	return items, nil
}

func topUpOrderItem(topUp model.TopUp) PaymentOrderItem {
	return PaymentOrderItem{
		Source:          model.PaymentOrderSourceTopUp,
		OrderType:       model.PaymentOrderTypeBalance,
		ID:              topUp.Id,
		UserID:          topUp.UserId,
		TradeNo:         topUp.TradeNo,
		PaymentMethod:   topUp.PaymentMethod,
		PaymentProvider: topUp.PaymentProvider,
		Status:          normalizePaymentOrderStatus(topUp.Status),
		Amount:          topUp.Amount,
		Money:           topUp.Money,
		PayAmount:       topUp.Money,
		CreatedTime:     topUp.CreateTime,
		CompleteTime:    topUp.CompleteTime,
	}
}

func subscriptionOrderItem(order model.SubscriptionOrder, planMeta subscriptionPlanOrderMeta) PaymentOrderItem {
	return PaymentOrderItem{
		Source:          model.PaymentOrderSourceSubscription,
		OrderType:       model.PaymentOrderTypeSubscription,
		ID:              order.Id,
		UserID:          order.UserId,
		TradeNo:         order.TradeNo,
		PaymentMethod:   order.PaymentMethod,
		PaymentProvider: order.PaymentProvider,
		Status:          normalizePaymentOrderStatus(order.Status),
		Money:           order.Money,
		PayAmount:       order.Money,
		PayCurrency:     planMeta.Currency,
		PlanID:          order.PlanId,
		PlanTitle:       planMeta.Title,
		CreatedTime:     order.CreateTime,
		CompleteTime:    order.CompleteTime,
	}
}

func subscriptionPlanOrderMetaMap(orders []model.SubscriptionOrder) map[int]subscriptionPlanOrderMeta {
	ids := make([]int, 0)
	seen := map[int]struct{}{}
	for _, order := range orders {
		if order.PlanId <= 0 {
			continue
		}
		if _, ok := seen[order.PlanId]; ok {
			continue
		}
		seen[order.PlanId] = struct{}{}
		ids = append(ids, order.PlanId)
	}
	if len(ids) == 0 {
		return map[int]subscriptionPlanOrderMeta{}
	}
	var plans []model.SubscriptionPlan
	if err := model.DB.Select("id", "title", "currency").Where("id IN ?", ids).Find(&plans).Error; err != nil {
		return map[int]subscriptionPlanOrderMeta{}
	}
	meta := make(map[int]subscriptionPlanOrderMeta, len(plans))
	for _, plan := range plans {
		meta[plan.Id] = subscriptionPlanOrderMeta{
			Title:    plan.Title,
			Currency: strings.ToUpper(strings.TrimSpace(plan.Currency)),
		}
	}
	return meta
}

func overlayPaymentOrderRefunds(items []PaymentOrderItem) error {
	var refunds []model.PaymentOrderRefund
	if err := model.DB.Where("status <> ?", "").Find(&refunds).Error; err != nil {
		return err
	}
	refundByOrder := make(map[string]model.PaymentOrderRefund, len(refunds))
	for _, refund := range refunds {
		key := paymentOrderKey(refund.Source, refund.SourceOrderId)
		current, ok := refundByOrder[key]
		if !ok || refund.UpdateTime > current.UpdateTime {
			refundByOrder[key] = refund
		}
	}
	for i := range items {
		if refund, ok := refundByOrder[paymentOrderKey(items[i].Source, items[i].ID)]; ok {
			items[i].RefundAmount = refund.Amount
			items[i].RefundReason = refund.Reason
			items[i].RefundRequestReason = refund.Reason
			switch refund.Status {
			case model.PaymentRefundStatusRequested:
				items[i].Status = UnifiedOrderStatusRefundRequested
			case model.PaymentRefundStatusProcessed:
				items[i].Status = UnifiedOrderStatusRefunded
			case model.PaymentRefundStatusFailed:
				items[i].Status = UnifiedOrderStatusRefundFailed
			}
		}
	}
	return nil
}

func filterPaymentOrderItems(items []PaymentOrderItem, params PaymentOrderListParams) []PaymentOrderItem {
	if params.Status == "" {
		return items
	}
	filtered := make([]PaymentOrderItem, 0, len(items))
	for _, item := range items {
		if item.Status == params.Status {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func normalizePaymentOrderStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case common.TopUpStatusPending:
		return UnifiedOrderStatusPending
	case common.TopUpStatusSuccess:
		return UnifiedOrderStatusCompleted
	case common.TopUpStatusExpired:
		return UnifiedOrderStatusExpired
	case "cancelled", "canceled":
		return UnifiedOrderStatusCancelled
	case common.TopUpStatusFailed:
		return UnifiedOrderStatusFailed
	default:
		if status == "" {
			return UnifiedOrderStatusPending
		}
		return strings.ToUpper(status)
	}
}

func RequestPaymentOrderRefund(ref PaymentOrderRef, userID int, reason string) error {
	if strings.TrimSpace(reason) == "" {
		return errors.New("退款原因不能为空")
	}
	item, err := getPaymentOrderItem(ref)
	if err != nil {
		return err
	}
	if userID > 0 && item.UserID != userID {
		return errors.New("订单不存在")
	}
	if item.Status != UnifiedOrderStatusCompleted {
		return errors.New("只有已完成订单可以申请退款")
	}
	refund := model.PaymentOrderRefund{
		Source:             item.Source,
		OrderType:          item.OrderType,
		SourceOrderId:      item.ID,
		SourceOrderTradeNo: item.TradeNo,
		UserId:             item.UserID,
		Amount:             item.PayAmount,
		Reason:             strings.TrimSpace(reason),
		Status:             model.PaymentRefundStatusRequested,
		RequestedBy:        "user",
	}
	return model.DB.Transaction(func(tx *gorm.DB) error {
		var existing model.PaymentOrderRefund
		err := tx.Where("source = ? AND source_order_id = ? AND status IN ?", item.Source, item.ID, []string{model.PaymentRefundStatusRequested, model.PaymentRefundStatusProcessed}).First(&existing).Error
		if err == nil {
			return errors.New("订单已存在退款记录")
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err := tx.Create(&refund).Error; err != nil {
			return err
		}
		return tx.Create(&model.PaymentOrderAuditLog{
			Source:             item.Source,
			OrderType:          item.OrderType,
			SourceOrderId:      item.ID,
			SourceOrderTradeNo: item.TradeNo,
			UserId:             item.UserID,
			Action:             "refund_requested",
			Detail:             refund.Reason,
			Operator:           "user",
		}).Error
	})
}

func GetPaymentActivityConfig(activityType string, defaultConfig any) (bool, any, error) {
	activityType = strings.TrimSpace(activityType)
	if activityType == "" {
		return false, defaultConfig, errors.New("活动类型为空")
	}
	var cfg model.PaymentActivityConfig
	if err := model.DB.Where("activity_type = ?", activityType).First(&cfg).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, defaultConfig, nil
		}
		return false, defaultConfig, err
	}
	if strings.TrimSpace(cfg.Config) == "" {
		return cfg.Enabled, defaultConfig, nil
	}
	if err := common.UnmarshalJsonStr(cfg.Config, &defaultConfig); err != nil {
		return cfg.Enabled, defaultConfig, err
	}
	return cfg.Enabled, defaultConfig, nil
}

func UpdatePaymentActivityConfig(activityType string, enabled bool, config any) error {
	payload, err := common.Marshal(config)
	if err != nil {
		return err
	}
	var cfg model.PaymentActivityConfig
	err = model.DB.Where("activity_type = ?", activityType).First(&cfg).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		cfg = model.PaymentActivityConfig{ActivityType: activityType, Enabled: enabled, Config: string(payload)}
		return model.DB.Create(&cfg).Error
	}
	if err != nil {
		return err
	}
	cfg.Enabled = enabled
	cfg.Config = string(payload)
	return model.DB.Save(&cfg).Error
}

func GetPaymentActivityStats(activityType string) (map[string]any, error) {
	var total int64
	var pending int64
	var used int64
	query := model.DB.Model(&model.PaymentActivityChance{}).Where("activity_type = ?", activityType)
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	if err := model.DB.Model(&model.PaymentActivityChance{}).Where("activity_type = ? AND status = ?", activityType, "pending").Count(&pending).Error; err != nil {
		return nil, err
	}
	if err := model.DB.Model(&model.PaymentActivityChance{}).Where("activity_type = ? AND used_chances > 0", activityType).Count(&used).Error; err != nil {
		return nil, err
	}
	return map[string]any{
		"total_chances":        total,
		"pending_chances":      pending,
		"drawn_chances":        used,
		"pending_fulfillments": pending,
	}, nil
}

func getPaymentOrderItem(ref PaymentOrderRef) (*PaymentOrderItem, error) {
	if ref.ID <= 0 {
		return nil, errors.New("订单不存在")
	}
	if ref.OrderType == "" || ref.OrderType == model.PaymentOrderTypeBalance {
		var topUp model.TopUp
		err := model.DB.Where("id = ?", ref.ID).First(&topUp).Error
		if err == nil {
			var subscriptionCount int64
			if countErr := model.DB.Model(&model.SubscriptionOrder{}).Where("trade_no = ?", topUp.TradeNo).Count(&subscriptionCount).Error; countErr != nil {
				return nil, countErr
			}
			if subscriptionCount == 0 {
				item := topUpOrderItem(topUp)
				items := []PaymentOrderItem{item}
				if err := overlayPaymentOrderRefunds(items); err != nil {
					return nil, err
				}
				return &items[0], nil
			}
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}
	if ref.OrderType == "" || ref.OrderType == model.PaymentOrderTypeSubscription {
		var order model.SubscriptionOrder
		err := model.DB.Where("id = ?", ref.ID).First(&order).Error
		if err == nil {
			planMeta := subscriptionPlanOrderMeta{}
			if order.PlanId > 0 {
				var plan model.SubscriptionPlan
				if planErr := model.DB.Select("id", "title", "currency").Where("id = ?", order.PlanId).First(&plan).Error; planErr == nil {
					planMeta = subscriptionPlanOrderMeta{
						Title:    plan.Title,
						Currency: strings.ToUpper(strings.TrimSpace(plan.Currency)),
					}
				} else if !errors.Is(planErr, gorm.ErrRecordNotFound) {
					return nil, planErr
				}
			}
			item := subscriptionOrderItem(order, planMeta)
			items := []PaymentOrderItem{item}
			if err := overlayPaymentOrderRefunds(items); err != nil {
				return nil, err
			}
			return &items[0], nil
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}
	return nil, errors.New("订单不存在")
}

func GrantPaymentActivityChanceForOrder(order *PaymentOrderItem, activityType string, tierID string, chances int) error {
	if order == nil {
		return errors.New("订单为空")
	}
	if order.Status != UnifiedOrderStatusCompleted {
		return nil
	}
	if strings.TrimSpace(activityType) == "" {
		return errors.New("活动类型为空")
	}
	if strings.TrimSpace(tierID) == "" {
		tierID = "default"
	}
	chance := model.PaymentActivityChance{
		ActivityType:       activityType,
		TierId:             tierID,
		UserId:             order.UserID,
		Source:             order.Source,
		OrderType:          order.OrderType,
		SourceOrderId:      order.ID,
		SourceOrderTradeNo: order.TradeNo,
		PayAmount:          order.PayAmount,
		Chances:            chances,
		Status:             "pending",
	}
	return model.DB.Where("activity_type = ? AND source_order_trade_no = ? AND tier_id = ?", activityType, order.TradeNo, tierID).
		FirstOrCreate(&chance).Error
}

func GetLuckyWheelConfig() (bool, model.LuckyWheelConfig, error) {
	return model.GetLuckyWheelConfig()
}

func UpdateLuckyWheelConfig(enabled bool, cfg model.LuckyWheelConfig) error {
	return model.UpdateLuckyWheelConfig(enabled, cfg)
}

func GetLuckyWheelSummary(userID int) (*model.LuckyWheelSummary, error) {
	return model.GetLuckyWheelSummary(userID)
}

func DrawLuckyWheel(userID int, sessionID int) (*model.LuckyWheelDrawResult, error) {
	return model.DrawLuckyWheel(userID, sessionID)
}

func GetLuckyWheelStats() (*model.LuckyWheelStats, error) {
	return model.GetLuckyWheelStats()
}

func GetRechargeActivityConfig() (bool, model.RechargeActivityConfig, error) {
	return model.GetRechargeActivityConfig()
}

func UpdateRechargeActivityConfig(enabled bool, cfg model.RechargeActivityConfig) error {
	return model.UpdateRechargeActivityConfig(enabled, cfg)
}

func GetRechargeActivitySummary(userID int) (*model.RechargeActivitySummary, error) {
	return model.GetRechargeActivitySummary(userID)
}

func DrawRechargeActivity(userID int, chanceID int) (*model.RechargeActivityDrawResult, error) {
	return model.DrawRechargeActivity(userID, chanceID)
}

func GetRechargeActivityStats(page int, pageSize int, keyword string) (*model.RechargeActivityStats, error) {
	return model.GetRechargeActivityStats(page, pageSize, keyword)
}

func UpdateRechargeActivityRecordFulfillment(id int, status string, note string, adminUserID int) (*model.RechargeActivityDrawRecord, error) {
	return model.UpdateRechargeActivityRecordFulfillment(id, status, note, adminUserID)
}

func paymentOrderKey(source string, id int) string {
	return source + ":" + strconv.Itoa(id)
}

func createPaymentOrderAuditTx(tx *gorm.DB, item *PaymentOrderItem, action string, detail string, operator string) error {
	return tx.Create(&model.PaymentOrderAuditLog{
		Source:             item.Source,
		OrderType:          item.OrderType,
		SourceOrderId:      item.ID,
		SourceOrderTradeNo: item.TradeNo,
		UserId:             item.UserID,
		Action:             action,
		Detail:             detail,
		Operator:           operator,
	}).Error
}
