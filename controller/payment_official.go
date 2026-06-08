package controller

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/thanhpk/randstr"
)

type OfficialPayRequest struct {
	Amount    int64  `json:"amount"`
	TradeType string `json:"trade_type"`
}

type OfficialSubscriptionPayRequest struct {
	PlanId    int    `json:"plan_id"`
	TradeType string `json:"trade_type"`
}

func RequestWechatPayAmount(c *gin.Context) {
	requestOfficialAmount(c, officialTopUpUnitPrice(setting.WechatPayUnitPrice))
}

func RequestAlipayAmount(c *gin.Context) {
	requestOfficialAmount(c, officialTopUpUnitPrice(setting.AlipayUnitPrice))
}

func RequestWechatPay(c *gin.Context) {
	if !isWechatPayTopUpEnabled() {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "微信支付未启用"})
		return
	}
	var req OfficialPayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "参数错误"})
		return
	}
	tradeType := service.NormalizeWechatPayTradeType(req.TradeType)
	if !isWechatPayTradeTypeEnabled(tradeType) {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "不支持的微信支付方式"})
		return
	}
	id := c.GetInt("id")
	amount, payMoney, err := validateOfficialTopUpAmount(id, req.Amount, officialTopUpUnitPrice(setting.WechatPayUnitPrice))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": err.Error()})
		return
	}
	user, err := model.GetUserById(id, false)
	if err != nil || user == nil {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "用户不存在"})
		return
	}

	tradeNo := fmt.Sprintf("WXPAY-%d-%d-%s", id, time.Now().UnixMilli(), randstr.String(6))
	topUp := &model.TopUp{
		UserId:          id,
		Amount:          amount,
		Money:           payMoney,
		TradeNo:         tradeNo,
		PaymentMethod:   model.PaymentMethodWechatPay,
		PaymentProvider: model.PaymentProviderWechatPay,
		CreateTime:      time.Now().Unix(),
		Status:          common.TopUpStatusPending,
	}
	if err := topUp.Insert(); err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("微信支付 创建充值订单失败 user_id=%d trade_no=%s error=%q", id, tradeNo, err.Error()))
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "创建订单失败"})
		return
	}

	notifyURL := officialNotifyURL(setting.WechatPayNotifyURL, "/api/wechat-pay/notify")
	resp, err := service.CreateWechatPayOrder(c.Request.Context(), service.WechatPayOrderRequest{
		TradeType:   tradeType,
		Description: fmt.Sprintf("Recharge %d credits", req.Amount),
		OutTradeNo:  tradeNo,
		AmountCents: service.MoneyToCNYCents(payMoney),
		NotifyURL:   notifyURL,
		ClientIP:    c.ClientIP(),
		OpenID:      user.WeChatId,
		RedirectURL: paymentReturnPath("/wallet?show_history=true"),
	})
	if err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("微信支付 拉起支付失败 user_id=%d trade_no=%s trade_type=%s error=%q", id, tradeNo, tradeType, err.Error()))
		topUp.Status = common.TopUpStatusFailed
		_ = topUp.Update()
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "拉起支付失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "success", "data": resp})
}

