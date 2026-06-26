export const OFFICIAL_PAYMENT_WECHAT = 'wechat_pay';
export const OFFICIAL_PAYMENT_ALIPAY = 'alipay_direct';
export const OFFICIAL_PAYMENT_CASDOOR = 'casdoor';

export function isWechatPayPayment(paymentType) {
  return paymentType === OFFICIAL_PAYMENT_WECHAT;
}

export function isAlipayDirectPayment(paymentType) {
  return paymentType === OFFICIAL_PAYMENT_ALIPAY;
}

export function isCasdoorPayment(paymentType) {
  return paymentType === OFFICIAL_PAYMENT_CASDOOR;
}

export function isOfficialPaymentMethod(paymentType) {
  return (
    isWechatPayPayment(paymentType) ||
    isAlipayDirectPayment(paymentType) ||
    isCasdoorPayment(paymentType)
  );
}

export function getDefaultOfficialTradeType(paymentType, userAgent) {
  const ua =
    userAgent !== undefined
      ? userAgent
      : typeof navigator !== 'undefined'
        ? navigator.userAgent
        : '';
  const isMobile = /Mobile|Android|iPhone|iPad|iPod/i.test(ua);
  const isWechatBrowser = /MicroMessenger/i.test(ua);

  if (isWechatPayPayment(paymentType)) {
    if (isWechatBrowser) return 'jsapi';
    return isMobile ? 'h5' : 'native';
  }

  if (isAlipayDirectPayment(paymentType)) {
    return isMobile ? 'wap' : 'page';
  }

  return '';
}

export function isSafeOfficialCheckoutUrl(value) {
  const trimmed = (value || '').trim();
  if (!trimmed) return false;
  try {
    const url = new URL(trimmed);
    return url.protocol === 'http:' || url.protocol === 'https:';
  } catch {
    return false;
  }
}

export function normalizeOfficialPaymentResult(data) {
  const tradeType = data?.trade_type || '';
  const checkoutUrl = data?.checkout_url || '';
  const qrValue = data?.code_url || data?.qr_code || data?.payUrl || data?.pay_url || '';

  if (checkoutUrl) {
    return { kind: 'redirect', url: checkoutUrl, tradeType };
  }

  if (qrValue) {
    return { kind: 'qr', qrValue, tradeType };
  }

  if (data?.jsapi_params) {
    return { kind: 'jsapi', jsapiParams: data.jsapi_params, tradeType };
  }

  return { kind: 'unknown', tradeType };
}
