/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/

const defaultInputs = {
  CasdoorPaymentEnabled: false,
  CasdoorIdentityEnabled: false,
  CasdoorIdentityApiRequired: false,
  CasdoorBaseURL: 'https://login.gepinkeji.com',
  CasdoorClientID: '',
  CasdoorClientSecret: '',
  CasdoorApplicationName: '',
  CasdoorIdentityCallbackURL: '',
  CasdoorPaymentProduct: 'external-pay-template',
  CasdoorPaymentProvider: 'provider_payment_wechat_gepinkeji',
  CasdoorPaymentCurrency: 'CNY',
  CasdoorPaymentUnitPrice: 0,
  CasdoorPaymentMinTopUp: 1,
};

const toBoolean = (value) => {
  if (typeof value === 'boolean') return value;
  return String(value).toLowerCase() === 'true';
};

const removeTrailingSlash = (url) => String(url || '').replace(/\/+$/, '');

const parseNumber = (value, fallback) => {
  if (value === undefined || value === null || value === '') return fallback;
  const parsed = Number(value);
  return Number.isNaN(parsed) ? fallback : parsed;
};

export const normalizeCasdoorSettingInputs = (options = {}) => ({
  CasdoorPaymentEnabled: toBoolean(
    options.CasdoorPaymentEnabled ?? defaultInputs.CasdoorPaymentEnabled,
  ),
  CasdoorIdentityEnabled: toBoolean(
    options.CasdoorIdentityEnabled ?? defaultInputs.CasdoorIdentityEnabled,
  ),
  CasdoorIdentityApiRequired: toBoolean(
    options.CasdoorIdentityApiRequired ??
      defaultInputs.CasdoorIdentityApiRequired,
  ),
  CasdoorBaseURL: options.CasdoorBaseURL || defaultInputs.CasdoorBaseURL,
  CasdoorClientID: options.CasdoorClientID || '',
  CasdoorClientSecret: '',
  CasdoorApplicationName: options.CasdoorApplicationName || '',
  CasdoorIdentityCallbackURL: options.CasdoorIdentityCallbackURL || '',
  CasdoorPaymentProduct:
    options.CasdoorPaymentProduct || defaultInputs.CasdoorPaymentProduct,
  CasdoorPaymentProvider:
    options.CasdoorPaymentProvider || defaultInputs.CasdoorPaymentProvider,
  CasdoorPaymentCurrency:
    options.CasdoorPaymentCurrency || defaultInputs.CasdoorPaymentCurrency,
  CasdoorPaymentUnitPrice: parseNumber(options.CasdoorPaymentUnitPrice, 0),
  CasdoorPaymentMinTopUp: parseNumber(options.CasdoorPaymentMinTopUp, 1),
});

export const buildCasdoorSettingOptions = (values = {}) => {
  const normalized = normalizeCasdoorSettingInputs(values);
  const options = [
    {
      key: 'CasdoorPaymentEnabled',
      value: normalized.CasdoorPaymentEnabled ? 'true' : 'false',
    },
    {
      key: 'CasdoorIdentityEnabled',
      value: normalized.CasdoorIdentityEnabled ? 'true' : 'false',
    },
    {
      key: 'CasdoorIdentityApiRequired',
      value: normalized.CasdoorIdentityApiRequired ? 'true' : 'false',
    },
    {
      key: 'CasdoorBaseURL',
      value: removeTrailingSlash(normalized.CasdoorBaseURL),
    },
    { key: 'CasdoorClientID', value: normalized.CasdoorClientID },
    {
      key: 'CasdoorApplicationName',
      value: normalized.CasdoorApplicationName,
    },
    {
      key: 'CasdoorPaymentProduct',
      value:
        normalized.CasdoorPaymentProduct || defaultInputs.CasdoorPaymentProduct,
    },
    {
      key: 'CasdoorPaymentProvider',
      value:
        normalized.CasdoorPaymentProvider ||
        defaultInputs.CasdoorPaymentProvider,
    },
    {
      key: 'CasdoorPaymentCurrency',
      value: (normalized.CasdoorPaymentCurrency || 'CNY').toUpperCase(),
    },
    {
      key: 'CasdoorPaymentUnitPrice',
      value: normalized.CasdoorPaymentUnitPrice.toString(),
    },
    {
      key: 'CasdoorPaymentMinTopUp',
      value: normalized.CasdoorPaymentMinTopUp.toString(),
    },
    {
      key: 'CasdoorIdentityCallbackURL',
      value: normalized.CasdoorIdentityCallbackURL,
    },
  ];

  if ((values.CasdoorClientSecret || '').trim()) {
    options.splice(5, 0, {
      key: 'CasdoorClientSecret',
      value: values.CasdoorClientSecret,
    });
  }

  return options;
};

export { defaultInputs as defaultCasdoorSettingInputs };
