import assert from 'node:assert/strict';
import {
  buildCasdoorSettingOptions,
  normalizeCasdoorSettingInputs,
} from './casdoorSettingsOptions.js';

const normalized = normalizeCasdoorSettingInputs({
  CasdoorPaymentEnabled: 'true',
  CasdoorIdentityEnabled: 'true',
  CasdoorIdentityApiRequired: 'false',
  CasdoorBaseURL: 'https://login.gepinkeji.com/',
  CasdoorClientID: 'app-token',
  CasdoorApplicationName: 'app-token-gepinkeji',
  CasdoorIdentityCallbackURL: 'https://token.gepinkeji.com/identity/callback',
});

assert.equal(normalized.CasdoorPaymentEnabled, true);
assert.equal(normalized.CasdoorIdentityEnabled, true);
assert.equal(normalized.CasdoorIdentityApiRequired, false);
assert.equal(normalized.CasdoorBaseURL, 'https://login.gepinkeji.com/');

const options = buildCasdoorSettingOptions({
  ...normalized,
  CasdoorPaymentEnabled: false,
  CasdoorIdentityApiRequired: true,
  CasdoorBaseURL: 'https://login.gepinkeji.com/',
  CasdoorClientSecret: 'new-secret',
});

assert.deepEqual(
  options.filter((item) =>
    [
      'CasdoorPaymentEnabled',
      'CasdoorIdentityEnabled',
      'CasdoorIdentityApiRequired',
      'CasdoorBaseURL',
      'CasdoorClientSecret',
      'CasdoorIdentityCallbackURL',
    ].includes(item.key),
  ),
  [
    { key: 'CasdoorPaymentEnabled', value: 'false' },
    { key: 'CasdoorIdentityEnabled', value: 'true' },
    { key: 'CasdoorIdentityApiRequired', value: 'true' },
    { key: 'CasdoorBaseURL', value: 'https://login.gepinkeji.com' },
    { key: 'CasdoorClientSecret', value: 'new-secret' },
    {
      key: 'CasdoorIdentityCallbackURL',
      value: 'https://token.gepinkeji.com/identity/callback',
    },
  ],
);

assert.equal(
  buildCasdoorSettingOptions({
    ...normalized,
    CasdoorClientSecret: '   ',
  }).some((item) => item.key === 'CasdoorClientSecret'),
  false,
  'blank Casdoor client secret keeps the existing secret',
);

console.log('casdoor settings option tests passed');
