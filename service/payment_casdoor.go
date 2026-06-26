package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting"
)

type CasdoorPaymentCreateRequest struct {
	ExternalOrderID string  `json:"externalOrderId"`
	UserID          string  `json:"userId,omitempty"`
	Owner           string  `json:"owner,omitempty"`
	UserName        string  `json:"userName,omitempty"`
	ProductName     string  `json:"productName"`
	ProviderName    string  `json:"providerName"`
	Amount          float64 `json:"amount"`
	Currency        string  `json:"currency"`
	DisplayName     string  `json:"displayName,omitempty"`
	Detail          string  `json:"detail,omitempty"`
}

type CasdoorPaymentCreateResult struct {
	OrderID         string  `json:"orderId"`
	PaymentID       string  `json:"paymentId"`
	ExternalOrderID string  `json:"externalOrderId"`
	PayURL          string  `json:"payUrl"`
	State           string  `json:"state"`
	Amount          float64 `json:"amount"`
	Currency        string  `json:"currency"`
	ProviderName    string  `json:"providerName"`
}

type CasdoorPaymentWebhookEvent struct {
	Event           string                  `json:"event"`
	Application     string                  `json:"application"`
	ExternalOrderID string                  `json:"externalOrderId"`
	OrderID         string                  `json:"orderId"`
	PaymentID       string                  `json:"paymentId"`
	UserID          string                  `json:"userId"`
	Products        []CasdoorWebhookProduct `json:"products"`
	Amount          float64                 `json:"amount"`
	Currency        string                  `json:"currency"`
	ProviderName    string                  `json:"providerName"`
	PaidTime        string                  `json:"paidTime"`
}

type CasdoorWebhookProduct struct {
	Owner       string  `json:"owner"`
	Name        string  `json:"name"`
	DisplayName string  `json:"displayName"`
	Price       float64 `json:"price"`
	Currency    string  `json:"currency"`
	Quantity    int     `json:"quantity"`
}

type casdoorPaymentAPIResponse struct {
	Status string                     `json:"status"`
	Msg    string                     `json:"msg"`
	Data   CasdoorPaymentCreateResult `json:"data"`
}

func BuildCasdoorPaymentSignature(secret string, timestamp string, nonce string, rawBody []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp))
	mac.Write([]byte("\n"))
	mac.Write([]byte(nonce))
	mac.Write([]byte("\n"))
	mac.Write(rawBody)
	return hex.EncodeToString(mac.Sum(nil))
}

func BuildCasdoorWebhookSignature(secret string, rawBody []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(rawBody)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func VerifyCasdoorWebhookSignature(secret string, rawBody []byte, signatureHeader string) bool {
	expected := BuildCasdoorWebhookSignature(secret, rawBody)
	actual := strings.TrimSpace(signatureHeader)
	return hmac.Equal([]byte(expected), []byte(actual))
}

func ParseCasdoorWebhookEvent(rawBody []byte) (*CasdoorPaymentWebhookEvent, error) {
	var event CasdoorPaymentWebhookEvent
	if err := common.Unmarshal(rawBody, &event); err != nil {
		return nil, err
	}
	return &event, nil
}

func CreateCasdoorPayment(ctx context.Context, req CasdoorPaymentCreateRequest) (*CasdoorPaymentCreateResult, string, error) {
	if strings.TrimSpace(req.ExternalOrderID) == "" {
		return nil, "", errors.New("externalOrderId is empty")
	}
	if req.Amount < 0.01 {
		return nil, "", errors.New("支付金额过低")
	}
	if strings.TrimSpace(req.ProductName) == "" {
		req.ProductName = setting.GetCasdoorPaymentProduct()
	}
	if strings.TrimSpace(req.ProviderName) == "" {
		req.ProviderName = setting.GetCasdoorPaymentProvider()
	}
	if strings.TrimSpace(req.Currency) == "" {
		req.Currency = setting.GetCasdoorPaymentCurrency()
	}

	body, err := common.Marshal(req)
	if err != nil {
		return nil, "", err
	}
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	nonce := common.GetRandomString(24)
	signature := BuildCasdoorPaymentSignature(setting.CasdoorClientSecret, timestamp, nonce, body)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, setting.GetCasdoorBaseURL()+"/api/external/payment/create", bytes.NewReader(body))
	if err != nil {
		return nil, "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Casdoor-App-Id", setting.CasdoorClientID)
	httpReq.Header.Set("X-Casdoor-Timestamp", timestamp)
	httpReq.Header.Set("X-Casdoor-Nonce", nonce)
	httpReq.Header.Set("X-Casdoor-Signature", signature)

	client := GetHttpClient()
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, string(body), err
	}
	defer resp.Body.Close()

	var apiResp casdoorPaymentAPIResponse
	if err := common.DecodeJson(resp.Body, &apiResp); err != nil {
		return nil, string(body), err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices || strings.ToLower(apiResp.Status) != "ok" {
		msg := strings.TrimSpace(apiResp.Msg)
		if msg == "" {
			msg = fmt.Sprintf("casdoor payment create failed with status %d", resp.StatusCode)
		}
		return nil, string(body), errors.New(msg)
	}
	return &apiResp.Data, string(body), nil
}
