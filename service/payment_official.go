package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strings"

	"github.com/QuantumNous/new-api/setting"
	alipay "github.com/smartwalle/alipay/v3"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/auth/verifiers"
	"github.com/wechatpay-apiv3/wechatpay-go/core/downloader"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/h5"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/jsapi"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"
)

const (
	WechatPayTradeTypeNative = "native"
	WechatPayTradeTypeH5     = "h5"
	WechatPayTradeTypeJSAPI  = "jsapi"

	AlipayTradeTypePage      = "page"
	AlipayTradeTypeWap       = "wap"
	AlipayTradeTypePrecreate = "precreate"
)

type WechatPayOrderRequest struct {
	TradeType   string
	Description string
	OutTradeNo  string
	AmountCents int64
	NotifyURL   string
	ClientIP    string
	OpenID      string
	RedirectURL string
}

type WechatPayOrderResponse struct {
	TradeType   string         `json:"trade_type"`
	CodeURL     string         `json:"code_url,omitempty"`
	CheckoutURL string         `json:"checkout_url,omitempty"`
	JSAPIParams map[string]any `json:"jsapi_params,omitempty"`
}

type WechatPayNotification struct {
	TradeNo     string
	TradeState  string
	AmountCents int64
	Payload     *payments.Transaction
}

type AlipayOrderRequest struct {
	TradeType  string
	Subject    string
	OutTradeNo string
	AmountYuan string
	NotifyURL  string
	ReturnURL  string
}

type AlipayOrderResponse struct {
	TradeType   string `json:"trade_type"`
	CheckoutURL string `json:"checkout_url,omitempty"`
	QRCode      string `json:"qr_code,omitempty"`
}

type AlipayNotification struct {
	TradeNo     string
	TradeStatus string
	AmountYuan  string
	Payload     *alipay.Notification
}

func MoneyToCNYCents(amount float64) int64 {
	return int64(math.Round(amount * 100))
}

func MoneyToCNYString(amount float64) string {
	return fmt.Sprintf("%.2f", amount)
}

func CreateWechatPayOrder(ctx context.Context, req WechatPayOrderRequest) (*WechatPayOrderResponse, error) {
	if req.AmountCents < 1 {
		return nil, errors.New("支付金额过低")
	}
	client, err := newWechatPayClient(ctx)
	if err != nil {
		return nil, err
	}
	tradeType := NormalizeWechatPayTradeType(req.TradeType)
	switch tradeType {
	case WechatPayTradeTypeNative:
		return createWechatPayNativeOrder(ctx, client, req)
	case WechatPayTradeTypeH5:
		return createWechatPayH5Order(ctx, client, req)
	case WechatPayTradeTypeJSAPI:
		if strings.TrimSpace(req.OpenID) == "" {
			return nil, errors.New("当前用户未绑定可用于微信支付 JSAPI 的 openid")
		}
		return createWechatPayJSAPIOrder(ctx, client, req)
	default:
		return nil, errors.New("不支持的微信支付方式")
	}
}

func ParseWechatPayNotification(ctx context.Context, req *http.Request) (*WechatPayNotification, error) {
	if req == nil {
		return nil, errors.New("request is nil")
	}
	if _, err := newWechatPayClient(ctx); err != nil {
		return nil, err
	}
	handler, err := newWechatPayNotifyHandler()
	if err != nil {
		return nil, err
	}
	transaction := new(payments.Transaction)
	if _, err = handler.ParseNotifyRequest(ctx, req, transaction); err != nil {
		return nil, err
	}
	outTradeNo := ""
	if transaction.OutTradeNo != nil {
		outTradeNo = *transaction.OutTradeNo
	}
	tradeState := ""
	if transaction.TradeState != nil {
		tradeState = *transaction.TradeState
	}
	var amount int64
	if transaction.Amount != nil && transaction.Amount.Total != nil {
		amount = *transaction.Amount.Total
	}
	return &WechatPayNotification{
		TradeNo:     outTradeNo,
		TradeState:  tradeState,
		AmountCents: amount,
		Payload:     transaction,
	}, nil
}

func NormalizeWechatPayTradeType(tradeType string) string {
	switch strings.ToLower(strings.TrimSpace(tradeType)) {
	case WechatPayTradeTypeH5:
		return WechatPayTradeTypeH5
	case WechatPayTradeTypeJSAPI:
		return WechatPayTradeTypeJSAPI
	default:
		return WechatPayTradeTypeNative
	}
}

