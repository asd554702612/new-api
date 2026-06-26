package controller

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting"
	"github.com/gin-gonic/gin"
	"github.com/thanhpk/randstr"
)

func RequestCasdoorAmount(c *gin.Context) {
	requestOfficialAmountWithMin(c, officialTopUpUnitPrice(setting.CasdoorPaymentUnitPrice), setting.GetCasdoorPaymentMinTopUp())
}

func RequestCasdoorPay(c *gin.Context) {
	if !isCasdoorTopUpEnabled() {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "Casdoor 支付未启用"})
		return
	}
	var req OfficialPayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "参数错误"})
		return
	}
	id := c.GetInt("id")
	amount, payMoney, err := validateOfficialTopUpAmountWithMin(id, req.Amount, officialTopUpUnitPrice(setting.CasdoorPaymentUnitPrice), setting.GetCasdoorPaymentMinTopUp())
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": err.Error()})
		return
	}
	user, err := model.GetUserById(id, false)
	if err != nil || user == nil {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "用户不存在"})
		return
	}

	tradeNo := fmt.Sprintf("CASDOOR-%d-%d-%s", id, time.Now().UnixMilli(), randstr.String(6))
	topUp := &model.TopUp{
		UserId:          id,
		Amount:          amount,
		Money:           payMoney,
		TradeNo:         tradeNo,
		PaymentMethod:   model.PaymentMethodCasdoor,
		PaymentProvider: model.PaymentProviderCasdoor,
		CreateTime:      time.Now().Unix(),
		Status:          common.TopUpStatusPending,
	}
	if err := topUp.Insert(); err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("Casdoor 创建充值订单失败 user_id=%d trade_no=%s error=%q", id, tradeNo, err.Error()))
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "创建订单失败"})
		return
	}

	result, requestBody, err := service.CreateCasdoorPayment(c.Request.Context(), service.CasdoorPaymentCreateRequest{
		ExternalOrderID: tradeNo,
		UserID:          casdoorUserID(user),
		ProductName:     setting.GetCasdoorPaymentProduct(),
		ProviderName:    setting.GetCasdoorPaymentProvider(),
		Amount:          payMoney,
		Currency:        setting.GetCasdoorPaymentCurrency(),
		DisplayName:     "GPTK 充值",
		Detail:          fmt.Sprintf("Recharge %d credits", req.Amount),
	})
	if err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("Casdoor 拉起支付失败 user_id=%d trade_no=%s amount=%d error=%q", id, tradeNo, req.Amount, err.Error()))
		topUp.Status = common.TopUpStatusFailed
		_ = topUp.Update()
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "拉起支付失败: " + err.Error()})
		return
	}
	if payload, err := common.Marshal(map[string]any{"request": requestBody, "response": result}); err == nil {
		topUp.ProviderPayload = string(payload)
		_ = topUp.Update()
	}
	c.JSON(http.StatusOK, gin.H{"message": "success", "data": casdoorPaymentResponse(tradeNo, result)})
}

func SubscriptionRequestCasdoorPay(c *gin.Context) {
	if !requirePaymentCompliance(c) {
		return
	}
	if !isCasdoorTopUpEnabled() {
		common.ApiErrorMsg(c, "Casdoor 支付未启用")
		return
	}
	var req OfficialSubscriptionPayRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.PlanId <= 0 {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	plan, user, payMoneyCNY, err := prepareOfficialSubscription(c, req.PlanId)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	tradeNo := fmt.Sprintf("CASDOOR_SUB-%d-%d-%s", user.Id, time.Now().UnixMilli(), randstr.String(6))
	order := &model.SubscriptionOrder{
		UserId:          user.Id,
		PlanId:          plan.Id,
		Money:           plan.PriceAmount,
		TradeNo:         tradeNo,
		PaymentMethod:   model.PaymentMethodCasdoor,
		PaymentProvider: model.PaymentProviderCasdoor,
		CreateTime:      time.Now().Unix(),
		Status:          common.TopUpStatusPending,
	}
	if err := order.Insert(); err != nil {
		common.ApiErrorMsg(c, "创建订单失败")
		return
	}
	result, requestBody, err := service.CreateCasdoorPayment(c.Request.Context(), service.CasdoorPaymentCreateRequest{
		ExternalOrderID: tradeNo,
		UserID:          casdoorUserID(user),
		ProductName:     setting.GetCasdoorPaymentProduct(),
		ProviderName:    setting.GetCasdoorPaymentProvider(),
		Amount:          payMoneyCNY,
		Currency:        setting.GetCasdoorPaymentCurrency(),
		DisplayName:     "Subscription: " + plan.Title,
		Detail:          fmt.Sprintf("Subscription plan %d", plan.Id),
	})
	if err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("Casdoor 订阅拉起支付失败 user_id=%d plan_id=%d trade_no=%s error=%q", user.Id, plan.Id, tradeNo, err.Error()))
		order.Status = common.TopUpStatusFailed
		_ = order.Update()
		common.ApiErrorMsg(c, "拉起支付失败: "+err.Error())
		return
	}
	if payload, err := common.Marshal(map[string]any{"request": requestBody, "response": result}); err == nil {
		order.ProviderPayload = string(payload)
		_ = order.Update()
	}
	common.ApiSuccess(c, casdoorPaymentResponse(tradeNo, result))
}

