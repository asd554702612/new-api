package controller

import (
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
)

type paymentRefundRequest struct {
	Reason string `json:"reason"`
}

type adminPaymentRefundRequest struct {
	Amount        float64 `json:"amount"`
	Reason        string  `json:"reason"`
	Force         bool    `json:"force"`
	DeductBalance bool    `json:"deduct_balance"`
}

type paymentActivityConfigRequest struct {
	Enabled bool           `json:"enabled"`
	Config  map[string]any `json:"config"`
}

type luckyWheelConfigRequest struct {
	Enabled bool                   `json:"enabled"`
	Config  model.LuckyWheelConfig `json:"config"`
}

type rechargeActivityConfigRequest struct {
	Enabled bool                         `json:"enabled"`
	Config  model.RechargeActivityConfig `json:"config"`
}

type luckyWheelDrawRequest struct {
	SessionId int `json:"session_id"`
}

type rechargeActivityDrawRequest struct {
	ChanceId int `json:"chance_id"`
}

type rechargeActivityFulfillmentRequest struct {
	Status string `json:"status"`
	Note   string `json:"note"`
}

func GetMyPaymentOrders(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	orders, total, err := service.ListPaymentOrders(service.PaymentOrderListParams{
		UserID:      c.GetInt("id"),
		Page:        pageInfo.GetPage(),
		PageSize:    pageInfo.GetPageSize(),
		Status:      c.Query("status"),
		OrderType:   c.Query("order_type"),
		PaymentType: c.Query("payment_type"),
		Keyword:     c.Query("keyword"),
	})
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(total)
	pageInfo.SetItems(orders)
	common.ApiSuccess(c, pageInfo)
}

func GetMyPaymentOrderDetail(c *gin.Context) {
	ref, ok := parsePaymentOrderRef(c)
	if !ok {
		return
	}
	detail, err := service.GetPaymentOrderDetail(ref, c.GetInt("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, detail)
}

func CancelMyPaymentOrder(c *gin.Context) {
	ref, ok := parsePaymentOrderRef(c)
	if !ok {
		return
	}
	if err := service.CancelPaymentOrder(ref, c.GetInt("id"), "user"); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, nil)
}

func RequestMyPaymentOrderRefund(c *gin.Context) {
	ref, ok := parsePaymentOrderRef(c)
	if !ok {
		return
	}
	var req paymentRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	if err := service.RequestPaymentOrderRefund(ref, c.GetInt("id"), req.Reason); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, nil)
}

func GetLuckyWheelSummary(c *gin.Context) {
	summary, err := service.GetLuckyWheelSummary(c.GetInt("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, summary)
}

func DrawLuckyWheel(c *gin.Context) {
	var req luckyWheelDrawRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.SessionId <= 0 {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	result, err := service.DrawLuckyWheel(c.GetInt("id"), req.SessionId)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, result)
}

func GetRechargeActivitySummary(c *gin.Context) {
	summary, err := service.GetRechargeActivitySummary(c.GetInt("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, summary)
}

func DrawRechargeActivity(c *gin.Context) {
	var req rechargeActivityDrawRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.ChanceId <= 0 {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	result, err := service.DrawRechargeActivity(c.GetInt("id"), req.ChanceId)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, result)
}

func AdminGetPaymentDashboard(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	stats, err := service.GetPaymentDashboard(days)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, stats)
}

func AdminListPaymentOrders(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	userID, _ := strconv.Atoi(c.Query("user_id"))
	orders, total, err := service.ListPaymentOrders(service.PaymentOrderListParams{
		UserID:      userID,
		Page:        pageInfo.GetPage(),
		PageSize:    pageInfo.GetPageSize(),
		Status:      c.Query("status"),
		OrderType:   c.Query("order_type"),
		PaymentType: c.Query("payment_type"),
		Keyword:     c.Query("keyword"),
	})
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(total)
	pageInfo.SetItems(orders)
	common.ApiSuccess(c, pageInfo)
}

func AdminGetPaymentOrderDetail(c *gin.Context) {
	ref, ok := parsePaymentOrderRef(c)
	if !ok {
		return
	}
	detail, err := service.GetPaymentOrderDetail(ref, 0)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, detail)
}

func AdminCancelPaymentOrder(c *gin.Context) {
	ref, ok := parsePaymentOrderRef(c)
	if !ok {
		return
	}
	if err := service.CancelPaymentOrder(ref, 0, "admin"); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, nil)
}

func AdminRetryPaymentOrder(c *gin.Context) {
	ref, ok := parsePaymentOrderRef(c)
	if !ok {
		return
	}
	if err := service.RetryPaymentOrderFulfillment(ref, c.ClientIP()); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, nil)
}