func RequestAlipay(c *gin.Context) {
	if !isAlipayTopUpEnabled() {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "支付宝支付未启用"})
		return
	}
	var req OfficialPayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "参数错误"})
		return
	}
	tradeType := service.NormalizeAlipayTradeType(req.TradeType)
	if !isAlipayTradeTypeEnabled(tradeType) {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "不支持的支付宝支付方式"})
		return
	}
	id := c.GetInt("id")
	amount, payMoney, err := validateOfficialTopUpAmount(id, req.Amount, officialTopUpUnitPrice(setting.AlipayUnitPrice))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": err.Error()})
		return
	}

	tradeNo := fmt.Sprintf("ALIPAY-%d-%d-%s", id, time.Now().UnixMilli(), randstr.String(6))
	topUp := &model.TopUp{
		UserId:          id,
		Amount:          amount,
		Money:           payMoney,
		TradeNo:         tradeNo,
		PaymentMethod:   model.PaymentMethodAlipayDirect,
		PaymentProvider: model.PaymentProviderAlipay,
		CreateTime:      time.Now().Unix(),
		Status:          common.TopUpStatusPending,
	}
	if err := topUp.Insert(); err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("支付宝 创建充值订单失败 user_id=%d trade_no=%s error=%q", id, tradeNo, err.Error()))
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "创建订单失败"})
		return
	}

	resp, err := service.CreateAlipayOrder(c.Request.Context(), service.AlipayOrderRequest{
		TradeType:  tradeType,
		Subject:    fmt.Sprintf("Recharge %d credits", req.Amount),
		OutTradeNo: tradeNo,
		AmountYuan: service.MoneyToCNYString(payMoney),
		NotifyURL:  officialNotifyURL(setting.AlipayNotifyURL, "/api/alipay/notify"),
		ReturnURL:  officialReturnURL(setting.AlipayReturnURL, "/api/alipay/return"),
	})
	if err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("支付宝 拉起支付失败 user_id=%d trade_no=%s trade_type=%s error=%q", id, tradeNo, tradeType, err.Error()))
		topUp.Status = common.TopUpStatusFailed
		_ = topUp.Update()
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "拉起支付失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "success", "data": resp})
}