func CreateAlipayOrder(ctx context.Context, req AlipayOrderRequest) (*AlipayOrderResponse, error) {
	if strings.TrimSpace(req.AmountYuan) == "" {
		return nil, errors.New("支付金额不能为空")
	}
	client, err := newAlipayClient()
	if err != nil {
		return nil, err
	}
	tradeType := NormalizeAlipayTradeType(req.TradeType)
	switch tradeType {
	case AlipayTradeTypePage:
		p := alipay.TradePagePay{}
		p.NotifyURL = req.NotifyURL
		p.ReturnURL = req.ReturnURL
		p.Subject = req.Subject
		p.OutTradeNo = req.OutTradeNo
		p.TotalAmount = req.AmountYuan
		p.ProductCode = "FAST_INSTANT_TRADE_PAY"
		payURL, err := client.TradePagePay(p)
		if err != nil {
			return nil, err
		}
		return &AlipayOrderResponse{TradeType: tradeType, CheckoutURL: payURL.String()}, nil
	case AlipayTradeTypeWap:
		p := alipay.TradeWapPay{}
		p.NotifyURL = req.NotifyURL
		p.ReturnURL = req.ReturnURL
		p.Subject = req.Subject
		p.OutTradeNo = req.OutTradeNo
		p.TotalAmount = req.AmountYuan
		p.ProductCode = "QUICK_WAP_WAY"
		payURL, err := client.TradeWapPay(p)
		if err != nil {
			return nil, err
		}
		return &AlipayOrderResponse{TradeType: tradeType, CheckoutURL: payURL.String()}, nil
	case AlipayTradeTypePrecreate:
		p := alipay.TradePreCreate{}
		p.NotifyURL = req.NotifyURL
		p.Subject = req.Subject
		p.OutTradeNo = req.OutTradeNo
		p.TotalAmount = req.AmountYuan
		rsp, err := client.TradePreCreate(ctx, p)
		if err != nil {
			return nil, err
		}
		if rsp.IsFailure() {
			return nil, fmt.Errorf("%s: %s", rsp.Code, rsp.SubMsg)
		}
		return &AlipayOrderResponse{TradeType: tradeType, QRCode: rsp.QRCode}, nil
	default:
		return nil, errors.New("不支持的支付宝支付方式")
	}
}

func ParseAlipayNotification(ctx context.Context, values url.Values) (*AlipayNotification, error) {
	client, err := newAlipayClient()
	if err != nil {
		return nil, err
	}
	notification, err := client.DecodeNotification(ctx, values)
	if err != nil {
		return nil, err
	}
	return &AlipayNotification{
		TradeNo:     notification.OutTradeNo,
		TradeStatus: string(notification.TradeStatus),
		AmountYuan:  notification.TotalAmount,
		Payload:     notification,
	}, nil
}

func VerifyAlipayReturn(ctx context.Context, values url.Values) error {
	client, err := newAlipayClient()
	if err != nil {
		return err
	}
	return client.VerifySign(ctx, values)
}

func NormalizeAlipayTradeType(tradeType string) string {
	switch strings.ToLower(strings.TrimSpace(tradeType)) {
	case AlipayTradeTypeWap:
		return AlipayTradeTypeWap
	case AlipayTradeTypePrecreate:
		return AlipayTradeTypePrecreate
	default:
		return AlipayTradeTypePage
	}
}

func IsAlipayTradeSuccess(status string) bool {
	return status == string(alipay.TradeStatusSuccess) || status == string(alipay.TradeStatusFinished)
}

func newWechatPayClient(ctx context.Context) (*core.Client, error) {
	privateKey, err := utils.LoadPrivateKey(normalizeWechatPayPEM(setting.WechatPayPrivateKey, "PRIVATE KEY"))
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(setting.WechatPayPublicKeyID) != "" || strings.TrimSpace(setting.WechatPayPublicKey) != "" {
		if strings.TrimSpace(setting.WechatPayPublicKeyID) == "" || strings.TrimSpace(setting.WechatPayPublicKey) == "" {
			return nil, errors.New("微信支付公钥 ID 和公钥必须同时配置")
		}
		publicKey, err := utils.LoadPublicKey(normalizeWechatPayPEM(setting.WechatPayPublicKey, "PUBLIC KEY"))
		if err != nil {
			return nil, err
		}
		return core.NewClient(ctx, option.WithWechatPayPublicKeyAuthCipher(
			setting.WechatPayMchID,
			setting.WechatPayMerchantSerialNo,
			privateKey,
			setting.WechatPayPublicKeyID,
			publicKey,
		))
	}
	return core.NewClient(ctx, option.WithWechatPayAutoAuthCipher(
		setting.WechatPayMchID,
		setting.WechatPayMerchantSerialNo,
		privateKey,
		setting.WechatPayAPIv3Key,
	))
}

func newWechatPayNotifyHandler() (*notify.Handler, error) {
	certVisitor := downloader.MgrInstance().GetCertificateVisitor(setting.WechatPayMchID)
	if strings.TrimSpace(setting.WechatPayPublicKeyID) != "" || strings.TrimSpace(setting.WechatPayPublicKey) != "" {
		if strings.TrimSpace(setting.WechatPayPublicKeyID) == "" || strings.TrimSpace(setting.WechatPayPublicKey) == "" {
			return nil, errors.New("微信支付公钥 ID 和公钥必须同时配置")
		}
		publicKey, err := utils.LoadPublicKey(normalizeWechatPayPEM(setting.WechatPayPublicKey, "PUBLIC KEY"))
		if err != nil {
			return nil, err
		}
		return notify.NewRSANotifyHandler(
			setting.WechatPayAPIv3Key,
			verifiers.NewSHA256WithRSACombinedVerifier(certVisitor, setting.WechatPayPublicKeyID, *publicKey),
		)
	}
	return notify.NewRSANotifyHandler(setting.WechatPayAPIv3Key, verifiers.NewSHA256WithRSAVerifier(certVisitor))
}

