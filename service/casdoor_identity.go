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
	"net/url"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
)

type CasdoorIdentityClient struct {
	BaseURL      string
	ClientID     string
	ClientSecret string
	HTTPClient   *http.Client
}

type CasdoorIdentity struct {
	UserID      string `json:"userId"`
	Owner       string `json:"owner"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	IsVerified  bool   `json:"isVerified"`
	AgeChecked  bool   `json:"ageChecked"`
	IsOver16    bool   `json:"isOver16"`
}

type casdoorIdentitySyncRequest struct {
	UserID string `json:"userId"`
}

type casdoorIdentitySyncResponse struct {
	Status string           `json:"status"`
	Msg    string           `json:"msg"`
	Data   CasdoorIdentity  `json:"data"`
	Data2  *CasdoorIdentity `json:"data2"`
}

func NewCasdoorIdentityClient(baseURL string, clientID string, clientSecret string) *CasdoorIdentityClient {
	return &CasdoorIdentityClient{
		BaseURL:      strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		ClientID:     strings.TrimSpace(clientID),
		ClientSecret: strings.TrimSpace(clientSecret),
		HTTPClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (client *CasdoorIdentityClient) SyncUser(ctx context.Context, userID string) (*CasdoorIdentity, error) {
	if client == nil {
		return nil, errors.New("missing Casdoor identity client")
	}
	if client.BaseURL == "" || client.ClientID == "" || client.ClientSecret == "" {
		return nil, errors.New("missing Casdoor identity config")
	}
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, errors.New("missing Casdoor user id")
	}

	rawBodyBytes, err := common.Marshal(casdoorIdentitySyncRequest{UserID: userID})
	if err != nil {
		return nil, err
	}
	rawBody := string(rawBodyBytes)
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	nonce := common.GetUUID()
	signature := signCasdoorIdentityBody(client.ClientSecret, timestamp, nonce, rawBody)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, client.BaseURL+"/api/external/user/sync", bytes.NewReader(rawBodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Casdoor-App-Id", client.ClientID)
	req.Header.Set("X-Casdoor-Timestamp", timestamp)
	req.Header.Set("X-Casdoor-Nonce", nonce)
	req.Header.Set("X-Casdoor-Signature", signature)

	httpClient := client.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 5 * time.Second}
	}
	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var syncResponse casdoorIdentitySyncResponse
	if err := common.DecodeJson(res.Body, &syncResponse); err != nil {
		return nil, err
	}
	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices || syncResponse.Status != "ok" {
		if strings.TrimSpace(syncResponse.Msg) != "" {
			return nil, errors.New(syncResponse.Msg)
		}
		return nil, fmt.Errorf("sync Casdoor user identity failed: status=%d", res.StatusCode)
	}
	if syncResponse.Data.UserID == "" {
		syncResponse.Data.UserID = userID
	}
	return &syncResponse.Data, nil
}

func (client *CasdoorIdentityClient) BuildVerificationURL(userID string, redirectURI string, state string) (string, error) {
	if client == nil {
		return "", errors.New("missing Casdoor identity client")
	}
	if client.BaseURL == "" || client.ClientID == "" || client.ClientSecret == "" {
		return "", errors.New("missing Casdoor identity config")
	}
	userID = strings.TrimSpace(userID)
	redirectURI = strings.TrimSpace(redirectURI)
	state = strings.TrimSpace(state)
	if userID == "" || redirectURI == "" || state == "" {
		return "", errors.New("missing Casdoor identity verification params")
	}

	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	nonce := common.GetUUID()
	signature := signCasdoorIdentityVerificationURL(client.ClientSecret, timestamp, nonce, client.ClientID, userID, redirectURI, state)

	verificationURL, err := url.Parse(client.BaseURL + "/identity-verification/submit")
	if err != nil {
		return "", err
	}
	query := verificationURL.Query()
	query.Set("clientId", client.ClientID)
	query.Set("userId", userID)
	query.Set("redirectUri", redirectURI)
	query.Set("state", state)
	query.Set("timestamp", timestamp)
	query.Set("nonce", nonce)
	query.Set("signature", signature)
	verificationURL.RawQuery = query.Encode()
	return verificationURL.String(), nil
}

func CanEnterCasdoorIdentityBusiness(identity *CasdoorIdentity) bool {
	return identity != nil && identity.IsVerified && identity.AgeChecked && identity.IsOver16
}

func signCasdoorIdentityBody(secret string, timestamp string, nonce string, rawBody string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(timestamp + "\n" + nonce + "\n" + rawBody))
	return hex.EncodeToString(mac.Sum(nil))
}

func signCasdoorIdentityVerificationURL(secret string, timestamp string, nonce string, clientID string, userID string, redirectURI string, state string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(timestamp + "\n" + nonce + "\n" + clientID + "\n" + userID + "\n" + redirectURI + "\n" + state))
	return hex.EncodeToString(mac.Sum(nil))
}
