import assert from 'node:assert/strict';
import {
  getDefaultOfficialTradeType,
  isCasdoorPayment,
  isOfficialPaymentMethod,
  isSafeOfficialCheckoutUrl,
  normalizeOfficialPaymentResult,
} from './officialPayment.js';

assert.equal(getDefaultOfficialTradeType('wechat_pay', ''), 'native');
assert.equal(
  getDefaultOfficialTradeType('wechat_pay', 'Mozilla/5.0 iPhone Mobile'),
  'h5',
);
assert.equal(
  getDefaultOfficialTradeType('wechat_pay', 'MicroMessenger iPhone Mobile'),
  'jsapi',
);
assert.equal(getDefaultOfficialTradeType('alipay_direct', ''), 'page');
assert.equal(
  getDefaultOfficialTradeType('alipay_direct', 'Mozilla/5.0 Android Mobile'),
  'wap',
);

assert.equal(isOfficialPaymentMethod('wechat_pay'), true);
assert.equal(isOfficialPaymentMethod('alipay_direct'), true);
assert.equal(isOfficialPaymentMethod('casdoor'), true);
assert.equal(isOfficialPaymentMethod('alipay'), false);
assert.equal(isOfficialPaymentMethod('wxpay'), false);
assert.equal(isCasdoorPayment('casdoor'), true);
assert.equal(isCasdoorPayment('wechat_pay'), false);

assert.equal(isSafeOfficialCheckoutUrl('https://example.com/pay'), true);
assert.equal(isSafeOfficialCheckoutUrl('http://example.com/pay'), true);
assert.equal(isSafeOfficialCheckoutUrl('/relative/pay'), false);
assert.equal(isSafeOfficialCheckoutUrl('javascript:alert(1)'), false);
assert.equal(isSafeOfficialCheckoutUrl('data:text/html,ok'), false);

assert.deepEqual(
  normalizeOfficialPaymentResult({
    checkout_url: 'https://example.com/pay',
    trade_type: 'page',
  }),
  { kind: 'redirect', url: 'https://example.com/pay', tradeType: 'page' },
);
assert.deepEqual(
  normalizeOfficialPaymentResult({
    code_url: 'weixin://wxpay/bizpayurl?pr=x',
    trade_type: 'native',
  }),
  {
    kind: 'qr',
    qrValue: 'weixin://wxpay/bizpayurl?pr=x',
    tradeType: 'native',
  },
);
assert.deepEqual(
  normalizeOfficialPaymentResult({
    qr_code: 'https://qr.alipay.com/example',
    trade_type: 'precreate',
  }),
  {
    kind: 'qr',
    qrValue: 'https://qr.alipay.com/example',
    tradeType: 'precreate',
  },
);
assert.deepEqual(
  normalizeOfficialPaymentResult({
    payUrl: 'weixin://wxpay/bizpayurl?pr=casdoor',
  }),
  {
    kind: 'qr',
    qrValue: 'weixin://wxpay/bizpayurl?pr=casdoor',
    tradeType: '',
  },
);
assert.deepEqual(
  normalizeOfficialPaymentResult({
    pay_url: 'weixin://wxpay/bizpayurl?pr=casdoor_snake',
  }),
  {
    kind: 'qr',
    qrValue: 'weixin://wxpay/bizpayurl?pr=casdoor_snake',
    tradeType: '',
  },
);
assert.deepEqual(
  normalizeOfficialPaymentResult({
    jsapi_params: { appId: 'wx', timeStamp: '1' },
    trade_type: 'jsapi',
  }),
  {
    kind: 'jsapi',
    jsapiParams: { appId: 'wx', timeStamp: '1' },
    tradeType: 'jsapi',
  },
);
