package setting

var (
	WechatPayEnabled          bool
	WechatPayUnitPrice        float64
	WechatPayAppID            string
	WechatPayMchID            string
	WechatPayAPIv3Key         string
	WechatPayPrivateKey       string
	WechatPayMerchantSerialNo string
	WechatPayPublicKeyID      string
	WechatPayPublicKey        string
	WechatPayNotifyURL        string
	WechatPayReturnURL        string
	WechatPayJSAPIEnabled     bool
	WechatPayH5Enabled        bool = true
	WechatPayNativeEnabled    bool = true
)

var (
	AlipayEnabled     bool
	AlipayUnitPrice   float64
	AlipayAppID       string
	AlipayPrivateKey  string
	AlipayPublicKey   string
	AlipaySandbox     bool
	AlipayNotifyURL   string
	AlipayReturnURL   string
	AlipayPageEnabled bool = true
	AlipayWapEnabled  bool = true
	AlipayFaceEnabled bool = true
)
