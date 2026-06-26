package common

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	PhoneVerificationPurposeLogin          = "sms_login"
	PhoneVerificationPurposeRegister       = "sms_register"
	PhoneVerificationPurposeBind           = "sms_bind"
	PhoneVerificationPurposeChangePassword = "sms_change_password"
	PhoneVerificationPurposePasswordReset  = "sms_password_reset"

	defaultIHuyiSMSBaseURL    = "https://api.ihuyi.com/sms/Submit.json"
	defaultIHuyiSMSTemplateID = "309190"
)

var (
	ErrSMSNotConfigured     = errors.New("sms service not configured")
	ErrSMSProviderDisabled  = errors.New("sms provider is disabled")
	ErrSMSCredentialInvalid = errors.New("sms credential is invalid")
	ErrSMSPhoneInvalid      = errors.New("phone number is invalid")
	ErrSMSQuotaExceeded     = errors.New("sms quota exceeded")
	ErrSMSSendFailed        = errors.New("failed to send sms verification code")
	ErrSMSCodeTooFrequent   = errors.New("sms verification code sent too frequently")
	ErrSMSCodeInvalid       = errors.New("sms verification code is incorrect or has expired")
)

type SMSIHuyiSettings struct {
	Enabled    bool
	Account    string
	Password   string
	TemplateID string
}

type SMSProvider interface {
	SendVerificationCode(ctx context.Context, phoneNumber string, code string) error
}

var smsProviderFactory = func(settings SMSIHuyiSettings) SMSProvider {
	return NewIHuyiSMSProvider(settings, nil)
}

func ResolveSMSIHuyiSettings() SMSIHuyiSettings {
	settings := SMSIHuyiSettings{
		Enabled:    optionBoolWithDefault("SMSIHuyiEnabled", true),
		Account:    optionString("SMSIHuyiAPIID"),
		Password:   optionString("SMSIHuyiAPIKey"),
		TemplateID: optionString("SMSIHuyiTemplateID"),
	}

	if os.Getenv("SMS_IHUYI_ENABLED") != "" {
		settings.Enabled = strings.EqualFold(strings.TrimSpace(os.Getenv("SMS_IHUYI_ENABLED")), "true")
	}
	if v := strings.TrimSpace(os.Getenv("SMS_IHUYI_API_ID")); v != "" {
		settings.Account = v
	}
	if v := strings.TrimSpace(os.Getenv("SMS_IHUYI_API_KEY")); v != "" {
		settings.Password = v
	}
	if v := strings.TrimSpace(os.Getenv("SMS_IHUYI_TEMPLATE_ID")); v != "" {
		settings.TemplateID = v
	}
	if settings.TemplateID == "" {
		settings.TemplateID = defaultIHuyiSMSTemplateID
	}
	return settings
}

func optionString(key string) string {
	OptionMapRWMutex.RLock()
	defer OptionMapRWMutex.RUnlock()
	if OptionMap == nil {
		return ""
	}
	return strings.TrimSpace(OptionMap[key])
}

func optionBool(key string) bool {
	return strings.EqualFold(optionString(key), "true")
}

func optionBoolWithDefault(key string, defaultValue bool) bool {
	value := optionString(key)
	if value == "" {
		return defaultValue
	}
	return strings.EqualFold(value, "true")
}

func IsPhoneVerificationEnabled() bool {
	if v := strings.TrimSpace(os.Getenv("PHONE_VERIFICATION_ENABLED")); v != "" {
		return strings.EqualFold(v, "true")
	}
	if v := optionString("PhoneVerificationEnabled"); v != "" {
		return strings.EqualFold(v, "true")
	}
	return PhoneVerificationEnabled
}

func GenerateNumericVerificationCode(length int) (string, error) {
	if length <= 0 {
		length = 6
	}
	const digits = "0123456789"
	code := make([]byte, length)
	for i := range code {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		code[i] = digits[n.Int64()]
	}
	return string(code), nil
}

func SendPhoneVerificationCode(ctx context.Context, phoneNumber string, purpose string) error {
	phoneNumber = NormalizePhoneNumber(phoneNumber, "86")
	if phoneNumber == "" {
		return ErrSMSPhoneInvalid
	}
	settings := ResolveSMSIHuyiSettings()
	if !settings.Enabled {
		return ErrSMSProviderDisabled
	}
	provider := smsProviderFactory(settings)
	if provider == nil {
		return ErrSMSNotConfigured
	}

	code, err := GenerateNumericVerificationCode(6)
	if err != nil {
		return err
	}
	RegisterVerificationCodeWithKey(phoneNumber, code, purpose)
	if err := provider.SendVerificationCode(ctx, phoneNumber, code); err != nil {
		DeleteKey(phoneNumber, purpose)
		return err
	}
	return nil
}