func normalizeWechatPayPEM(value string, label string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || strings.Contains(trimmed, "-----BEGIN ") {
		return value
	}
	body := strings.Join(strings.Fields(trimmed), "")
	if body == "" {
		return value
	}
	var builder strings.Builder
	builder.WriteString("-----BEGIN ")
	builder.WriteString(label)
	builder.WriteString("-----\n")
	for len(body) > 64 {
		builder.WriteString(body[:64])
		builder.WriteByte('\n')
		body = body[64:]
	}
	builder.WriteString(body)
	builder.WriteString("\n-----END ")
	builder.WriteString(label)
	builder.WriteString("-----\n")
	return builder.String()
}

func createWechatPayNativeOrder(ctx context.Context, client *core.Client, req WechatPayOrderRequest) (*WechatPayOrderResponse, error) {
	svc := native.NativeApiService{Client: client}
	resp, _, err := svc.Prepay(ctx, native.PrepayRequest{
		Appid:       core.String(setting.WechatPayAppID),
		Mchid:       core.String(setting.WechatPayMchID),
		Description: core.String(req.Description),
		OutTradeNo:  core.String(req.OutTradeNo),
		NotifyUrl:   core.String(req.NotifyURL),
		Amount: &native.Amount{
			Currency: core.String("CNY"),
			Total:    core.Int64(req.AmountCents),
		},
	})
	if err != nil {
		return nil, err
	}
	codeURL := ""
	if resp.CodeUrl != nil {
		codeURL = *resp.CodeUrl
	}
	return &WechatPayOrderResponse{TradeType: WechatPayTradeTypeNative, CodeURL: codeURL}, nil
}

func createWechatPayH5Order(ctx context.Context, client *core.Client, req WechatPayOrderRequest) (*WechatPayOrderResponse, error) {
	svc := h5.H5ApiService{Client: client}
	resp, _, err := svc.Prepay(ctx, h5.PrepayRequest{
		Appid:       core.String(setting.WechatPayAppID),
		Mchid:       core.String(setting.WechatPayMchID),
		Description: core.String(req.Description),
		OutTradeNo:  core.String(req.OutTradeNo),
		NotifyUrl:   core.String(req.NotifyURL),
		Amount: &h5.Amount{
			Currency: core.String("CNY"),
			Total:    core.Int64(req.AmountCents),
		},
		SceneInfo: &h5.SceneInfo{
			PayerClientIp: core.String(req.ClientIP),
			H5Info: &h5.H5Info{
				Type: core.String("Wap"),
			},
		},
	})
	if err != nil {
		return nil, err
	}
	checkoutURL := ""
	if resp.H5Url != nil {
		checkoutURL = *resp.H5Url
		if req.RedirectURL != "" {
			checkoutURL += "&redirect_url=" + url.QueryEscape(req.RedirectURL)
		}
	}
	return &WechatPayOrderResponse{TradeType: WechatPayTradeTypeH5, CheckoutURL: checkoutURL}, nil
}

func createWechatPayJSAPIOrder(ctx context.Context, client *core.Client, req WechatPayOrderRequest) (*WechatPayOrderResponse, error) {
	svc := jsapi.JsapiApiService{Client: client}
	resp, _, err := svc.PrepayWithRequestPayment(ctx, jsapi.PrepayRequest{
		Appid:       core.String(setting.WechatPayAppID),
		Mchid:       core.String(setting.WechatPayMchID),
		Description: core.String(req.Description),
		OutTradeNo:  core.String(req.OutTradeNo),
		NotifyUrl:   core.String(req.NotifyURL),
		Amount: &jsapi.Amount{
			Currency: core.String("CNY"),
			Total:    core.Int64(req.AmountCents),
		},
		Payer: &jsapi.Payer{
			Openid: core.String(req.OpenID),
		},
	})
	if err != nil {
		return nil, err
	}
	return &WechatPayOrderResponse{
		TradeType: WechatPayTradeTypeJSAPI,
		JSAPIParams: map[string]any{
			"appId":     valueOfString(resp.Appid),
			"timeStamp": valueOfString(resp.TimeStamp),
			"nonceStr":  valueOfString(resp.NonceStr),
			"package":   valueOfString(resp.Package),
			"signType":  valueOfString(resp.SignType),
			"paySign":   valueOfString(resp.PaySign),
		},
	}, nil
}

func newAlipayClient() (*alipay.Client, error) {
	client, err := alipay.New(setting.AlipayAppID, setting.AlipayPrivateKey, !setting.AlipaySandbox)
	if err != nil {
		return nil, err
	}
	if err = client.LoadAliPayPublicKey(setting.AlipayPublicKey); err != nil {
		return nil, err
	}
	return client, nil
}

func valueOfString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
