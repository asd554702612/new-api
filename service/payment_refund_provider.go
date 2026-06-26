package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	alipay "github.com/smartwalle/alipay/v3"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/services/refunddomestic"
)

type paymentProviderRefundRequest struct {
	Provider          string
	TradeNo           string
	ProviderRefundNo  string
	Reason            string
	RefundAmountCents int64
	TotalAmountCents  int64
}

type paymentProviderRefundResult struct {
	ProviderRefundID     string
	ProviderRefundStatus string
	ProviderResponse     string
}

var paymentProviderRefunder = refundPaymentProvider

func setPaymentProviderRefunderForTest(fn func(paymentProviderRefundRequest) (*paymentProviderRefundResult, error)) func() {
	previous := paymentProviderRefunder
	paymentProviderRefunder = fn
	return func() {
		paymentProviderRefunder = previous
	}
}

func refundPaymentProvider(req paymentProviderRefundRequest) (*paymentProviderRefundResult, error) {
	switch req.Provider {
	case model.PaymentProviderWechatPay:
		return refundWechatPayProvider(context.Background(), req)
	case model.PaymentProviderAlipay:
		return refundAlipayProvider(context.Background(), req)
	default:
		return nil, fmt.Errorf("支付渠道 %s 不支持自动退款，请手动处理", req.Provider)
	}
}

func refundWechatPayProvider(ctx context.Context, req paymentProviderRefundRequest) (*paymentProviderRefundResult, error) {
	if req.RefundAmountCents <= 0 || req.TotalAmountCents <= 0 {
		return nil, errors.New("微信退款金额无效")
	}
	client, err := newWechatPayClient(ctx)
	if err != nil {
		return nil, err
	}
	svc := refunddomestic.RefundsApiService{Client: client}
	resp, _, err := svc.Create(ctx, refunddomestic.CreateRequest{
		OutTradeNo:  core.String(req.TradeNo),
		OutRefundNo: core.String(req.ProviderRefundNo),
		Reason:      core.String(req.Reason),
		Amount: &refunddomestic.AmountReq{
			Currency: core.String("CNY"),
			Refund:   core.Int64(req.RefundAmountCents),
			Total:    core.Int64(req.TotalAmountCents),
		},
	})
	if err != nil {
		return nil, err
	}
	status := ""
	if resp != nil && resp.Status != nil {
		status = string(*resp.Status)
	}
	refundID := ""
	if resp != nil && resp.RefundId != nil {
		refundID = *resp.RefundId
	}
	return &paymentProviderRefundResult{
		ProviderRefundID:     refundID,
		ProviderRefundStatus: status,
		ProviderResponse:     common.GetJsonString(resp),
	}, nil
}

func refundAlipayProvider(ctx context.Context, req paymentProviderRefundRequest) (*paymentProviderRefundResult, error) {
	if req.RefundAmountCents <= 0 {
		return nil, errors.New("支付宝退款金额无效")
	}
	client, err := newAlipayClient()
	if err != nil {
		return nil, err
	}
	resp, err := client.TradeRefund(ctx, alipay.TradeRefund{
		OutTradeNo:   req.TradeNo,
		RefundAmount: centsToYuanString(req.RefundAmountCents),
		RefundReason: req.Reason,
		OutRequestNo: req.ProviderRefundNo,
	})
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, errors.New("支付宝退款响应为空")
	}
	if resp.IsFailure() {
		msg := strings.TrimSpace(resp.Msg + " " + resp.SubMsg)
		if msg == "" {
			msg = string(resp.Code)
		}
		return nil, errors.New(msg)
	}
	status := "SUCCESS"
	if strings.EqualFold(resp.FundChange, "N") {
		status = "NO_FUND_CHANGE"
	}
	return &paymentProviderRefundResult{
		ProviderRefundID:     resp.TradeNo,
		ProviderRefundStatus: status,
		ProviderResponse:     common.GetJsonString(resp),
	}, nil
}

func centsToYuanString(cents int64) string {
	return fmt.Sprintf("%.2f", float64(cents)/100)
}