func AdminRefundPaymentOrder(c *gin.Context) {
	ref, ok := parsePaymentOrderRef(c)
	if !ok {
		return
	}
	var req adminPaymentRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	refund, err := service.ProcessPaymentOrderRefund(ref, req.Amount, req.Reason, req.Force, req.DeductBalance, "admin")
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, refund)
}

func AdminGetPaymentActivityConfig(c *gin.Context) {
	activityType := c.Param("activity_type")
	enabled, cfg, err := service.GetPaymentActivityConfig(activityType, defaultPaymentActivityConfig(activityType))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, gin.H{"enabled": enabled, "config": cfg})
}

func AdminUpdatePaymentActivityConfig(c *gin.Context) {
	activityType := c.Param("activity_type")
	var req paymentActivityConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	if req.Config == nil {
		req.Config = defaultPaymentActivityConfig(activityType)
	}
	if err := service.UpdatePaymentActivityConfig(activityType, req.Enabled, req.Config); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, gin.H{"enabled": req.Enabled, "config": req.Config})
}

func AdminGetPaymentActivityStats(c *gin.Context) {
	stats, err := service.GetPaymentActivityStats(c.Param("activity_type"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, stats)
}

func AdminGetLuckyWheelConfig(c *gin.Context) {
	enabled, cfg, err := service.GetLuckyWheelConfig()
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, gin.H{"enabled": enabled, "config": cfg})
}

func AdminUpdateLuckyWheelConfig(c *gin.Context) {
	var req luckyWheelConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	if err := service.UpdateLuckyWheelConfig(req.Enabled, req.Config); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, gin.H{"enabled": req.Enabled, "config": req.Config})
}

func AdminGetLuckyWheelStats(c *gin.Context) {
	stats, err := service.GetLuckyWheelStats()
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, stats)
}

func AdminGetRechargeActivityConfig(c *gin.Context) {
	enabled, cfg, err := service.GetRechargeActivityConfig()
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, gin.H{"enabled": enabled, "config": cfg})
}

func AdminUpdateRechargeActivityConfig(c *gin.Context) {
	var req rechargeActivityConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	if err := service.UpdateRechargeActivityConfig(req.Enabled, req.Config); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, gin.H{"enabled": req.Enabled, "config": req.Config})
}

func AdminGetRechargeActivityStats(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	stats, err := service.GetRechargeActivityStats(pageInfo.GetPage(), pageInfo.GetPageSize(), c.Query("user_keyword"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, stats)
}

func AdminUpdateRechargeActivityRecordFulfillment(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		common.ApiErrorMsg(c, "记录 ID 错误")
		return
	}
	var req rechargeActivityFulfillmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	record, err := service.UpdateRechargeActivityRecordFulfillment(id, req.Status, req.Note, c.GetInt("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, record)
}

func parsePaymentOrderRef(c *gin.Context) (service.PaymentOrderRef, bool) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		common.ApiErrorMsg(c, "订单 ID 错误")
		return service.PaymentOrderRef{}, false
	}
	return service.PaymentOrderRef{
		OrderType: c.Param("order_type"),
		ID:        id,
	}, true
}

func defaultPaymentActivityConfig(activityType string) map[string]any {
	switch activityType {
	case "lucky_wheel":
		return map[string]any{
			"eligible_order_types": []string{"balance", "subscription"},
			"tiers": []map[string]any{
				{"id": "default", "name": "默认梯度", "min_amount": 0, "chances": 1},
			},
			"prizes": []map[string]any{},
		}
	case "recharge_activity":
		return map[string]any{
			"first_recharge_enabled": false,
			"member_level_enabled":   false,
			"tiers": []map[string]any{
				{"id": "default", "name": "默认梯度", "pay_amount": 0, "bonus_amount": 0},
			},
			"member_levels": []map[string]any{},
		}
	default:
		return map[string]any{}
	}
}
