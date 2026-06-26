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

import assert from 'node:assert/strict';
import test from 'node:test';

import { formatPaymentOrderMoney } from './paymentOrderMoney.js';

function storageWithStatus(status) {
  return {
    getItem(key) {
      if (key === 'status') {
        return JSON.stringify(status);
      }
      return null;
    },
  };
}

test('formats USD subscription order as RMB using status exchange rate', () => {
  const storage = storageWithStatus({ usd_exchange_rate: 7.3 });

  const result = formatPaymentOrderMoney(
    9.99,
    { order_type: 'subscription', pay_currency: 'USD' },
    storage,
  );

  assert.equal(result, '¥72.93');
});

test('keeps CNY subscription order amount unchanged', () => {
  const storage = storageWithStatus({ usd_exchange_rate: 7.3 });

  const result = formatPaymentOrderMoney(
    29,
    { order_type: 'subscription', pay_currency: 'CNY' },
    storage,
  );

  assert.equal(result, '¥29.00');
});

test('keeps top-up order amount unchanged', () => {
  const storage = storageWithStatus({ usd_exchange_rate: 7.3 });

  const result = formatPaymentOrderMoney(
    68,
    { order_type: 'balance', pay_currency: 'USD' },
    storage,
  );

  assert.equal(result, '¥68.00');
});