func WechatPayNotify(c *gin.Context) {
	if !isWechatPayWebhookEnabled() {
		logger.LogWarn(c.Request.Context(), fmt.Sprintf("微信支付 webhook 被拒绝 reason=webhook_disabled path=%q client_ip=%s", c.Request.RequestURI, c.ClientIP()))
		c.JSON(http.StatusForbidden, gin.H{"code": "FAIL", "message": "webhook disabled"})
		return
	}
	notification, err := service.ParseWechatPayNotification(c.Request.Context(), c.Request)
	if err != nil {
		logger.LogWarn(c.Request.Context(), fmt.Sprintf("微信支付 webhook 验签或解析失败 client_ip=%s error=%q", c.ClientIP(), err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"code": "FAIL", "message": "invalid notification"})
		return
	}
	if notification.TradeState != "SUCCESS" {
		logger.LogInfo(c.Request.Context(), fmt.Sprintf("微信支付 webhook 忽略事件 trade_no=%s trade_state=%s", notification.TradeNo, notification.TradeState))
		c.JSON(http.StatusOK, gin.H{"code": "SUCCESS", "message": "成功"})
		return
	}

	LockOrder(notification.TradeNo)
	defer UnlockOrder(notification.TradeNo)
	if err := completeOfficialPayment(c, notification.TradeNo, notification.AmountCents, model.PaymentProviderWechatPay, notification.Payload); err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("微信支付 订单完成失败 trade_no=%s amount=%d error=%q", notification.TradeNo, notification.AmountCents, err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"code": "FAIL", "message": "process failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "SUCCESS", "message": "成功"})
}

func AlipayNotify(c *gin.Context) {
	if !isAlipayWebhookEnabled() {
		logger.LogWarn(c.Request.Context(), fmt.Sprintf("支付宝 webhook 被拒绝 reason=webhook_disabled path=%q client_ip=%s", c.Request.RequestURI, c.ClientIP()))
		_, _ = c.Writer.Write([]byte("fail"))
		return
	}
	if err := c.Request.ParseForm(); err != nil {
		_, _ = c.Writer.Write([]byte("fail"))
		return
	}
	notification, err := service.ParseAlipayNotification(c.Request.Context(), c.Request.Form)
	if err != nil {
		logger.LogWarn(c.Request.Context(), fmt.Sprintf("支付宝 webhook 验签失败 client_ip=%s error=%q", c.ClientIP(), err.Error()))
		_, _ = c.Writer.Write([]byte("fail"))
		return
	}
	if !service.IsAlipayTradeSuccess(notification.TradeStatus) {
		logger.LogInfo(c.Request.Context(), fmt.Sprintf("支付宝 webhook 忽略事件 trade_no=%s trade_status=%s", notification.TradeNo, notification.TradeStatus))
		_, _ = c.Writer.Write([]byte("success"))
		return
	}
	amountCents, err := parseYuanToCents(notification.AmountYuan)
	if err != nil {
		logger.LogWarn(c.Request.Context(), fmt.Sprintf("支付宝 webhook 金额解析失败 trade_no=%s total_amount=%s error=%q", notification.TradeNo, notification.AmountYuan, err.Error()))
		_, _ = c.Writer.Write([]byte("fail"))
		return
	}

	LockOrder(notification.TradeNo)
	defer UnlockOrder(notification.TradeNo)
	if err := completeOfficialPayment(c, notification.TradeNo, amountCents, model.PaymentProviderAlipay, notification.Payload); err != nil {
		logger.LogError(c.Request.Context(), fmt.Sprintf("支付宝 订单完成失败 trade_no=%s amount=%d error=%q", notification.TradeNo, amountCents, err.Error()))
		_, _ = c.Writer.Write([]byte("fail"))
		return
	}
	_, _ = c.Writer.Write([]byte("success"))
}

func AlipayReturn(c *gin.Context) {
	if err := service.VerifyAlipayReturn(c.Request.Context(), c.Request.URL.Query()); err != nil {
		c.Redirect(http.StatusFound, paymentReturnPath("/wallet?pay=fail&show_history=true"))
		return
	}
	c.Redirect(http.StatusFound, paymentReturnPath("/wallet?pay=pending&show_history=true"))
}

func SubscriptionRequestWechatPay(c *gin.Context) {
	if !requirePaymentCompliance(c) {
		return
	}
	if !isWechatPayTopUpEnabled() {
		common.ApiErrorMsg(c, "微信支付未启用")
		return
	}
	var req OfficialSubscriptionPayRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.PlanId <= 0 {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	tradeType := service.NormalizeWechatPayTradeType(req.TradeType)
	if !isWechatPayTradeTypeEnabled(tradeType) {
		common.ApiErrorMsg(c, "不支持的微信支付方式")
		return
	}
	plan, user, payMoneyCNY, err := prepareOfficialSubscription(c, req.PlanId)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	tradeNo := fmt.Sprintf("WXPAY-SUB-%d-%d-%s", user.Id, time.Now().UnixMilli(), randstr.String(6))
	order := &model.SubscriptionOrder{
		UserId:          user.Id,
		PlanId:          plan.Id,
		Money:           plan.PriceAmount,
		TradeNo:         tradeNo,
		PaymentMethod:   model.PaymentMethodWechatPay,
		PaymentProvider: model.PaymentProviderWechatPay,
		CreateTime:      time.Now().Unix(),
		Status:          common.TopUpStatusPending,
	}
	if err := order.Insert(); err != nil {
		common.ApiErrorMsg(c, "创建订单失败")
		return
	}
	resp, err := service.CreateWechatPayOrder(c.Request.Context(), service.WechatPayOrderRequest{
		TradeType:   tradeType,
		Description: "Subscription: " + plan.Title,
		OutTradeNo:  tradeNo,
		AmountCents: service.MoneyToCNYCents(payMoneyCNY),
		NotifyURL:   officialNotifyURL(setting.WechatPayNotifyURL, "/api/wechat-pay/notify"),
		ClientIP:    c.ClientIP(),
		OpenID:      user.WeChatId,
		RedirectURL: paymentReturnPath("/wallet?show_history=true"),
	})
	if err != nil {
		order.Status = common.TopUpStatusFailed
		_ = order.Update()
		common.ApiErrorMsg(c, "拉起支付失败: "+err.Error())
		return
	}
	common.ApiSuccess(c, resp)
}

func SubscriptionRequestAlipay(c *gin.Context) {
	if !requirePaymentCompliance(c) {
		return
	}
	if !isAlipayTopUpEnabled() {
		common.ApiErrorMsg(c, "支付宝支付未启用")
		return
	}
	var req OfficialSubscriptionPayRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.PlanId <= 0 {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	tradeType := service.NormalizeAlipayTradeType(req.TradeType)
	if !isAlipayTradeTypeEnabled(tradeType) {
		common.ApiErrorMsg(c, "不支持的支付宝支付方式")
		return
	}
	plan, user, payMoneyCNY, err := prepareOfficialSubscription(c, req.PlanId)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	tradeNo := fmt.Sprintf("ALIPAY-SUB-%d-%d-%s", user.Id, time.Now().UnixMilli(), randstr.String(6))
	order := &model.SubscriptionOrder{
		UserId:          user.Id,
		PlanId:          plan.Id,
		Money:           plan.PriceAmount,
		TradeNo:         tradeNo,
		PaymentMethod:   model.PaymentMethodAlipayDirect,
		PaymentProvider: model.PaymentProviderAlipay,
		CreateTime:      time.Now().Unix(),
		Status:          common.TopUpStatusPending,
	}
	if err := order.Insert(); err != nil {
		common.ApiErrorMsg(c, "创建订单失败")
		return
	}
	resp, err := service.CreateAlipayOrder(c.Request.Context(), service.AlipayOrderRequest{
		TradeType:  tradeType,
		Subject:    "Subscription: " + plan.Title,
		OutTradeNo: tradeNo,
		AmountYuan: service.MoneyToCNYString(payMoneyCNY),
		NotifyURL:  officialNotifyURL(setting.AlipayNotifyURL, "/api/alipay/notify"),
		ReturnURL:  officialReturnURL(setting.AlipayReturnURL, "/api/alipay/return"),
	})
	if err != nil {
		order.Status = common.TopUpStatusFailed
		_ = order.Update()
		common.ApiErrorMsg(c, "拉起支付失败")
		return
	}
	common.ApiSuccess(c, resp)
}

func requestOfficialAmount(c *gin.Context, unitPrice float64) {
	var req AmountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "参数错误"})
		return
	}
	id := c.GetInt("id")
	_, payMoney, err := validateOfficialTopUpAmount(id, req.Amount, unitPrice)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "success", "data": service.MoneyToCNYString(payMoney)})
}