func VerifyPhoneVerificationCode(phoneNumber string, code string, purpose string) bool {
	phoneNumber = NormalizePhoneNumber(phoneNumber, "86")
	code = strings.TrimSpace(code)
	if phoneNumber == "" || code == "" {
		return false
	}
	return verifyPhoneVerificationCodeByKey(phoneNumber, code, purpose)
}

func verifyPhoneVerificationCodeByKey(phoneNumber string, code string, purpose string) bool {
	key := verificationCacheKey(phoneNumber, purpose)
	if RedisEnabled && RDB != nil {
		value, err := RedisGet(key)
		if err == nil {
			if subtle.ConstantTimeCompare([]byte(value), []byte(code)) != 1 {
				return false
			}
			_ = RedisDel(key)
			deleteVerificationCodeFromMemory(key)
			return true
		}
	}

	verificationMutex.Lock()
	defer verificationMutex.Unlock()
	value, okay := verificationMap[key]
	now := time.Now()
	if !okay || int(now.Sub(value.time).Seconds()) >= VerificationValidMinutes*60 {
		return false
	}
	if subtle.ConstantTimeCompare([]byte(value.code), []byte(code)) != 1 {
		return false
	}
	delete(verificationMap, key)
	return true
}

type IHuyiSMSProvider struct {
	enabled    bool
	baseURL    string
	account    string
	password   string
	templateID string
	client     *http.Client
}

type iHuyiSMSResponse struct {
	Code  int    `json:"code"`
	Msg   string `json:"msg"`
	SMSID string `json:"smsid"`
}

func NewIHuyiSMSProvider(settings SMSIHuyiSettings, client *http.Client) *IHuyiSMSProvider {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	templateID := strings.TrimSpace(settings.TemplateID)
	if templateID == "" {
		templateID = defaultIHuyiSMSTemplateID
	}
	return &IHuyiSMSProvider{
		enabled:    settings.Enabled,
		baseURL:    defaultIHuyiSMSBaseURL,
		account:    strings.TrimSpace(settings.Account),
		password:   strings.TrimSpace(settings.Password),
		templateID: templateID,
		client:     client,
	}
}

func (p *IHuyiSMSProvider) SendVerificationCode(ctx context.Context, phoneNumber string, code string) error {
	if p == nil {
		return ErrSMSNotConfigured
	}
	if !p.enabled {
		return ErrSMSProviderDisabled
	}
	if p.account == "" || p.password == "" {
		return ErrSMSNotConfigured
	}

	form := url.Values{}
	form.Set("account", p.account)
	form.Set("password", p.password)
	form.Set("mobile", normalizeIHuyiMobile(phoneNumber))
	form.Set("templateid", p.templateID)
	form.Set("content", code)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSMSSendFailed, err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSMSSendFailed, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSMSSendFailed, err)
	}

	var payload iHuyiSMSResponse
	if err := Unmarshal(body, &payload); err != nil {
		return fmt.Errorf("%w: %v", ErrSMSSendFailed, err)
	}
	if payload.Code == 2 {
		return nil
	}
	return mapIHuyiSMSResponseError(payload)
}

func normalizeIHuyiMobile(phoneNumber string) string {
	normalized := NormalizePhoneNumber(phoneNumber, "86")
	normalized = strings.TrimPrefix(normalized, "+")
	if strings.HasPrefix(normalized, "86") && len(normalized) > 11 {
		return normalized[2:]
	}
	return normalized
}

func mapIHuyiSMSResponseError(payload iHuyiSMSResponse) error {
	switch payload.Code {
	case 4010, 4011, 4012, 4013:
		return fmt.Errorf("%w: provider=ihuyi code=%s", ErrSMSCredentialInvalid, strconv.Itoa(payload.Code))
	case 4085, 4086:
		return fmt.Errorf("%w: provider=ihuyi code=%s", ErrSMSQuotaExceeded, strconv.Itoa(payload.Code))
	case 4050:
		return fmt.Errorf("%w: provider=ihuyi code=%s", ErrSMSPhoneInvalid, strconv.Itoa(payload.Code))
	default:
		return fmt.Errorf("%w: provider=ihuyi code=%s", ErrSMSSendFailed, strconv.Itoa(payload.Code))
	}
}