func CasdoorPaymentWebhook(c *gin.Context) {
	if !isCasdoorWebhookEnabled() {
		logger.LogWarn(c.Request.Context(), fmt.Sprintf("Casdoor webhook 被拒绝 reason=webhook_disabled path=%q client_ip=%s", c.Request.RequestURI, c.ClientIP()))
		c.JSON(http.StatusForbidden, gin.H{"message": "webhook disabled"})
		return
	}
	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid body"})
		return
	}
	signature := c.GetHeader("X-Casdoor-Webhook-Signature")
	if !service.VerifyCasdoorWebhookSignature(setting.CasdoorClientSecret, rawBody, signature) {
		logger.LogWarn(c.Request.Context(), fmt.Sprintf("Casdoor webhook 验签失败 client_ip=%s", c.ClientIP()))
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid signature"})
		return
	}
	event, err := service.ParseCasdoorWebhookEvent(rawBody)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid payload"})
		return
	}
	if err := validateCasdoorWebhookEvent(c, event); err != nil {
		logger.LogWarn(c.Request.Context(), fmt.Sprintf("Casdoor webhook 校验失败 trade_no=%s error=%q", event.ExternalOrderID, err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	LockOrder(event.ExternalOrderID)
	defer UnlockOrder(event.ExternalOrderID)
	if err := completeCasdoorPayment(c, event, string(rawBody)); err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("Casdoor webhook 完成订单失败 trade_no=%s error=%q", event.ExternalOrderID, err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "process failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "success"})
}

func casdoorUserID(user *model.User) string {
	username := strings.TrimSpace(user.Username)
	if username == "" {
		username = fmt.Sprintf("user-%d", user.Id)
	}
	return "gepin/" + username
}

func casdoorPaymentResponse(tradeNo string, result *service.CasdoorPaymentCreateResult) gin.H {
	if result == nil {
		return gin.H{"trade_no": tradeNo}
	}
	return gin.H{
		"trade_no":          tradeNo,
		"order_id":          result.OrderID,
		"payment_id":        result.PaymentID,
		"external_order_id": result.ExternalOrderID,
		"payUrl":            result.PayURL,
		"pay_url":           result.PayURL,
		"state":             result.State,
		"amount":            result.Amount,
		"currency":          result.Currency,
		"providerName":      result.ProviderName,
		"provider_name":     result.ProviderName,
	}
}

func validateCasdoorWebhookEvent(c *gin.Context, event *service.CasdoorPaymentWebhookEvent) error {
	if event == nil {
		return errors.New("事件为空")
	}
	headerEvent := strings.TrimSpace(c.GetHeader("X-Casdoor-Webhook-Event"))
	if headerEvent != "" && headerEvent != "payment.paid" {
		return errors.New("事件类型错误")
	}
	if event.Event != "payment.paid" {
		return errors.New("事件类型错误")
	}
	if strings.TrimSpace(event.Application) != strings.TrimSpace(setting.CasdoorApplicationName) {
		return errors.New("应用不匹配")
	}
	if strings.TrimSpace(event.ExternalOrderID) == "" {
		return errors.New("订单号为空")
	}
	if service.MoneyToCNYCents(event.Amount) <= 0 {
		return errors.New("支付金额无效")
	}
	if strings.ToUpper(strings.TrimSpace(event.Currency)) != setting.GetCasdoorPaymentCurrency() {
		return errors.New("币种不匹配")
	}
	if strings.TrimSpace(event.ProviderName) != setting.GetCasdoorPaymentProvider() {
		return errors.New("支付渠道不匹配")
	}
	return nil
}

func completeCasdoorPayment(c *gin.Context, event *service.CasdoorPaymentWebhookEvent, rawPayload string) error {
	tradeNo := event.ExternalOrderID
	paidCents := service.MoneyToCNYCents(event.Amount)
	if topUp := model.GetTopUpByTradeNo(tradeNo); topUp != nil {
		if topUp.PaymentProvider != model.PaymentProviderCasdoor {
			return model.ErrPaymentMethodMismatch
		}
		if service.MoneyToCNYCents(topUp.Money) != paidCents {
			return errors.New("支付金额不匹配")
		}
		return model.RechargeCasdoor(tradeNo, rawPayload, c.ClientIP())
	}

	order := model.GetSubscriptionOrderByTradeNo(tradeNo)
	if order == nil {
		return errors.New("订单不存在")
	}
	if order.PaymentProvider != model.PaymentProviderCasdoor {
		return model.ErrPaymentMethodMismatch
	}
	expectedCNY, err := expectedSubscriptionPayMoneyCNY(order)
	if err != nil {
		return err
	}
	if service.MoneyToCNYCents(expectedCNY) != paidCents {
		return errors.New("支付金额不匹配")
	}
	return model.CompleteSubscriptionOrder(tradeNo, rawPayload, model.PaymentProviderCasdoor, order.PaymentMethod)
}