func validateOfficialTopUpAmount(userId int, reqAmount int64, unitPrice float64) (int64, float64, error) {
	if reqAmount < getMinTopup() {
		return 0, 0, fmt.Errorf("充值数量不能小于 %d", getMinTopup())
	}
	group, err := model.GetUserGroup(userId, true)
	if err != nil {
		return 0, 0, errors.New("获取用户分组失败")
	}
	payMoney := getOfficialPayMoney(reqAmount, group, unitPrice)
	if payMoney < 0.01 {
		return 0, 0, errors.New("充值金额过低")
	}
	amount := reqAmount
	if operation_setting.GetQuotaDisplayType() == operation_setting.QuotaDisplayTypeTokens {
		amount = decimal.NewFromInt(reqAmount).Div(decimal.NewFromFloat(common.QuotaPerUnit)).IntPart()
	}
	return amount, payMoney, nil
}

func officialTopUpUnitPrice(unitPrice float64) float64 {
	if unitPrice <= 0 {
		return operation_setting.Price
	}
	return unitPrice
}

func getOfficialPayMoney(amount int64, group string, unitPrice float64) float64 {
	return getPayMoneyWithUnitPrice(amount, group, unitPrice)
}

func completeOfficialPayment(c *gin.Context, tradeNo string, paidCents int64, expectedProvider string, providerPayload any) error {
	if tradeNo == "" {
		return errors.New("订单号为空")
	}
	if topUp := model.GetTopUpByTradeNo(tradeNo); topUp != nil {
		if topUp.PaymentProvider != expectedProvider {
			return model.ErrPaymentMethodMismatch
		}
		if service.MoneyToCNYCents(topUp.Money) != paidCents {
			return errors.New("支付金额不匹配")
		}
		if expectedProvider == model.PaymentProviderWechatPay {
			return model.RechargeWechatPay(tradeNo, c.ClientIP())
		}
		return model.RechargeAlipay(tradeNo, c.ClientIP())
	}

	order := model.GetSubscriptionOrderByTradeNo(tradeNo)
	if order == nil {
		return errors.New("订单不存在")
	}
	if order.PaymentProvider != expectedProvider {
		return model.ErrPaymentMethodMismatch
	}
	expectedCNY, err := expectedSubscriptionPayMoneyCNY(order)
	if err != nil {
		return err
	}
	if service.MoneyToCNYCents(expectedCNY) != paidCents {
		return errors.New("支付金额不匹配")
	}
	payload := ""
	if providerPayload != nil {
		payload = common.GetJsonString(providerPayload)
	}
	return model.CompleteSubscriptionOrder(tradeNo, payload, expectedProvider, order.PaymentMethod)
}

