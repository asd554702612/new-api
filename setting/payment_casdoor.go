package setting

import "strings"

const (
	DefaultCasdoorBaseURL         = "https://login.gepinkeji.com"
	DefaultCasdoorPaymentProduct  = "external-pay-template"
	DefaultCasdoorPaymentProvider = "provider_payment_wechat_gepinkeji"
	DefaultCasdoorPaymentCurrency = "CNY"
)

var (
	CasdoorPaymentEnabled      bool
	CasdoorIdentityEnabled     bool
	CasdoorIdentityApiRequired bool
	CasdoorBaseURL             string = DefaultCasdoorBaseURL
	CasdoorClientID            string
	CasdoorClientSecret        string
	CasdoorApplicationName     string
	CasdoorIdentityCallbackURL string
	CasdoorPaymentProduct      string = DefaultCasdoorPaymentProduct
	CasdoorPaymentProvider     string = DefaultCasdoorPaymentProvider
	CasdoorPaymentCurrency     string = DefaultCasdoorPaymentCurrency
	CasdoorPaymentUnitPrice    float64
	CasdoorPaymentMinTopUp     int = 1
)

func GetCasdoorBaseURL() string {
	if strings.TrimSpace(CasdoorBaseURL) == "" {
		return DefaultCasdoorBaseURL
	}
	return strings.TrimRight(strings.TrimSpace(CasdoorBaseURL), "/")
}

func GetCasdoorIdentityCallbackURL(defaultBaseURL string) string {
	if strings.TrimSpace(CasdoorIdentityCallbackURL) != "" {
		return strings.TrimSpace(CasdoorIdentityCallbackURL)
	}
	return strings.TrimRight(strings.TrimSpace(defaultBaseURL), "/") + "/identity/callback"
}

func GetCasdoorPaymentProduct() string {
	if strings.TrimSpace(CasdoorPaymentProduct) == "" {
		return DefaultCasdoorPaymentProduct
	}
	return strings.TrimSpace(CasdoorPaymentProduct)
}

func GetCasdoorPaymentProvider() string {
	if strings.TrimSpace(CasdoorPaymentProvider) == "" {
		return DefaultCasdoorPaymentProvider
	}
	return strings.TrimSpace(CasdoorPaymentProvider)
}

func GetCasdoorPaymentCurrency() string {
	if strings.TrimSpace(CasdoorPaymentCurrency) == "" {
		return DefaultCasdoorPaymentCurrency
	}
	return strings.ToUpper(strings.TrimSpace(CasdoorPaymentCurrency))
}

func GetCasdoorPaymentMinTopUp() int {
	if CasdoorPaymentMinTopUp <= 0 {
		return 1
	}
	return CasdoorPaymentMinTopUp
}
