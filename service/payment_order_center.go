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

func ProcessPaymentOrderRefund(ref PaymentOrderRef, amount float64, reason string, force bool, deductBalance bool, operator string) error {
	if amount <= 0 {
		return errors.New("退款金额必须大于 0")
	}
	if strings.TrimSpace(reason) == "" {
		return errors.New("退款原因不能为空")
	}
	item, err := getPaymentOrderItem(ref)
	if err != nil {
		return err
	}
	if item.Status != UnifiedOrderStatusCompleted && item.Status != UnifiedOrderStatusRefundRequested {
		return errors.New("当前订单不支持退款")
	}
	if amount-item.PayAmount > 0.000001 && !force {
		return errors.New("退款金额不能大于订单支付金额")
	}
	return model.DB.Transaction(func(tx *gorm.DB) error {
		refund := model.PaymentOrderRefund{
			Source:             item.Source,
			OrderType:          item.OrderType,
			SourceOrderId:      item.ID,
			SourceOrderTradeNo: item.TradeNo,
			UserId:             item.UserID,
			Amount:             amount,
			Reason:             strings.TrimSpace(reason),
			Status:             model.PaymentRefundStatusProcessed,
			RequestedBy:        "admin",
			Force:              force,
			DeductBalance:      deductBalance,
			ProcessTime:        common.GetTimestamp(),
		}
		if err := tx.Create(&refund).Error; err != nil {
			return err
		}
		if deductBalance {
			quotaToDeduct := int(math.Round(amount * common.QuotaPerUnit))
			if quotaToDeduct > 0 {
				if err := tx.Model(&model.User{}).Where("id = ?", item.UserID).Update("quota", gorm.Expr("CASE WHEN quota >= ? THEN quota - ? ELSE 0 END", quotaToDeduct, quotaToDeduct)).Error; err != nil {
					return err
				}
			}
		}
		return createPaymentOrderAuditTx(tx, item, "refund_processed", fmt.Sprintf("退款 %.2f：%s", amount, reason), operator)
	})
}

func GetPaymentDashboard(days int) (*PaymentDashboardStats, error) {
	if days <= 0 {
		days = 30
	}
	if days > 365 {
		days = 365
	}
	orders, _, err := ListPaymentOrders(PaymentOrderListParams{Page: 1, PageSize: 10000})
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
		if order.CreatedTime < startUnix {
			continue
		}
		stats.TotalOrders++
		switch order.Status {
		case UnifiedOrderStatusCompleted:
			stats.CompletedOrders++
			stats.TotalAmount += order.PayAmount
			method := methods[order.PaymentMethod]
			method.Type = order.PaymentMethod
			method.Count++
			method.Amount += order.PayAmount
			methods[order.PaymentMethod] = method
			user := users[order.UserID]
			user.UserID = order.UserID
			user.Count++
			user.Amount += order.PayAmount
			users[order.UserID] = user
			date := time.Unix(order.CreatedTime, 0).Format("2006-01-02")
			day := daily[date]
			day.Date = date
			day.Count++
			day.Amount += order.PayAmount
			daily[date] = day
		case UnifiedOrderStatusPending:
			stats.PendingOrders++
		case UnifiedOrderStatusFailed:
			stats.FailedOrders++
		case UnifiedOrderStatusRefunded:
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
	planTitles := subscriptionPlanTitles(orders)
	items := make([]PaymentOrderItem, 0, len(orders))
	for _, order := range orders {
		items = append(items, subscriptionOrderItem(order, planTitles[order.PlanId]))
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

func subscriptionOrderItem(order model.SubscriptionOrder, planTitle string) PaymentOrderItem {
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
		PlanID:          order.PlanId,
		PlanTitle:       planTitle,
		CreatedTime:     order.CreateTime,
		CompleteTime:    order.CompleteTime,
	}
}

func subscriptionPlanTitles(orders []model.SubscriptionOrder) map[int]string {
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
		return map[int]string{}
	}
	var plans []model.SubscriptionPlan
	if err := model.DB.Select("id", "title").Where("id IN ?", ids).Find(&plans).Error; err != nil {
		return map[int]string{}
	}
	titles := make(map[int]string, len(plans))
	for _, plan := range plans {
		titles[plan.Id] = plan.Title
	}
	return titles
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
			planTitle := ""
			if order.PlanId > 0 {
				var plan model.SubscriptionPlan
				if planErr := model.DB.Select("id", "title").Where("id = ?", order.PlanId).First(&plan).Error; planErr == nil {
					planTitle = plan.Title
				} else if !errors.Is(planErr, gorm.ErrRecordNotFound) {
					return nil, planErr
				}
			}
			item := subscriptionOrderItem(order, planTitle)
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