func prepareOfficialSubscription(c *gin.Context, planId int) (*model.SubscriptionPlan, *model.User, float64, error) {
	plan, err := model.GetSubscriptionPlanById(planId)
	if err != nil {
		return nil, nil, 0, err
	}
	if !plan.Enabled {
		return nil, nil, 0, errors.New("套餐未启用")
	}
	payMoneyCNY, err := planPayMoneyCNY(plan)
	if err != nil {
		return nil, nil, 0, err
	}
	if payMoneyCNY < 0.01 {
		return nil, nil, 0, errors.New("套餐金额过低")
	}
	userId := c.GetInt("id")
	user, err := model.GetUserById(userId, false)
	if err != nil || user == nil {
		return nil, nil, 0, errors.New("用户不存在")
	}
	if plan.MaxPurchasePerUser > 0 {
		count, err := model.CountUserSubscriptionsByPlan(userId, plan.Id)
		if err != nil {
			return nil, nil, 0, err
		}
		if count >= int64(plan.MaxPurchasePerUser) {
			return nil, nil, 0, errors.New("已达到该套餐购买上限")
		}
	}
	return plan, user, payMoneyCNY, nil
}

func expectedSubscriptionPayMoneyCNY(order *model.SubscriptionOrder) (float64, error) {
	plan, err := model.GetSubscriptionPlanById(order.PlanId)
	if err != nil {
		return 0, err
	}
	return planPayMoneyCNY(plan)
}

func planPayMoneyCNY(plan *model.SubscriptionPlan) (float64, error) {
	currency := strings.ToUpper(strings.TrimSpace(plan.Currency))
	switch currency {
	case "", "CNY":
		return plan.PriceAmount, nil
	case "USD":
		return plan.PriceAmount * operation_setting.USDExchangeRate, nil
	default:
		return 0, fmt.Errorf("官方微信/支付宝暂不支持该套餐币种: %s", currency)
	}
}

func parseYuanToCents(amount string) (int64, error) {
	d, err := decimal.NewFromString(amount)
	if err != nil {
		return 0, err
	}
	return d.Mul(decimal.NewFromInt(100)).Round(0).IntPart(), nil
}

func officialNotifyURL(override string, path string) string {
	if strings.TrimSpace(override) != "" {
		return strings.TrimSpace(override)
	}
	return service.GetCallbackAddress() + path
}

func officialReturnURL(override string, path string) string {
	if strings.TrimSpace(override) != "" {
		return strings.TrimSpace(override)
	}
	return service.GetCallbackAddress() + path
}

func isWechatPayTradeTypeEnabled(tradeType string) bool {
	switch tradeType {
	case service.WechatPayTradeTypeJSAPI:
		return setting.WechatPayJSAPIEnabled
	case service.WechatPayTradeTypeH5:
		return setting.WechatPayH5Enabled
	case service.WechatPayTradeTypeNative:
		return setting.WechatPayNativeEnabled
	default:
		return false
	}
}

func isAlipayTradeTypeEnabled(tradeType string) bool {
	switch tradeType {
	case service.AlipayTradeTypeWap:
		return setting.AlipayWapEnabled
	case service.AlipayTradeTypePrecreate:
		return setting.AlipayFaceEnabled
	case service.AlipayTradeTypePage:
		return setting.AlipayPageEnabled
	default:
		return false
	}
}
